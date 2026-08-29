export type AppErrorKind = 'auth' | 'forbidden' | 'not-found' | 'conflict' | 'validation' | 'network' | 'server' | 'unknown'

export class AppError extends Error {
  readonly status: number
  readonly code: number
  readonly kind: AppErrorKind
  readonly requestId?: string

  constructor(message: string, status: number, code: number, kind: AppErrorKind, requestId?: string) {
    super(message)
    this.name = 'AppError'
    this.status = status
    this.code = code
    this.kind = kind
    this.requestId = requestId
  }
}

export function classifyError(status: number): AppErrorKind {
  if (status === 401) return 'auth'
  if (status === 403) return 'forbidden'
  if (status === 404) return 'not-found'
  if (status === 409) return 'conflict'
  if (status === 400 || status === 422) return 'validation'
  if (status >= 500) return 'server'
  return 'unknown'
}
