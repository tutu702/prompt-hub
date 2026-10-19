/**
 * Axios HTTP 客户端封装
 *
 * 约定：
 * - 请求统一带 baseURL 与凭证
 * - 响应统一解包 { code, data, message }，由业务侧根据 code 判断
 * - 401 时尝试静默 refresh（白名单跳过 /auth/refresh 与 /auth/login），单飞避免并发刷新
 * - 鉴权白名单接口（登录/注册/刷新）的 HTTP 401 是业务错误，直接透出后端 message；
 *   其他接口的 401 才是 token 过期，走 refresh 流程
 * - 刷新失败则清 token + 跳 /login?redirect=...
 *
 * 真正的业务接口定义在 src/api/modules/*，本文件只提供基础设施。
 */
import axios, {
  AxiosError,
  type AxiosInstance,
  type AxiosRequestConfig,
  type InternalAxiosRequestConfig,
} from 'axios'
import { ElMessage } from 'element-plus'

const baseURL = import.meta.env.VITE_API_BASE_URL || '/api'

const http: AxiosInstance = axios.create({
  baseURL,
  timeout: 15000,
  withCredentials: false,
  headers: {
    'Content-Type': 'application/json',
  },
})

// store getter 通过 main.ts 注入；避免 http.ts 顶层 import useUserStore 形成循环依赖
// 之所以不能用 require()：Vite + ESM 浏览器运行时没有 CommonJS require。
type UserStoreLike = {
  token(): string
  clear(): void
  refresh(): Promise<void>
}
let getUserStore: (() => UserStoreLike | null) | null = null

/** 由 main.ts 在 pinia 安装后调用，把 store getter 注入到 http */
export function setUserStoreGetter(fn: () => UserStoreLike | null) {
  getUserStore = fn
}

/** 仅供诊断使用 */
export function _debugHasToken(): string {
  try {
    const s = getUserStore?.()
    return s?.token()?.slice(0, 10) || '(no-store)'
  } catch {
    return '(error)'
  }
}

// 请求拦截器：从已注入的 Pinia store 读取 access token 注入 Authorization
http.interceptors.request.use(
  (config) => {
    try {
      const store = getUserStore?.()
      const accessToken = store?.token()
      if (accessToken) {
        config.headers = config.headers || {}
        ;(config.headers as Record<string, string>)['Authorization'] =
          `Bearer ${accessToken}`
      }
    } catch {
      // store 还没注入（main.ts 还未执行 setUserStoreGetter）—— 忽略
    }
    return config
  },
  (error) => Promise.reject(error),
)

// 刷新 token 单飞锁：同一时刻只允许一个 refresh 在途
let refreshingPromise: Promise<string> | null = null

// 业务接口 URL 白名单（其 HTTP 401 表示业务错误而非 token 过期）
//
// 包含 /auth/change-password 是为了把"当前密码错误"这类业务错误从 token
// 过期流程里区分出来：旧逻辑会把 401 当 token 过期触发 refresh，重试仍
// 401 后又吐硬编码的"登录已过期"，把真正的"当前密码错误"吞掉。
const AUTH_ENDPOINTS = [
  '/auth/refresh',
  '/auth/login',
  '/auth/register',
  '/auth/change-password',
] as const
function isAuthEndpoint(url: string): boolean {
  return AUTH_ENDPOINTS.some((p) => url.includes(p))
}

// 响应拦截器：解包 {code, data, message}；code:200 直接返回 data，
// code !== 200 时构造一个带 serverMessage / businessCode 的 Error 抛出，
// 由统一的错误拦截器处理 toast。
http.interceptors.response.use(
  (response) => {
    const { data } = response
    if (data && typeof data === 'object' && 'code' in data) {
      if (data.code === 200) {
        return data.data
      }
      // 业务错误：构造带消息的错误抛出（不直接 toast，避免与错误拦截器重复）
      const businessErr = new Error(data.message || '请求失败') as Error & {
        businessCode?: number
        serverMessage?: string
        config?: InternalAxiosRequestConfig
      }
      businessErr.businessCode = data.code
      businessErr.serverMessage = data.message
      businessErr.config = response.config
      return Promise.reject(businessErr)
    }
    return data
  },
  async (error: AxiosError) => {
    const status = error?.response?.status
    const originalConfig = error.config as
      | (InternalAxiosRequestConfig & { _retry?: boolean })
      | undefined
    const url = originalConfig?.url || ''
    const serverMsg =
      (error?.response?.data as { message?: string } | undefined)?.message

    // ---- 401 处理：区分"业务错误"与"token 过期" ----
    if (status === 401) {
      // 鉴权白名单接口（登录/注册/刷新）本身的 401 是业务错误（邮箱或密码错误、refresh 失败等）
      if (isAuthEndpoint(url) || (error as { businessCode?: number }).businessCode === 401) {
        ElMessage.error(serverMsg || '登录失败')
        return Promise.reject(error)
      }

      // 其他接口的 401：token 过期，尝试 refresh 单飞
      if (!originalConfig || originalConfig._retry) {
        ElMessage.error('登录已过期，请重新登录')
        return Promise.reject(error)
      }

      const store = getUserStore?.()
      if (!store) {
        ElMessage.error('登录已过期，请重新登录')
        return Promise.reject(error)
      }

      try {
        if (!refreshingPromise) {
          refreshingPromise = store
            .refresh()
            .then(() => store.token())
            .finally(() => {
              setTimeout(() => {
                refreshingPromise = null
              }, 0)
            })
        }
        const newToken = await refreshingPromise
        originalConfig._retry = true
        originalConfig.headers = originalConfig.headers || {}
        ;(originalConfig.headers as Record<string, string>)['Authorization'] =
          `Bearer ${newToken}`
        return http(originalConfig)
      } catch {
        // refresh 失败 → 清 token 并跳登录
        try {
          store.clear()
          const router = (await import('@/router')).default
          const currentPath =
            router.currentRoute.value.fullPath || '/'
          const safeRedirect = isSafeRedirect(currentPath)
          if (router.currentRoute.value.name !== 'Login') {
            router.replace({
              name: 'Login',
              query: safeRedirect ? { redirect: currentPath } : undefined,
            })
          }
        } catch {
          // 静默
        }
        ElMessage.error('登录已过期，请重新登录')
      }
      return Promise.reject(error)
    }

    // ---- 403 / 5xx / 业务错误 / 网络错误 ----
    if (status === 403) {
      ElMessage.error('没有访问权限')
    } else if (status && status >= 500) {
      ElMessage.error('服务器错误，请稍后再试')
    } else {
      // 业务错误（构造的 Error 没有 .response）或网络异常
      const businessMsg = (error as { serverMessage?: string }).serverMessage
      ElMessage.error(businessMsg || serverMsg || error.message || '网络异常')
    }
    return Promise.reject(error)
  },
)

/** 仅接受站内相对路径（/ 开头且非 // 开头），防止 open-redirect。 */
function isSafeRedirect(path: string): boolean {
  return typeof path === 'string' && /^\/(?!\/)/.test(path)
}

export interface ApiResponse<T = unknown> {
  code: number
  data: T
  message?: string
}

export interface PageRequest {
  page?: number
  page_size?: number
  keyword?: string
}

export interface PageResult<T = unknown> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export type RequestConfig = AxiosRequestConfig

export default http