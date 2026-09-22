/**
 * Axios HTTP 客户端封装
 *
 * 约定：
 * - 请求统一带 baseURL 与凭证
 * - 响应统一解包 { code, data, message }，由业务侧根据 code 判断
 * - 401 时尝试静默 refresh（白名单跳过 /auth/refresh 与 /auth/login），单飞避免并发刷新
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

// 请求拦截器：从 Pinia store 读取 access token 注入 Authorization
http.interceptors.request.use(
  (config) => {
    // 动态 import 避免在 pinia 未安装时调用 useUserStore()
    // （http.ts 在 src/api/* 中，可能被其它模块早于 main.ts 的 app.use(pinia) 引用）
    try {
      // eslint-disable-next-line @typescript-eslint/no-var-requires
      const { useUserStore } = require('@/stores/user')
      const userStore = useUserStore()
      if (userStore.token) {
        config.headers = config.headers || {}
        ;(config.headers as Record<string, string>)['Authorization'] =
          `Bearer ${userStore.token}`
      }
    } catch {
      // 第一次冷启动时 pinia 尚未安装，跳过 token 注入
    }
    return config
  },
  (error) => Promise.reject(error),
)

// 刷新 token 单飞锁：同一时刻只允许一个 refresh 在途
let refreshingPromise: Promise<string> | null = null

// 响应拦截器：解包 + 401 静默刷新
http.interceptors.response.use(
  (response) => {
    const { data } = response
    // 后端约定格式：{ code: 200, data: any, message: string }
    if (data && typeof data === 'object' && 'code' in data) {
      if (data.code === 200) {
        return data.data
      }
      ElMessage.error(data.message || '请求失败')
      return Promise.reject(new Error(data.message || '请求失败'))
    }
    return data
  },
  async (error: AxiosError) => {
    const status = error?.response?.status
    const originalConfig = error.config as
      | (InternalAxiosRequestConfig & { _retry?: boolean })
      | undefined

    if (status === 401 && originalConfig) {
      const url = originalConfig.url || ''
      // 公开鉴权接口本身不要触发刷新逻辑
      const isAuthEndpoint =
        url.includes('/auth/refresh') ||
        url.includes('/auth/login') ||
        url.includes('/auth/register')

      if (isAuthEndpoint || originalConfig._retry) {
        // 登录失败 / 刷新失败：交给上层 Toast + 路由守卫处理
      } else {
        try {
          if (!refreshingPromise) {
            // eslint-disable-next-line @typescript-eslint/no-var-requires
            const { useUserStore } = require('@/stores/user')
            const userStore = useUserStore()
            refreshingPromise = userStore
              .refresh()
              .then(() => userStore.token)
              .finally(() => {
                // 等待当前请求头更新后再清空引用
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
          // 刷新失败 → 清 token 并跳登录
          try {
            // eslint-disable-next-line @typescript-eslint/no-var-requires
            const { useUserStore } = require('@/stores/user')
            const userStore = useUserStore()
            userStore.clear()
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
        }
      }
    }

    // 通用错误提示
    if (status === 401) {
      ElMessage.error('登录已过期，请重新登录')
    } else if (status === 403) {
      ElMessage.error('没有访问权限')
    } else if (status && status >= 500) {
      ElMessage.error('服务器错误，请稍后再试')
    } else {
      const serverMsg =
        (error?.response?.data as { message?: string } | undefined)?.message
      ElMessage.error(serverMsg || error.message || '网络异常')
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