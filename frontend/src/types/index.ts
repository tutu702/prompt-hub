/**
 * 通用类型定义
 *
 * 业务模型（提示词 / 分类 / 用户等）按需在此或独立文件追加。
 */

/** 后端统一响应结构 */
export interface ApiResult<T = unknown> {
  code: number
  data: T
  message?: string
}

/** 分页请求 */
export interface PageQuery {
  page?: number
  page_size?: number
}

/** 分页结果 */
export interface PageData<T = unknown> {
  list: T[]
  total: number
  page: number
  page_size: number
}

/* ---------- 鉴权相关类型 ---------- */

/** 已登录用户信息 */
export interface User {
  id: number
  email: string
  nickname: string
  created_at?: string
}

/** 登录请求体 */
export interface LoginPayload {
  email: string
  password: string
  remember: boolean
}

/** 注册请求体 */
export interface RegisterPayload {
  email: string
  password: string
  confirmPassword: string
  nickname?: string
  agreement: boolean
}

/** 后端鉴权成功返回的完整数据 */
export interface AuthResult {
  user: User
  access_token: string
  refresh_token: string
  expires_in: number
}

/** 刷新 access token 返回 */
export interface RefreshResult {
  user?: User
  access_token: string
  refresh_token: string
  expires_in: number
}