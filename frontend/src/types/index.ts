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
