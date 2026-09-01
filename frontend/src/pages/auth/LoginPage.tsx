import { ArrowRight, Building2, CheckCircle2, Eye, ShieldCheck, UserRound } from 'lucide-react'
import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { isAdminRole } from '../../auth'
import { useAuth } from '../../app/providers'
import { Button, Field, Input } from '../../components/ui'
import { useAsyncLock } from '../../hooks/useAsyncLock'
import { authApi } from '../../services/api'
import { AppError } from '../../services/http/errors'

type LoginMode = 'user' | 'cinema' | 'platform'

function initialMode(value: string | null): LoginMode {
  if (value === 'cinema') return 'cinema'
  if (value === 'platform' || value === 'admin') return 'platform'
  return 'user'
}

export function LoginPage() {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const { signIn } = useAuth()
  const [mode, setMode] = useState<LoginMode>(() => initialMode(params.get('mode')))
  const [account, setAccount] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const { run, isLocked } = useAsyncLock()

  function changeMode(next: LoginMode) {
    setMode(next)
    setAccount('')
    setPassword('')
    setError('')
  }

  async function submit(event: FormEvent) {
    event.preventDefault()
    setError('')
    await run(async () => {
      try {
        const session = mode === 'user' ? await authApi.userLogin({ email: account, password }) : await authApi.adminLogin({ username: account, password })
        if (mode === 'cinema' && session.role !== 'CINEMA_ADMIN') {
          setError('该账号不是影院管理员，请切换正确入口')
          return
        }
        if (mode === 'platform' && session.role !== 'SUPER_ADMIN' && session.role !== 'FINANCE') {
          setError('该账号不属于平台工作区，请切换正确入口')
          return
        }
        signIn(session)
        const redirect = params.get('redirect')
        navigate(redirect || (isAdminRole(session.role) ? '/admin/dashboard' : '/'))
      } catch (cause) {
        setError(cause instanceof AppError ? cause.message : '登录失败，请稍后重试')
      }
    })
  }

  return <div className="auth-page"><div className="auth-panel"><Link to="/" className="brand"><span className="brand-mark"><Eye size={17} /></span><span>LTerm</span></Link><div className="auth-copy"><span className="eyebrow">WELCOME BACK</span><h1>{mode === 'user' ? <>把时间留给<br /><em>值得的电影。</em></> : mode === 'cinema' ? <>管理影院的<br /><em>每一场放映。</em></> : <>掌握平台的<br /><em>内容与经营。</em></>}</h1><p>{mode === 'user' ? '使用注册邮箱登录，继续选座与购票。' : mode === 'cinema' ? '进入绑定影院的排片、影厅、核销与数据工作台。' : '进入全局影片、运营内容、账号与票房工作台。'}</p></div><div className="auth-tabs auth-tabs-three" role="tablist" aria-label="选择登录身份"><button type="button" className={mode === 'user' ? 'active' : ''} onClick={() => changeMode('user')}><UserRound size={15} />观众</button><button type="button" className={mode === 'cinema' ? 'active' : ''} onClick={() => changeMode('cinema')}><Building2 size={15} />影院运营</button><button type="button" className={mode === 'platform' ? 'active' : ''} onClick={() => changeMode('platform')}><ShieldCheck size={15} />平台管理</button></div>{params.get('reset') === '1' && <div className="form-success"><CheckCircle2 size={14} />密码已重置，请重新登录</div>}<form className="auth-form" onSubmit={submit}><Field label={mode === 'user' ? '邮箱' : '管理员账号'}><Input value={account} onChange={(event) => setAccount(event.target.value)} type={mode === 'user' ? 'email' : 'text'} placeholder={mode === 'user' ? 'name@example.com' : '请输入管理员账号'} autoComplete="username" required /></Field><Field label="密码"><Input value={password} onChange={(event) => setPassword(event.target.value)} type="password" placeholder="请输入密码" autoComplete="current-password" required /></Field>{mode === 'user' && <div className="auth-form-link"><Link to="/forgot-password">忘记密码？</Link></div>}{error && <div className="form-error">{error}</div>}<Button type="submit" size="lg" disabled={isLocked}>{isLocked ? '登录中…' : '进入'}<ArrowRight size={16} /></Button></form><div className="auth-foot">{mode === 'user' ? <span>还没有账号？<Link to="/register">邮箱注册</Link></span> : <span>{mode === 'cinema' ? '影院账号由平台管理员创建' : '财务账号也从此入口登录'}</span>}</div></div><div className="auth-aside"><div className="auth-aside-grid" /><div className="auth-aside-content"><span className="eyebrow">LTERM CINEMA / 2026</span><h2>今天看什么，<br />由你决定。</h2><div className="auth-aside-rule" /><p>影院热映 · 全局搜索 · 实时选座</p></div></div></div>
}
