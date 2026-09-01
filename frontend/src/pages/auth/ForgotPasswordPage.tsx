import { ArrowLeft, ArrowRight, CheckCircle2, Eye, KeyRound } from 'lucide-react'
import { useEffect, useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Button, Field, Input } from '../../components/ui'
import { useAsyncLock } from '../../hooks/useAsyncLock'
import { authApi } from '../../services/api'
import { AppError } from '../../services/http/errors'

export function ForgotPasswordPage() {
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [code, setCode] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [requested, setRequested] = useState(false)
  const [sendingCode, setSendingCode] = useState(false)
  const [cooldown, setCooldown] = useState(0)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
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
      setError('请先填写注册邮箱')
      return
    }
    setSendingCode(true)
    try {
      const result = await authApi.requestPasswordReset(email)
      setCode(result.dev_code ?? '')
      setRequested(true)
      setCooldown(60)
      setMessage('验证码已发送，15 分钟内有效')
    } catch (cause) {
      setError(cause instanceof AppError ? cause.message : '验证码发送失败')
    } finally {
      setSendingCode(false)
    }
  }

  async function submit(event: FormEvent) {
    event.preventDefault()
    setError('')
    if (newPassword !== confirmPassword) {
      setError('两次输入的密码不一致')
      return
    }
    await run(async () => {
      try {
        await authApi.resetPassword({ email, code, new_password: newPassword })
        navigate('/login?mode=user&reset=1', { replace: true })
      } catch (cause) {
        setError(cause instanceof AppError ? cause.message : '密码重置失败')
      }
    })
  }

  return <div className="auth-page auth-page-simple"><div className="auth-panel"><Link to="/" className="brand"><span className="brand-mark"><Eye size={17} /></span><span>LTerm</span></Link><div className="auth-copy"><span className="eyebrow">RESET PASSWORD</span><h1>找回你的<br /><em>观影账户。</em></h1><p>验证码会发送到注册邮箱。</p></div><form className="auth-form" onSubmit={submit}><Field label="注册邮箱"><Input value={email} onChange={(event) => { setEmail(event.target.value); setRequested(false); setCode(''); setMessage('') }} type="email" autoComplete="email" placeholder="name@example.com" required /></Field><div className="auth-code-row"><Field label="邮箱验证码"><Input value={code} onChange={(event) => setCode(event.target.value.replace(/\D/g, '').slice(0, 6))} inputMode="numeric" autoComplete="one-time-code" placeholder="6 位验证码" minLength={6} required /></Field><Button type="button" variant="secondary" onClick={requestCode} disabled={sendingCode || cooldown > 0}>{sendingCode ? '发送中…' : cooldown > 0 ? `${cooldown}s` : requested ? '重新发送' : '获取验证码'}</Button></div><Field label="新密码"><Input value={newPassword} onChange={(event) => setNewPassword(event.target.value)} type="password" minLength={6} autoComplete="new-password" placeholder="至少 6 位字符" required /></Field><Field label="确认新密码"><Input value={confirmPassword} onChange={(event) => setConfirmPassword(event.target.value)} type="password" minLength={6} autoComplete="new-password" placeholder="再次输入新密码" required /></Field>{message && <div className="form-success"><CheckCircle2 size={14} />{message}</div>}{error && <div className="form-error">{error}</div>}<Button type="submit" size="lg" disabled={isLocked || !requested}>{isLocked ? '提交中…' : '重置密码'}<ArrowRight size={16} /></Button></form><div className="auth-foot"><Link className="back-link" to="/login?mode=user"><ArrowLeft size={14} />返回登录</Link></div></div><div className="auth-aside auth-aside-light"><div className="auth-aside-content"><KeyRound size={24} /><span className="eyebrow">ACCOUNT SECURITY</span><h2>确认邮箱，<br />重新设置密码。</h2><div className="auth-aside-rule" /><p>验证码仅可使用一次。</p></div></div></div>
}
