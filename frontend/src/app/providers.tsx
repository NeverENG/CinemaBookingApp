import { QueryClient, QueryClientProvider, useQueryClient } from '@tanstack/react-query'
import { createContext, useCallback, useContext, useEffect, useMemo, useState, type PropsWithChildren } from 'react'
import { clearStoredAuth, getStoredAuth, setStoredAuth } from '../auth'
import { AUTH_EXPIRED_EVENT } from '../services/http/client'
import type { AuthSession } from '../types'

interface AuthContextValue {
  session: AuthSession | null
  signIn: (session: AuthSession) => void
  signOut: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: PropsWithChildren) {
  const queryClient = useQueryClient()
  const [session, setSession] = useState<AuthSession | null>(() => getStoredAuth())

  useEffect(() => {
    const handleExpired = () => {
      queryClient.clear()
      setSession(null)
    }
    window.addEventListener(AUTH_EXPIRED_EVENT, handleExpired)
    return () => window.removeEventListener(AUTH_EXPIRED_EVENT, handleExpired)
  }, [queryClient])

  const value = useMemo<AuthContextValue>(() => ({
    session,
    signIn: (next) => {
      queryClient.clear()
      setStoredAuth(next)
      setSession(next)
    },
    signOut: () => {
      queryClient.clear()
      clearStoredAuth()
      setSession(null)
    },
  }), [queryClient, session])

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const value = useContext(AuthContext)
  if (!value) throw new Error('useAuth must be used inside AuthProvider')
  return value
}

interface CinemaContextValue {
  cinemaId: number
  setCinemaId: (cinemaId: number) => void
}

const CinemaContext = createContext<CinemaContextValue | null>(null)
const CINEMA_KEY = 'lterm.cinema_id'

export function CinemaProvider({ children }: PropsWithChildren) {
  const [cinemaId, setCinemaState] = useState<number>(() => Number(localStorage.getItem(CINEMA_KEY)) || 1)
  const setCinemaId = useCallback((next: number) => {
    if (!Number.isInteger(next) || next <= 0) return
    localStorage.setItem(CINEMA_KEY, String(next))
    setCinemaState(next)
  }, [])
  const value = useMemo<CinemaContextValue>(() => ({ cinemaId, setCinemaId }), [cinemaId, setCinemaId])
  return <CinemaContext.Provider value={value}>{children}</CinemaContext.Provider>
}

export function useCinema() {
  const value = useContext(CinemaContext)
  if (!value) throw new Error('useCinema must be used inside CinemaProvider')
  return value
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60_000,
      refetchOnWindowFocus: false,
      retry: false,
    },
    mutations: {
      retry: false,
    },
  },
})

export function AppProviders({ children }: PropsWithChildren) {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <CinemaProvider>{children}</CinemaProvider>
      </AuthProvider>
    </QueryClientProvider>
  )
}
