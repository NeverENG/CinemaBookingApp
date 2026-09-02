import { Building2, Plus, Search, UsersRound } from 'lucide-react'
import { useMemo, useState, type FormEvent } from 'react'
import { roleLabel } from '../../auth'
import { Avatar, Button, DemoBadge, Field, Input, LoadingBlock, Modal, PageHeader, Select, StatusBadge } from '../../components/ui'
import { useAdminAccountsQuery, useCreateAdminMutation } from '../../features/admin/hooks'
import { useCinemaQuery } from '../../features/catalog/hooks'
import { AppError } from '../../services/http/errors'
import type { AdminAccount } from '../../types'

export function AdminsPage() {
  const query = useAdminAccountsQuery()
  const [open, setOpen] = useState(false)
  const [message, setMessage] = useState('')
  const [keyword, setKeyword] = useState('')
  const accounts = useMemo(() => {
    const normalized = keyword.trim().toLowerCase()
    if (!normalized) return query.data ?? []
    return (query.data ?? []).filter((account) => `${account.nickname} ${account.username} ${roleLabel(account.role)} ${account.cinemaName ?? ''}`.toLowerCase().includes(normalized))
  }, [keyword, query.data])

  return <div className="admin-page">
    <PageHeader eyebrow="SYSTEM" title="管理员账号" description="查看账号归属，并创建影院运营与财务分析账号。" actions={<div className="dashboard-actions">{query.isDemo && <DemoBadge />}<Button onClick={() => { setMessage(''); setOpen(true) }}><Plus size={15} />创建账号</Button></div>} />
    {message && <div className="form-success">{message}</div>}
    {query.error && <div className="form-error">账号列表加载失败：{query.error.message}</div>}
    <div className="admin-toolbar"><div className="toolbar-search"><Search size={16} /><input value={keyword} onChange={(event) => setKeyword(event.target.value)} placeholder="搜索账号、姓名、角色或影院" /></div><span className="toolbar-count">{accounts.length} 个账号</span></div>
    <div className="admin-table-card"><table className="admin-table admin-account-table"><thead><tr><th>账号名称</th><th>角色</th><th>所属影院</th><th>状态</th><th>创建时间</th></tr></thead><tbody>{query.isPending ? <tr><td colSpan={5}><LoadingBlock lines={3} /></td></tr> : <>{accounts.map((account) => <AdminAccountRow key={account.id} account={account} />)}{accounts.length === 0 && <tr><td colSpan={5}><div className="admin-table-empty"><UsersRound size={22} /><strong>{keyword ? '没有匹配的账号' : '暂无管理员账号'}</strong><span>{keyword ? '请调整搜索关键词。' : '创建影院运营或财务账号后会显示在这里。'}</span></div></td></tr>}</>}</tbody></table></div>
    <AdminModal open={open} onClose={() => setOpen(false)} onSuccess={setMessage} />
  </div>
}

function AdminAccountRow({ account }: { account: AdminAccount }) {
  const cinemaLabel = account.cinemaName || (account.role === 'FINANCE' ? '全部影院' : '平台全局')
  const createdAt = account.createdAt ? new Date(account.createdAt).toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }) : '--'
  return <tr className="admin-account-row"><td><div className="admin-account"><Avatar name={account.nickname || account.username} size="sm" /><div><strong>{account.nickname || account.username}</strong><span>@{account.username}</span></div></div></td><td data-label="角色"><span className={`admin-role-label role-${account.role.toLowerCase().replaceAll('_', '-')}`}>{roleLabel(account.role)}</span></td><td data-label="所属影院"><div className="admin-cinema-cell"><Building2 size={15} /><div><strong>{cinemaLabel}</strong>{account.cinemaId && <span>影院 ID {account.cinemaId}</span>}</div></div></td><td data-label="状态"><StatusBadge status={account.status} /></td><td data-label="创建时间">{createdAt}</td></tr>
}

function AdminModal({ open, onClose, onSuccess }: { open: boolean; onClose: () => void; onSuccess: (message: string) => void }) {
  const create = useCreateAdminMutation()
  const cinemasQuery = useCinemaQuery()
  const cinemas = cinemasQuery.data ?? []
  const [form, setForm] = useState({ username: '', password: '', nickname: '', role: 'CINEMA_ADMIN', cinema_id: '' })
  const [error, setError] = useState('')
  const selectedCinemaID = form.cinema_id || String(cinemas[0]?.id ?? '')
  async function submit(event: FormEvent) {
    event.preventDefault()
    setError('')
    if (form.role === 'CINEMA_ADMIN' && !selectedCinemaID) {
      setError('当前没有可绑定的影院')
      return
    }
    try {
      await create.mutateAsync({ ...form, cinema_id: form.role === 'CINEMA_ADMIN' ? Number(selectedCinemaID) : undefined })
      onSuccess(`管理员账号“${form.nickname}”创建成功`)
      setForm({ username: '', password: '', nickname: '', role: 'CINEMA_ADMIN', cinema_id: '' })
      onClose()
    } catch (cause) {
      setError(cause instanceof AppError ? cause.message : '账号创建失败')
    }
  }
  return <Modal open={open} title="创建管理员账号" onClose={onClose}><form className="modal-form" onSubmit={submit}><Field label="账号名称"><Input value={form.nickname} onChange={(event) => setForm({ ...form, nickname: event.target.value })} placeholder="例如：万象城影院运营" required /></Field><Field label="登录账号"><Input value={form.username} onChange={(event) => setForm({ ...form, username: event.target.value })} placeholder="例如：mixc_ops" autoComplete="off" required /></Field><Field label="初始密码"><Input type="password" minLength={6} value={form.password} onChange={(event) => setForm({ ...form, password: event.target.value })} autoComplete="new-password" required /></Field><Field label="角色"><Select value={form.role} onChange={(event) => setForm({ ...form, role: event.target.value })}><option value="CINEMA_ADMIN">影院运营</option><option value="FINANCE">财务分析</option></Select></Field>{form.role === 'CINEMA_ADMIN' && <Field label="所属影院" hint="账号登录后只能管理这家影院"><Select value={selectedCinemaID} onChange={(event) => setForm({ ...form, cinema_id: event.target.value })} disabled={cinemas.length === 0}>{cinemas.length === 0 ? <option value="">暂无可用影院</option> : cinemas.map((cinema) => <option key={cinema.id} value={cinema.id}>{cinema.name} · {cinema.city}</option>)}</Select></Field>}{error && <div className="form-error">{error}</div>}<div className="modal-form-actions"><Button type="button" variant="secondary" onClick={onClose}>取消</Button><Button type="submit" disabled={create.isPending || (form.role === 'CINEMA_ADMIN' && cinemas.length === 0)}>{create.isPending ? '创建中…' : '创建账号'}</Button></div></form></Modal>
}
