/**
 * 用户鉴权状态（Pinia 组合式）
 *
 * 持久化策略：直接读写 localStorage（key 前缀 ph_），不引入持久化插件。
 * 退出登录时调用 clear() 重置 state 并清空 localStorage。
 */
import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { authApi } from '@/api'
import type { LoginPayload, RegisterPayload, User } from '@/types'

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

  /** 当前是否已登录且 access token 未过期 */
  const isLoggedIn = computed<boolean>(
    () => !!accessToken.value && expiresAt.value > Date.now(),
  )

  /** 用于 axios 拦截器注入 Authorization */
  const token = computed<string>(() => accessToken.value)

  /** 设置 token 并写入 localStorage */
  function setTokens(access: string, refresh: string, expiresIn: number) {
    accessToken.value = access
    refreshToken.value = refresh
    expiresAt.value = Date.now() + expiresIn * 1000
    localStorage.setItem(STORAGE.access, access)
    localStorage.setItem(STORAGE.refresh, refresh)
    localStorage.setItem(STORAGE.exp, String(expiresAt.value))
  }

  /** 清空所有状态与 localStorage */
  function clear() {
    accessToken.value = ''
    refreshToken.value = ''
    expiresAt.value = 0
    userInfo.value = null
    localStorage.removeItem(STORAGE.access)
    localStorage.removeItem(STORAGE.refresh)
    localStorage.removeItem(STORAGE.user)
    localStorage.removeItem(STORAGE.exp)
  }

  /** 登录 */
  async function login(payload: LoginPayload): Promise<void> {
    const res = await authApi.login(payload)
    applyAuthResult(res)
  }

  /** 注册（成功后自动登录） */
  async function register(payload: RegisterPayload): Promise<void> {
    const res = await authApi.register(payload)
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
    // getters
    isLoggedIn,
    token,
    // actions
    login,
    register,
    logout,
    refresh,
    clear,
    setTokens,
  }
})