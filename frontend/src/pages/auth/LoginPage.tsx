import { ArrowRight, Eye, LockKeyhole, UserRound } from 'lucide-react'
import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { isAdminRole } from '../../auth'
import { useAuth } from '../../app/providers'
import { Button, DemoBadge, Field, Input } from '../../components/ui'
import { useAsyncLock } from '../../hooks/useAsyncLock'
import { authApi } from '../../services/api'
import { AppError } from '../../services/http/errors'
import type { Role } from '../../types'

export function LoginPage() {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const { signIn } = useAuth()
  const [mode, setMode] = useState<'user' | 'admin'>(params.get('mode') === 'admin' ? 'admin' : 'user')
  const [account, setAccount] = useState(mode === 'admin' ? 'admin' : '')
  const [password, setPassword] = useState(mode === 'admin' ? 'admin123' : '')
  const [error, setError] = useState('')
  const { run, isLocked } = useAsyncLock()

  function changeMode(next: 'user' | 'admin') {
    setMode(next)
    setAccount(next === 'admin' ? 'admin' : '')
    setPassword(next === 'admin' ? 'admin123' : '')
    setError('')
  }

  async function submit(event: FormEvent) {
    event.preventDefault()
    setError('')
    await run(async () => {
      try {
        const session = mode === 'admin' ? await authApi.adminLogin({ username: account, password }) : await authApi.userLogin({ email: account, password })
        signIn(session)
        const redirect = params.get('redirect')
        navigate(redirect || (isAdminRole(session.role) ? '/admin/dashboard' : '/'))
      } catch (cause) {
        setError(cause instanceof AppError ? cause.message : '登录失败，请稍后重试')
      }
    })
  }

  function demoLogin(role: Role) {
    signIn({ token: `demo-${role.toLowerCase()}`, userId: 1, role })
    navigate(role === 'USER' ? '/' : '/admin/dashboard')
  }

  return <div className="auth-page"><div className="auth-panel"><Link to="/" className="brand"><span className="brand-mark"><Eye size={17} /></span><span>LTerm</span></Link><div className="auth-copy"><span className="eyebrow">WELCOME BACK</span><h1>把时间留给<br /><em>值得的电影。</em></h1><p>登录后继续选座，或进入影院工作台。</p></div><div className="auth-tabs"><button className={mode === 'user' ? 'active' : ''} onClick={() => changeMode('user')}><UserRound size={15} />观众登录</button><button className={mode === 'admin' ? 'active' : ''} onClick={() => changeMode('admin')}><LockKeyhole size={15} />管理端登录</button></div><form className="auth-form" onSubmit={submit}><Field label={mode === 'admin' ? '管理员账号' : '邮箱'}><Input value={account} onChange={(event) => setAccount(event.target.value)} placeholder={mode === 'admin' ? '例如 admin' : 'name@example.com'} autoComplete="username" required /></Field><Field label="密码"><Input value={password} onChange={(event) => setPassword(event.target.value)} type="password" placeholder="请输入密码" autoComplete="current-password" required /></Field>{error && <div className="form-error">{error}</div>}<Button type="submit" size="lg" disabled={isLocked}>{isLocked ? '登录中…' : '继续'}<ArrowRight size={16} /></Button></form><div className="auth-foot">{mode === 'user' ? <span>还没有账号？<Link to="/register">注册账号</Link></span> : <span>管理员账号由平台管理员创建</span>}<button className="demo-login" onClick={() => demoLogin(mode === 'user' ? 'USER' : 'SUPER_ADMIN')}><DemoBadge /> 直接体验界面</button></div></div><div className="auth-aside"><div className="auth-aside-grid" /><div className="auth-aside-content"><span className="eyebrow">LTERM CINEMA / 2026</span><h2>今天看什么，<br />由你决定。</h2><div className="auth-aside-rule" /><p>影院热映 · 全局搜索 · 实时选座</p></div></div></div>
}
