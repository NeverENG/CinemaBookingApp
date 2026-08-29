import { http } from '../http/client'
import { demoCinemas, demoHome, demoMovies, demoSessions } from '../../demo'
import type { Cinema, HomeView, Movie, Session } from '../../types'
import { normalizeCinema, normalizeHome, normalizeMovie, normalizeSession } from './normalize'

export const catalogApi = {
  home: async (): Promise<HomeView> => normalizeHome(await http.get('/home').then((response) => response.data)),
  searchMovies: async (keyword: string): Promise<Movie[]> => {
    const raw = await http.get('/movies', { params: { keyword, status: 'ON_SALE' } }).then((response) => response.data)
    return (Array.isArray(raw) ? raw : []).map(normalizeMovie)
  },
  getMovie: async (movieId: number): Promise<Movie> => normalizeMovie(await http.get(`/movies/${movieId}`).then((response) => response.data)),
  cinemas: async (params: { keyword?: string; city?: string } = {}): Promise<Cinema[]> => {
    const raw = await http.get('/cinemas', { params }).then((response) => response.data)
    return (Array.isArray(raw) ? raw : []).map(normalizeCinema)
  },
  sessions: async (params: { movieId?: number; cinemaId?: number } = {}): Promise<Session[]> => {
    const raw = await http.get('/sessions', { params: { movie_id: params.movieId, cinema_id: params.cinemaId } }).then((response) => response.data)
    return (Array.isArray(raw) ? raw : []).map(normalizeSession)
  },
}

export const catalogFallbacks = {
  home: () => demoHome,
  movies: (keyword: string) => demoMovies.filter((movie) => movie.title.includes(keyword) || movie.genre.includes(keyword)),
  movie: (movieId: number) => demoMovies.find((movie) => movie.id === movieId) ?? demoMovies[0],
  cinemas: (keyword = '') => demoCinemas.filter((cinema) => `${cinema.name}${cinema.city}${cinema.address}`.includes(keyword)),
  sessions: (movieId?: number, cinemaId?: number) => demoSessions.filter((session) => (!movieId || session.movieId === movieId) && (!cinemaId || session.cinemaId === cinemaId)),
}
