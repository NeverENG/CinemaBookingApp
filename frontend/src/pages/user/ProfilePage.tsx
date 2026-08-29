import { Check, KeyRound, UserRound } from 'lucide-react'
import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { roleLabel } from '../../auth'
import { useAuth } from '../../app/providers'
import { Avatar, Button, Field, Input, PageHeader } from '../../components/ui'
import { useAsyncLock } from '../../hooks/useAsyncLock'
import { authApi } from '../../services/api'
import { AppError } from '../../services/http/errors'

export function ProfilePage() {
  const { session } = useAuth()
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const { run, isLocked } = useAsyncLock()

  async function submit(event: FormEvent) {
    event.preventDefault()
    setMessage('')
    setError('')
    await run(async () => {
      try {
        await authApi.changePassword({ old_password: oldPassword, new_password: newPassword })
        setMessage('密码已修改，请重新登录。')
        setOldPassword('')
        setNewPassword('')
      } catch (cause) {
        setError(cause instanceof AppError ? cause.message : '密码修改失败')
      }
    })
  }

  return <div className="content-container narrow-container"><PageHeader eyebrow="PROFILE" title="账户设置" description="管理登录安全和你的观影偏好。" /><div className="profile-grid"><section className="profile-card"><Avatar name={roleLabel(session?.role)} size="lg" /><div><strong>{roleLabel(session?.role)}</strong><span>{session?.role === 'USER' ? '观众账户' : '管理工作台账户'}</span></div><span className="profile-id">ID {session?.userId ?? '--'}</span></section><section className="profile-card profile-card-column"><div className="section-label"><KeyRound size={15} />修改密码</div><form className="profile-form" onSubmit={submit}><Field label="当前密码"><Input value={oldPassword} onChange={(event) => setOldPassword(event.target.value)} type="password" required /></Field><Field label="新密码"><Input value={newPassword} onChange={(event) => setNewPassword(event.target.value)} type="password" minLength={6} required /></Field>{message && <div className="form-success"><Check size={14} />{message}</div>}{error && <div className="form-error">{error}</div>}<Button type="submit" disabled={isLocked}>{isLocked ? '提交中…' : '保存密码'}</Button></form></section></div><div className="profile-note"><UserRound size={16} /><span>头像使用系统生成的颜色头像，不上传图片；管理员账户的角色和影院范围由平台配置。</span><Link className="text-link" to="/">返回热映</Link></div></div>
}
