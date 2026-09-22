/**
 * Axios HTTP 客户端封装
 *
 * 约定：
 * - 请求统一带 baseURL 与凭证
 * - 响应统一解包 { code, data, message }，由业务侧根据 code 判断
 * - 全局处理 401、网络错误等通用错误
 *
 * 真正的业务接口定义在 src/api/modules/*，本文件只提供基础设施。
 */
import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'
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

// 请求拦截器：后续接 token 时在此统一注入 Authorization
http.interceptors.request.use(
  (config) => {
    // const token = localStorage.getItem('access_token')
    // if (token) config.headers.Authorization = `Bearer ${token}`
    return config
  },
  (error) => Promise.reject(error),
)

// 响应拦截器：解包后端约定的统一响应结构
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
  (error) => {
    const status = error?.response?.status
    if (status === 401) {
      ElMessage.error('登录已过期，请重新登录')
    } else if (status === 403) {
      ElMessage.error('没有访问权限')
    } else if (status && status >= 500) {
      ElMessage.error('服务器错误，请稍后再试')
    } else {
      ElMessage.error(error.message || '网络异常')
    }
    return Promise.reject(error)
  },
)

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
