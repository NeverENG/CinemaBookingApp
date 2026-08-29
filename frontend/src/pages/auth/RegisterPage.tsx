import { ArrowRight, Eye } from 'lucide-react'
import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../../app/providers'
import { Button, Field, Input } from '../../components/ui'
import { useAsyncLock } from '../../hooks/useAsyncLock'
import { authApi } from '../../services/api'
import { AppError } from '../../services/http/errors'

export function RegisterPage() {
  const navigate = useNavigate()
  const { signIn } = useAuth()
  const [email, setEmail] = useState('')
  const [nickname, setNickname] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const { run, isLocked } = useAsyncLock()

  async function submit(event: FormEvent) {
    event.preventDefault()
    setError('')
    await run(async () => {
      try {
        const session = await authApi.register({ email, password, nickname })
        signIn(session)
        navigate('/')
      } catch (cause) {
        setError(cause instanceof AppError ? cause.message : '注册失败，请稍后重试')
      }
    })
  }

  return <div className="auth-page auth-page-simple"><div className="auth-panel"><Link to="/" className="brand"><span className="brand-mark"><Eye size={17} /></span><span>LTerm</span></Link><div className="auth-copy"><span className="eyebrow">CREATE ACCOUNT</span><h1>创建你的<br /><em>观影记录。</em></h1><p>注册后即可选座、购票并积累会员积分。</p></div><form className="auth-form" onSubmit={submit}><Field label="昵称"><Input value={nickname} onChange={(event) => setNickname(event.target.value)} placeholder="怎么称呼你" required /></Field><Field label="邮箱"><Input value={email} onChange={(event) => setEmail(event.target.value)} type="email" placeholder="name@example.com" required /></Field><Field label="密码"><Input value={password} onChange={(event) => setPassword(event.target.value)} type="password" placeholder="至少 6 位字符" minLength={6} required /></Field>{error && <div className="form-error">{error}</div>}<Button type="submit" size="lg" disabled={isLocked}>{isLocked ? '创建中…' : '创建账号'}<ArrowRight size={16} /></Button></form><div className="auth-foot"><span>已经有账号？<Link to="/login">返回登录</Link></span></div></div><div className="auth-aside auth-aside-light"><div className="auth-aside-content"><span className="eyebrow">YOUR NEXT SCREENING</span><h2>每一场电影，<br />都值得被记住。</h2><div className="auth-aside-rule" /><p>记录你的座位、片单和下一次期待。</p></div></div></div>
}
