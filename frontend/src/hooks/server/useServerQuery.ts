import { useQuery, type QueryKey } from '@tanstack/react-query'
import { AppError } from '../../services/http/errors'

interface ServerEnvelope<T> {
  value: T
  isDemo: boolean
}

interface ServerQueryOptions<T> {
  queryKey: QueryKey
  queryFn: () => Promise<T>
  fallback: () => T
  enabled?: boolean
  staleTime?: number
  refetchInterval?: number | false
}

export interface ServerQueryResult<T> {
  data: T | undefined
  isDemo: boolean
  isPending: boolean
  isFetching: boolean
  error: AppError | null
  refetch: () => Promise<unknown>
}

export function useServerQuery<T>(options: ServerQueryOptions<T>): ServerQueryResult<T> {
  const query = useQuery<ServerEnvelope<T>>({
    queryKey: options.queryKey,
    enabled: options.enabled,
    staleTime: options.staleTime,
    refetchInterval: options.refetchInterval,
    queryFn: async () => {
      try {
        return { value: await options.queryFn(), isDemo: false }
      } catch (cause) {
        if (cause instanceof AppError && cause.kind === 'network') {
          return { value: options.fallback(), isDemo: true }
        }
        throw cause
      }
    },
  })
  return {
    data: query.data?.value,
    isDemo: query.data?.isDemo ?? false,
    isPending: query.isPending,
    isFetching: query.isFetching,
    error: query.error instanceof AppError ? query.error : null,
    refetch: query.refetch,
  }
}
