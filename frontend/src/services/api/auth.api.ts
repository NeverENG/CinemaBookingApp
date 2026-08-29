import { http } from '../http/client'
import { normalizeAuth } from './normalize'

export const authApi = {
  userLogin: (payload: { email: string; password: string }) => http.post('/auth/login', payload).then((response) => normalizeAuth(response.data)),
  adminLogin: (payload: { username: string; password: string }) => http.post('/admin/auth/login', payload).then((response) => normalizeAuth(response.data)),
  register: (payload: { email: string; password: string; nickname: string }) => http.post('/auth/register', payload).then((response) => normalizeAuth(response.data)),
  changePassword: (payload: { old_password: string; new_password: string }) => http.post<{ status: string }>('/me/password', payload).then((response) => response.data),
}
