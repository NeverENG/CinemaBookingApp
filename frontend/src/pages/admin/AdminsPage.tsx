import { KeyRound, Plus, ShieldCheck, UsersRound } from 'lucide-react'
import { useState, type FormEvent } from 'react'
import { Button, Field, Input, Modal, PageHeader, Select } from '../../components/ui'
import { useCreateAdminMutation } from '../../features/admin/hooks'
import { AppError } from '../../services/http/errors'

export function AdminsPage() {
  const [open, setOpen] = useState(false)
  const [message, setMessage] = useState('')
  return <div className="admin-page"><PageHeader eyebrow="SYSTEM" title="管理员账号" description="平台管理员创建影院运营与财务分析账号。" actions={<Button onClick={() => { setMessage(''); setOpen(true) }}><Plus size={15} />创建账号</Button>} /><div className="admin-info-grid"><div className="admin-info-card"><UsersRound size={20} /><strong>角色隔离</strong><span>每个管理员只通过角色进入对应工作空间。</span></div><div className="admin-info-card"><ShieldCheck size={20} /><strong>影院范围</strong><span>影院运营账号的数据范围由服务端 JWT 注入。</span></div><div className="admin-info-card"><KeyRound size={20} /><strong>敏感操作</strong><span>改价、发券、取消场次会写入操作日志。</span></div></div>{message && <div className="form-success">{message}</div>}<div className="admin-notice">当前后端已提供创建管理员接口，尚未提供管理员列表查询接口；第一版先把创建流程做成明确的操作抽屉。</div><AdminModal open={open} onClose={() => setOpen(false)} onSuccess={setMessage} /></div>
}

function AdminModal({ open, onClose, onSuccess }: { open: boolean; onClose: () => void; onSuccess: (message: string) => void }) {
  const create = useCreateAdminMutation()
  const [form, setForm] = useState({ username: '', password: '', nickname: '', role: 'CINEMA_ADMIN', cinema_id: '1' })
  async function submit(event: FormEvent) {
    event.preventDefault()
    try {
      await create.mutateAsync({ ...form, cinema_id: form.role === 'CINEMA_ADMIN' ? Number(form.cinema_id) : undefined })
      onSuccess('管理员账号创建成功')
      onClose()
    } catch (cause) {
      window.alert(cause instanceof AppError ? cause.message : '账号创建失败')
    }
  }
  return <Modal open={open} title="创建管理员账号" onClose={onClose}><form className="modal-form" onSubmit={submit}><Field label="登录账号"><Input value={form.username} onChange={(event) => setForm({ ...form, username: event.target.value })} required /></Field><Field label="显示昵称"><Input value={form.nickname} onChange={(event) => setForm({ ...form, nickname: event.target.value })} required /></Field><Field label="初始密码"><Input type="password" minLength={6} value={form.password} onChange={(event) => setForm({ ...form, password: event.target.value })} required /></Field><Field label="角色"><Select value={form.role} onChange={(event) => setForm({ ...form, role: event.target.value })}><option value="CINEMA_ADMIN">影院运营</option><option value="FINANCE">财务分析</option></Select></Field>{form.role === 'CINEMA_ADMIN' && <Field label="绑定影院"><Select value={form.cinema_id} onChange={(event) => setForm({ ...form, cinema_id: event.target.value })}><option value="1">LTerm 光影中心</option><option value="2">LTerm 北岸影院</option></Select></Field>}<div className="modal-form-actions"><Button type="button" variant="secondary" onClick={onClose}>取消</Button><Button type="submit" disabled={create.isPending}>{create.isPending ? '创建中…' : '创建账号'}</Button></div></form></Modal>
}
