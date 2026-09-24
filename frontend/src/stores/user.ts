/**
 * 用户鉴权状态（Pinia 组合式）
 *
 * 持久化策略：直接读写 localStorage（key 前缀 ph_），不引入持久化插件。
 * 退出登录时调用 clear() 重置 state 并清空 localStorage。
 */
import { ref } from 'vue'
import { defineStore } from 'pinia'
import { authApi, meApi } from '@/api'
import type {
  LoginPayload,
  RegisterPayload,
  UpdateProfilePayload,
  User,
} from '@/types'

const STORAGE = {
  access: 'ph_access',
  refresh: 'ph_refresh',
  user: 'ph_user',
  exp: 'ph_exp',
} as const

function readAccess(): string {
  return localStorage.getItem(STORAGE.access) || ''
}
function readRefresh(): string {
  return localStorage.getItem(STORAGE.refresh) || ''
}
function readUser(): User | null {
  const raw = localStorage.getItem(STORAGE.user)
  if (!raw) return null
  try {
    return JSON.parse(raw) as User
  } catch {
    return null
  }
}
function readExpiresAt(): number {
  return Number(localStorage.getItem(STORAGE.exp) || 0)
}

export const useUserStore = defineStore('user', () => {
  const accessToken = ref<string>(readAccess())
  const refreshToken = ref<string>(readRefresh())
  const expiresAt = ref<number>(readExpiresAt())
  const userInfo = ref<User | null>(readUser())

  /**
   * 当前是否已登录且 access token 未过期。
   *
   * 实现为方法而非 computed：调试发现 Pinia setup-style store 的 computed 在
   * useUserStore() 跨调用时存在缓存/失效异常（hasAccess=true 且 expiresAt=值
   * 都对，但 computed 返回 false）。改为方法后每次重新求值，不缓存，行为稳定。
   *
   * 安全缓冲：提前 60 秒认为过期，避免边界时间点后端已 401 但前端仍以为登录。
   */
  function isLoggedIn(): boolean {
    return !!accessToken.value && expiresAt.value > Date.now() + 60_000
  }

  /** 用于 axios 拦截器注入 Authorization */
  function token(): string {
    return accessToken.value
  }

  /** 设置 token 并写入 localStorage */
  function setTokens(access: string, refresh: string, expiresIn: number) {
    const exp = Date.now() + expiresIn * 1000
    // 先写 localStorage（保证持久化先于任何异常路径）
    try {
      localStorage.setItem(STORAGE.access, access)
      localStorage.setItem(STORAGE.refresh, refresh)
      localStorage.setItem(STORAGE.exp, String(exp))
    } catch (e) {
      console.warn('[userStore] 写入 localStorage 失败', e)
    }
    // 再更新 store 响应式状态
    accessToken.value = access
    refreshToken.value = refresh
    expiresAt.value = exp
    // 兜底再写一次（防御中间某步抛错）
    try {
      localStorage.setItem(STORAGE.access, access)
      localStorage.setItem(STORAGE.refresh, refresh)
      localStorage.setItem(STORAGE.exp, String(exp))
    } catch {}
  }

  /** 清空所有状态与 localStorage */
  function clear() {
    accessToken.value = ''
    refreshToken.value = ''
    expiresAt.value = 0
    userInfo.value = null
    try {
      localStorage.removeItem(STORAGE.access)
      localStorage.removeItem(STORAGE.refresh)
      localStorage.removeItem(STORAGE.user)
      localStorage.removeItem(STORAGE.exp)
    } catch {}
  }

  /** 登录 */
  async function login(payload: LoginPayload): Promise<void> {
    const res = await authApi.login(payload)
    console.log('[userStore] login 响应 expires_in =', res.expires_in, '（应为 7200 秒 = 2 小时）')
    applyAuthResult(res)
  }

  /** 注册（成功后自动登录） */
  async function register(payload: RegisterPayload): Promise<void> {
    const res = await authApi.register(payload)
    console.log('[userStore] register 响应 expires_in =', res.expires_in, '（应为 7200 秒 = 2 小时）')
    applyAuthResult(res)
  }

  /** 用 refresh token 刷新 access token */
  async function refresh(): Promise<void> {
    if (!refreshToken.value) {
      throw new Error('无 refresh token')
    }
    const res = await authApi.refreshToken(refreshToken.value)
    setTokens(res.access_token, res.refresh_token || refreshToken.value, res.expires_in)
    if (res.user) {
      userInfo.value = res.user
      localStorage.setItem(STORAGE.user, JSON.stringify(res.user))
    }
  }

  /** 退出登录（前端清 token；后端无副作用） */
  async function logout(): Promise<void> {
    clear()
  }

  /**
   * 从 /me 拉取最新个人资料并刷新本地 userInfo。
   * 用于：进入个人设置页时确保数据最新；登录态有效但本地 userInfo 缺失时补全。
   */
  async function fetchProfile(): Promise<User> {
    const u = await meApi.getMe()
    userInfo.value = u
    localStorage.setItem(STORAGE.user, JSON.stringify(u))
    return u
  }

  /**
   * 更新个人资料（当前仅支持昵称），成功后刷新本地 userInfo 并持久化。
   */
  async function updateProfile(payload: UpdateProfilePayload): Promise<User> {
    const u = await meApi.updateMe(payload)
    userInfo.value = u
    localStorage.setItem(STORAGE.user, JSON.stringify(u))
    return u
  }

  function applyAuthResult(res: { user: User; access_token: string; refresh_token: string; expires_in: number }) {
    setTokens(res.access_token, res.refresh_token, res.expires_in)
    userInfo.value = res.user
    localStorage.setItem(STORAGE.user, JSON.stringify(res.user))
  }

  return {
    // state
    accessToken,
    refreshToken,
    expiresAt,
    userInfo,
    // getters（每次重新求值，无缓存）
    isLoggedIn,
    token,
    // actions
    login,
    register,
    logout,
    refresh,
    fetchProfile,
    updateProfile,
    clear,
    setTokens,
  }
})