/**
 * 个人资料相关 API（P09）+ 修改密码（P10）。
 *
 * - GET  /api/v1/me                     取当前用户完整资料
 * - PATCH /api/v1/me                    修改昵称（仅昵称；邮箱 / 注册时间不可改）
 * - POST /api/v1/auth/change-password   修改密码（成功后旧 refresh_token 失效）
 *
 * URL 约定：`http.ts` baseURL='/api'，这里只写 '/v1/...'。
 */
import http from '../http'
import type { ChangePasswordPayload, UpdateProfilePayload, User } from '@/types'

/** GET /api/v1/me */
export function getMe(): Promise<User> {
  return http.get('/v1/me') as unknown as Promise<User>
}

/** PATCH /api/v1/me */
export function updateMe(payload: UpdateProfilePayload): Promise<User> {
  return http.patch('/v1/me', payload) as unknown as Promise<User>
}

/** POST /api/v1/auth/change-password */
export function changePassword(payload: ChangePasswordPayload): Promise<void> {
  return http.post('/v1/auth/change-password', payload) as unknown as Promise<void>
}