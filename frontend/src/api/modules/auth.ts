/**
 * 鉴权相关 API
 *
 * 复用统一 http 实例（已在响应拦截器中解包 {code,data,message}），
 * 本文件只关心业务入参与出参的类型化。
 *
 * 注意：调用 http.post 时不传泛型，因为响应拦截器已经返回解包后的业务对象
 * （不是 AxiosResponse），所以方法返回类型就是业务类型本身。
 */
import http from '../http'
import type {
  AuthResult,
  LoginPayload,
  RegisterPayload,
  RefreshResult,
} from '@/types'

/** POST /api/v1/auth/register */
// 注意：http 实例 baseURL 已是 '/api'，这里只写 '/v1/...'
export function register(payload: RegisterPayload): Promise<AuthResult> {
  return http.post('/v1/auth/register', payload) as unknown as Promise<AuthResult>
}

/** POST /api/v1/auth/login */
export function login(payload: LoginPayload): Promise<AuthResult> {
  return http.post('/v1/auth/login', payload) as unknown as Promise<AuthResult>
}

/** POST /api/v1/auth/logout（需 Authorization header） */
export function logout(): Promise<void> {
  return http.post('/v1/auth/logout') as unknown as Promise<void>
}

/** POST /api/v1/auth/refresh */
export function refreshToken(refreshToken: string): Promise<RefreshResult> {
  return http.post('/v1/auth/refresh', {
    refresh_token: refreshToken,
  }) as unknown as Promise<RefreshResult>
}