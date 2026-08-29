import axios, { type AxiosError, type AxiosInstance, type InternalAxiosRequestConfig } from 'axios'
import { clearStoredAuth, getStoredAuth } from '../../auth'
import type { ApiEnvelope } from '../../types'
import { AppError, classifyError } from './errors'

const API_BASE = (import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api/v1').replace(/\/$/, '')
export const AUTH_EXPIRED_EVENT = 'lterm:auth-expired'

function requestId() {
  return typeof crypto.randomUUID === 'function' ? crypto.randomUUID() : `${Date.now()}-${Math.random().toString(16).slice(2)}`
}

function addHeaders(config: InternalAxiosRequestConfig) {
  const auth = getStoredAuth()
  config.headers.set('X-Request-ID', requestId())
  if (auth?.token) {
    config.headers.set('Authorization', `Bearer ${auth.token}`)
  }
  return config
}

function toAppError(error: AxiosError<ApiEnvelope<unknown>>): AppError {
  if (!error.response) {
    return new AppError('网络连接失败，请检查服务是否启动', 0, 0, 'network')
  }
  const status = error.response.status
  const payload = error.response.data
  const requestIdHeader = error.response.headers['x-request-id'] as string | undefined
  return new AppError(payload?.msg ?? '请求失败，请稍后重试', status, payload?.code ?? status, classifyError(status), requestIdHeader)
}

export const http: AxiosInstance = axios.create({
  baseURL: API_BASE,
  timeout: 12_000,
  headers: {
    'Content-Type': 'application/json',
  },
})

http.interceptors.request.use(addHeaders)

http.interceptors.response.use(
  (response) => {
    const payload = response.data as ApiEnvelope<unknown> | undefined
    if (payload && typeof payload.code === 'number') {
      if (payload.code !== 0) {
        throw new AppError(payload.msg || '请求失败', response.status, payload.code, classifyError(response.status))
      }
      response.data = payload.data
    }
    return response
  },
  (error: AxiosError<ApiEnvelope<unknown>>) => {
    const appError = toAppError(error)
    if (appError.kind === 'auth') {
      clearStoredAuth()
      window.dispatchEvent(new CustomEvent(AUTH_EXPIRED_EVENT))
    }
    return Promise.reject(appError)
  },
)

export function withIdempotency(config: InternalAxiosRequestConfig, key: string) {
  config.headers.set('Idempotency-Key', key)
  return config
}

export function getApiBaseUrl() {
  return API_BASE
}
