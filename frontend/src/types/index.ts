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

/* ---------- 提示词 / 分类 ---------- */

/** 提示词 */
export interface Prompt {
  id: number
  user_id: number
  category_id?: number | null
  title: string
  description: string
  content: string
  is_favorite: boolean
  tags: string[]
  version_no: number
  created_at: string
  updated_at: string
}

/** 分类 */
export interface Category {
  id: number
  user_id: number
  name: string
  color: string
  sort: number
  created_at: string
  updated_at: string
  prompt_count?: number
}

/** 列表查询参数 */
export interface PromptListQuery {
  page?: number
  page_size?: number
  keyword?: string
  q_in?: 'title' | 'all'
  category_id?: number | null
  is_favorite?: boolean
  sort?: 'updated_desc' | 'created_desc' | 'title_asc'
}

/** 列表返回（与 response.Page(c, ...) 一致） */
export interface PromptListResult {
  list: Prompt[]
  total: number
  page: number
  page_size: number
}

/** 创建提示词 */
export interface CreatePromptPayload {
  title: string
  description?: string
  content: string
  category_id?: number | null
  tags?: string[]
}

/** 更新提示词（所有字段可选） */
export interface UpdatePromptPayload {
  title?: string
  description?: string
  content?: string
  category_id?: number | null
  tags?: string[]
  is_favorite?: boolean
}

/** 创建分类 */
export interface CreateCategoryPayload {
  name: string
  color?: string
  sort?: number
}

/** 更新分类 */
export interface UpdateCategoryPayload {
  name?: string
  color?: string
  sort?: number
}

/** 批量删除 */
export interface BatchDeletePayload {
  ids: number[]
}

/** 切换收藏 */
export interface ToggleFavoritePayload {
  is_favorite: boolean
}