import { Navigate, Outlet, useLocation } from 'react-router-dom'
import type { PropsWithChildren } from 'react'
import { isAdminRole } from '../../auth'
import { useAuth } from '../providers'
import type { Role } from '../../types'

export function RequireAuth() {
  const { session } = useAuth()
  const location = useLocation()
  if (!session) return <Navigate to={`/login?redirect=${encodeURIComponent(location.pathname + location.search)}`} replace />
  if (session.role !== 'USER') return <Navigate to="/forbidden" replace />
  return <Outlet />
}

// Public catalog pages remain available to guests, but an admin session must stay in the admin workspace.
export function RequireUserArea({ children }: PropsWithChildren) {
  const { session } = useAuth()
  if (session && isAdminRole(session.role)) return <Navigate to="/admin/dashboard" replace />
  return children ?? <Outlet />
}

export function RequireRole({ roles, children }: { roles: Role[]; children?: React.ReactNode }) {
  const { session } = useAuth()
  if (!session) return <Navigate to="/login?mode=platform" replace />
  if (!isAdminRole(session.role) || !roles.includes(session.role)) return <Navigate to="/forbidden" replace />
  return children ?? <Outlet />
}
