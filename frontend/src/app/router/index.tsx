import { lazy, Suspense, type ComponentType, type LazyExoticComponent } from 'react'
import { createBrowserRouter, Navigate } from 'react-router-dom'
import { AdminLayout } from '../../layouts/AdminLayout'
import { UserLayout } from '../../layouts/UserLayout'
import { RequireAuth, RequireRole } from './guards'

const LoginPage = lazy(() => import('../../pages/auth/LoginPage').then(({ LoginPage: page }) => ({ default: page })))
const ForgotPasswordPage = lazy(() => import('../../pages/auth/ForgotPasswordPage').then(({ ForgotPasswordPage: page }) => ({ default: page })))
const RegisterPage = lazy(() => import('../../pages/auth/RegisterPage').then(({ RegisterPage: page }) => ({ default: page })))
const AdminsPage = lazy(() => import('../../pages/admin/AdminsPage').then(({ AdminsPage: page }) => ({ default: page })))
const DashboardPage = lazy(() => import('../../pages/admin/DashboardPage').then(({ DashboardPage: page }) => ({ default: page })))
const HallsPage = lazy(() => import('../../pages/admin/HallsPage').then(({ HallsPage: page }) => ({ default: page })))
const MarketingPage = lazy(() => import('../../pages/admin/MarketingPage').then(({ MarketingPage: page }) => ({ default: page })))
const MoviesPage = lazy(() => import('../../pages/admin/MoviesPage').then(({ MoviesPage: page }) => ({ default: page })))
const SessionsPage = lazy(() => import('../../pages/admin/SessionsPage').then(({ SessionsPage: page }) => ({ default: page })))
const TicketsPage = lazy(() => import('../../pages/admin/TicketsPage').then(({ TicketsPage: page }) => ({ default: page })))
const CheckoutPage = lazy(() => import('../../pages/user/CheckoutPage').then(({ CheckoutPage: page }) => ({ default: page })))
const CinemasPage = lazy(() => import('../../pages/user/CinemasPage').then(({ CinemasPage: page }) => ({ default: page })))
const HomePage = lazy(() => import('../../pages/user/HomePage').then(({ HomePage: page }) => ({ default: page })))
const MovieDetailPage = lazy(() => import('../../pages/user/MovieDetailPage').then(({ MovieDetailPage: page }) => ({ default: page })))
const OrderDetailPage = lazy(() => import('../../pages/user/OrderDetailPage').then(({ OrderDetailPage: page }) => ({ default: page })))
const ChangeTicketPage = lazy(() => import('../../pages/user/ChangeTicketPage').then(({ ChangeTicketPage: page }) => ({ default: page })))
const OrdersPage = lazy(() => import('../../pages/user/OrdersPage').then(({ OrdersPage: page }) => ({ default: page })))
const PaymentPage = lazy(() => import('../../pages/user/PaymentPage').then(({ PaymentPage: page }) => ({ default: page })))
const ProfilePage = lazy(() => import('../../pages/user/ProfilePage').then(({ ProfilePage: page }) => ({ default: page })))
const RewardsPage = lazy(() => import('../../pages/user/RewardsPage').then(({ RewardsPage: page }) => ({ default: page })))
const SearchPage = lazy(() => import('../../pages/user/SearchPage').then(({ SearchPage: page }) => ({ default: page })))
const SeatSelectionPage = lazy(() => import('../../pages/user/SeatSelectionPage').then(({ SeatSelectionPage: page }) => ({ default: page })))
const ForbiddenPage = lazy(() => import('../../pages/system/FeedbackPages').then(({ ForbiddenPage: page }) => ({ default: page })))
const NotFoundPage = lazy(() => import('../../pages/system/FeedbackPages').then(({ NotFoundPage: page }) => ({ default: page })))

type LazyPage = LazyExoticComponent<ComponentType>

function lazyElement(Page: LazyPage) {
  return <Suspense fallback={<div className="route-loading">加载页面中…</div>}><Page /></Suspense>
}

export const router = createBrowserRouter([
  { path: '/login', element: lazyElement(LoginPage) },
  { path: '/forgot-password', element: lazyElement(ForgotPasswordPage) },
  { path: '/register', element: lazyElement(RegisterPage) },
  { path: '/forbidden', element: lazyElement(ForbiddenPage) },
  {
    path: '/',
    element: <UserLayout />,
    children: [
      { index: true, element: lazyElement(HomePage) },
      { path: 'recommend', element: lazyElement(HomePage) },
      { path: 'search', element: lazyElement(SearchPage) },
      { path: 'cinemas', element: lazyElement(CinemasPage) },
      { path: 'movies/:movieId', element: lazyElement(MovieDetailPage) },
      { path: 'sessions/:sessionId/seats', element: lazyElement(SeatSelectionPage) },
      { element: <RequireAuth />, children: [
        { path: 'checkout', element: lazyElement(CheckoutPage) },
        { path: 'payment/:orderNo', element: lazyElement(PaymentPage) },
        { path: 'orders', element: lazyElement(OrdersPage) },
        { path: 'orders/:orderNo', element: lazyElement(OrderDetailPage) },
        { path: 'orders/:orderNo/change', element: lazyElement(ChangeTicketPage) },
        { path: 'rewards', element: lazyElement(RewardsPage) },
        { path: 'profile', element: lazyElement(ProfilePage) },
      ] },
    ],
  },
  {
    path: '/admin',
    element: <RequireRole roles={['SUPER_ADMIN', 'CINEMA_ADMIN', 'FINANCE']}><AdminLayout /></RequireRole>,
    children: [
      { index: true, element: <Navigate to="dashboard" replace /> },
      { path: 'dashboard', element: <RequireRole roles={['SUPER_ADMIN', 'CINEMA_ADMIN', 'FINANCE']}>{lazyElement(DashboardPage)}</RequireRole> },
      { path: 'movies', element: <RequireRole roles={['SUPER_ADMIN']}>{lazyElement(MoviesPage)}</RequireRole> },
      { path: 'halls', element: <RequireRole roles={['SUPER_ADMIN', 'CINEMA_ADMIN']}>{lazyElement(HallsPage)}</RequireRole> },
      { path: 'sessions', element: <RequireRole roles={['SUPER_ADMIN', 'CINEMA_ADMIN']}>{lazyElement(SessionsPage)}</RequireRole> },
      { path: 'tickets', element: <RequireRole roles={['SUPER_ADMIN', 'CINEMA_ADMIN']}>{lazyElement(TicketsPage)}</RequireRole> },
      { path: 'marketing', element: <RequireRole roles={['SUPER_ADMIN']}>{lazyElement(MarketingPage)}</RequireRole> },
      { path: 'admins', element: <RequireRole roles={['SUPER_ADMIN']}>{lazyElement(AdminsPage)}</RequireRole> },
      { path: 'profile', element: <RequireRole roles={['SUPER_ADMIN', 'CINEMA_ADMIN', 'FINANCE']}>{lazyElement(ProfilePage)}</RequireRole> },
    ],
  },
  { path: '*', element: lazyElement(NotFoundPage) },
])
