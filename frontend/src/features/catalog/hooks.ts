import { useQueryClient } from '@tanstack/react-query'
import { catalogApi, catalogFallbacks } from '../../services/api'
import { useServerQuery } from '../../hooks/server/useServerQuery'

export function useHomeQuery() {
  return useServerQuery({ queryKey: ['home'], queryFn: catalogApi.home, fallback: catalogFallbacks.home, staleTime: 120_000 })
}

export function useMovieSearchQuery(keyword: string) {
  return useServerQuery({ queryKey: ['movies', 'search', keyword], queryFn: () => catalogApi.searchMovies(keyword), fallback: () => catalogFallbacks.movies(keyword), enabled: keyword.trim().length > 0, staleTime: 60_000 })
}

export function useMovieQuery(movieId: number) {
  return useServerQuery({ queryKey: ['movie', movieId], queryFn: () => catalogApi.getMovie(movieId), fallback: () => catalogFallbacks.movie(movieId), enabled: movieId > 0, staleTime: 120_000 })
}

export function useCinemaQuery(keyword = '') {
  return useServerQuery({ queryKey: ['cinemas', keyword], queryFn: () => catalogApi.cinemas({ keyword }), fallback: () => catalogFallbacks.cinemas(keyword), staleTime: 300_000 })
}

export function useSessionQuery(movieId?: number, cinemaId?: number) {
  return useServerQuery({ queryKey: ['sessions', { movieId, cinemaId }], queryFn: () => catalogApi.sessions({ movieId, cinemaId }), fallback: () => catalogFallbacks.sessions(movieId, cinemaId), staleTime: 30_000 })
}

export function useInvalidateCatalog() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: ['home'] })
}
