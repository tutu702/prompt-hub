/**
 * 个人资料相关 API（P09）。
 *
 * - GET  /api/v1/me        取当前用户完整资料
 * - PATCH /api/v1/me       修改昵称（仅昵称；邮箱 / 注册时间不可改）
 *
 * URL 约定：`http.ts` baseURL='/api'，这里只写 '/v1/...'。
 */
import http from '../http'
import type { UpdateProfilePayload, User } from '@/types'

/** GET /api/v1/me */
export function getMe(): Promise<User> {
  return http.get('/v1/me') as unknown as Promise<User>
}

/** PATCH /api/v1/me */
export function updateMe(payload: UpdateProfilePayload): Promise<User> {
  return http.patch('/v1/me', payload) as unknown as Promise<User>
}