import { ArrowRight, CheckCircle2, Eye, MailCheck } from 'lucide-react'
import { useEffect, useState, type FormEvent } from 'react'
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
  const [code, setCode] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')
  const [sendingCode, setSendingCode] = useState(false)
  const [cooldown, setCooldown] = useState(0)
  const { run, isLocked } = useAsyncLock()

  useEffect(() => {
    if (cooldown <= 0) return undefined
    const timer = window.setTimeout(() => setCooldown((current) => current - 1), 1000)
    return () => window.clearTimeout(timer)
  }, [cooldown])

  async function requestCode() {
    setError('')
    setMessage('')
    if (!email) {
      setError('请先填写邮箱')
      return
    }
    setSendingCode(true)
    try {
      const result = await authApi.requestRegistrationCode(email)
      setCode(result.dev_code ?? '')
      setMessage('验证码已发送，15 分钟内有效')
      setCooldown(60)
    } catch (cause) {
      setError(cause instanceof AppError ? cause.message : '验证码发送失败')
    } finally {
      setSendingCode(false)
    }
  }

  async function submit(event: FormEvent) {
    event.preventDefault()
    setError('')
    if (password !== confirmPassword) {
      setError('两次输入的密码不一致')
      return
    }
    await run(async () => {
      try {
        const session = await authApi.register({ email, code, password, nickname })
        signIn(session)
        navigate('/')
      } catch (cause) {
        setError(cause instanceof AppError ? cause.message : '注册失败，请稍后重试')
      }
    })
  }

  return <div className="auth-page auth-page-simple"><div className="auth-panel"><Link to="/" className="brand"><span className="brand-mark"><Eye size={17} /></span><span>LTerm</span></Link><div className="auth-copy"><span className="eyebrow">CREATE ACCOUNT</span><h1>创建你的<br /><em>观影记录。</em></h1><p>验证邮箱后即可选座、购票并积累会员积分。</p></div><form className="auth-form" onSubmit={submit}><Field label="昵称"><Input value={nickname} onChange={(event) => setNickname(event.target.value)} placeholder="怎么称呼你" required /></Field><Field label="邮箱"><Input value={email} onChange={(event) => { setEmail(event.target.value); setCode(''); setMessage('') }} type="email" placeholder="name@example.com" autoComplete="email" required /></Field><div className="auth-code-row"><Field label="邮箱验证码"><Input value={code} onChange={(event) => setCode(event.target.value.replace(/\D/g, '').slice(0, 6))} inputMode="numeric" autoComplete="one-time-code" placeholder="6 位验证码" minLength={6} required /></Field><Button type="button" variant="secondary" onClick={requestCode} disabled={sendingCode || cooldown > 0}>{sendingCode ? '发送中…' : cooldown > 0 ? `${cooldown}s` : '获取验证码'}</Button></div><Field label="密码"><Input value={password} onChange={(event) => setPassword(event.target.value)} type="password" placeholder="至少 6 位字符" minLength={6} autoComplete="new-password" required /></Field><Field label="确认密码"><Input value={confirmPassword} onChange={(event) => setConfirmPassword(event.target.value)} type="password" placeholder="再次输入密码" minLength={6} autoComplete="new-password" required /></Field>{message && <div className="form-success"><MailCheck size={14} />{message}</div>}{error && <div className="form-error">{error}</div>}<Button type="submit" size="lg" disabled={isLocked}>{isLocked ? '创建中…' : '验证并创建账号'}<ArrowRight size={16} /></Button></form><div className="auth-foot"><span>已经有账号？<Link to="/login">返回登录</Link></span><span className="auth-verified-note"><CheckCircle2 size={13} />邮箱验证</span></div></div><div className="auth-aside auth-aside-light"><div className="auth-aside-content"><span className="eyebrow">YOUR NEXT SCREENING</span><h2>每一场电影，<br />都值得被记住。</h2><div className="auth-aside-rule" /><p>记录你的座位、片单和下一次期待。</p></div></div></div>
}
