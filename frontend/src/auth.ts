import type { AuthSession, Role } from './types'

const STORAGE_KEY = 'lterm.auth'

export function getStoredAuth(): AuthSession | null {
  const raw = sessionStorage.getItem(STORAGE_KEY)
  if (!raw) {
    return null
  }
  try {
    const value = JSON.parse(raw) as AuthSession
    if (!value.token || !value.role) {
      return null
    }
    return value
  } catch {
    return null
  }
}

export function setStoredAuth(session: AuthSession) {
  sessionStorage.setItem(STORAGE_KEY, JSON.stringify(session))
}

export function clearStoredAuth() {
  sessionStorage.removeItem(STORAGE_KEY)
}

export function isAdminRole(role: Role | undefined) {
  return role === 'SUPER_ADMIN' || role === 'CINEMA_ADMIN' || role === 'FINANCE'
}

export function roleLabel(role: Role | undefined) {
  if (role === 'SUPER_ADMIN') return '平台管理员'
  if (role === 'CINEMA_ADMIN') return '影院运营'
  if (role === 'FINANCE') return '财务分析'
  return '观众'
}
