import { http } from '../http/client'
import { normalizeAuth } from './normalize'

export const authApi = {
  userLogin: (payload: { email: string; password: string }) => http.post('/auth/login', payload).then((response) => normalizeAuth(response.data)),
  adminLogin: (payload: { username: string; password: string }) => http.post('/admin/auth/login', payload).then((response) => normalizeAuth(response.data)),
  requestRegistrationCode: (email: string) => http.post<{ status: string; dev_code?: string }>('/auth/email-verification/request', { email }).then((response) => response.data),
  register: (payload: { email: string; code: string; password: string; nickname: string }) => http.post('/auth/register', payload).then((response) => normalizeAuth(response.data)),
  requestPasswordReset: (email: string) => http.post<{ status: string; dev_code?: string }>('/auth/password-reset/request', { email }).then((response) => response.data),
  resetPassword: (payload: { email: string; code: string; new_password: string }) => http.post<{ status: string }>('/auth/password-reset/reset', payload).then((response) => response.data),
  changePassword: (payload: { old_password: string; new_password: string }) => http.post<{ status: string }>('/me/password', payload).then((response) => response.data),
  changeAdminPassword: (payload: { old_password: string; new_password: string }) => http.post<{ status: string }>('/admin/me/password', payload).then((response) => response.data),
}
