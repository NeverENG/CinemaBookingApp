import { BadgeCheck, ScanLine, TicketCheck } from 'lucide-react'
import { useState, type FormEvent } from 'react'
import { Button, EmptyState, Field, Input, PageHeader, StatusBadge } from '../../components/ui'
import { useVerifyTicketMutation } from '../../features/admin/hooks'
import { dateText, timeText } from '../../lib/format'
import { AppError } from '../../services/http/errors'
import type { TicketVerification } from '../../types'

export function TicketsPage() {
  const [ticketNo, setTicketNo] = useState('')
  const [result, setResult] = useState<TicketVerification | null>(null)
  const verify = useVerifyTicketMutation()

  async function submit(event: FormEvent) {
    event.preventDefault()
    const value = ticketNo.trim()
    if (!value) return
    try {
      setResult(await verify.mutateAsync(value))
    } catch {
      setResult(null)
    }
  }

  const error = verify.error instanceof AppError ? verify.error.message : verify.error ? '核销失败，请稍后重试' : ''

  return <div className="admin-page">
    <PageHeader eyebrow="TICKET SERVICE" title="票券核销" description="验证支付成功的取票码，完成入场核销。" />
    <div className="ticket-verify-layout">
      <section className="ticket-verify-panel">
        <div className="ticket-verify-icon"><ScanLine size={24} /></div>
        <h2>核销取票码</h2>
        <p>输入用户订单详情中的票券码，每张票独立核销。</p>
        <form className="ticket-verify-form" onSubmit={submit}>
          <Field label="票券码"><Input autoFocus value={ticketNo} onChange={(event) => { setTicketNo(event.target.value); setResult(null) }} placeholder="例如 TKSEED..." autoComplete="off" /></Field>
          <Button type="submit" disabled={!ticketNo.trim() || verify.isPending}><TicketCheck size={16} />{verify.isPending ? '核销中…' : '确认核销'}</Button>
        </form>
        {error && <div className="form-error">{error}</div>}
      </section>
      <section className="ticket-verify-aside">
        <div className="ticket-verify-aside-icon"><BadgeCheck size={19} /></div>
        <strong>核销规则</strong>
        <span>仅支付成功且未退款的订单可以入场。</span>
        <span>同一票券重复提交会返回原核销结果，不会重复扣减。</span>
        <span>多张票订单支持逐张核销，全部完成后订单变为已完成。</span>
      </section>
    </div>
    {result ? <VerificationResult result={result} /> : <div className="ticket-verify-empty"><EmptyState icon={TicketCheck} title="等待核销" description="核销成功后，票券和订单状态会显示在这里。" /></div>}
  </div>
}

function VerificationResult({ result }: { result: TicketVerification }) {
  return <section className="ticket-verify-result">
    <div className="ticket-verify-result-heading"><div className="ticket-verify-result-icon"><BadgeCheck size={21} /></div><div><strong>{result.alreadyUsed ? '票券已核销' : '核销成功'}</strong><span>{result.alreadyUsed ? '本次返回历史核销结果' : '该票券已记录入场核销时间'}</span></div><StatusBadge status={result.orderStatus} /></div>
    <div className="ticket-verify-result-grid"><div><small>票券码</small><strong>{result.ticketNo}</strong></div><div><small>座位</small><strong>{result.seatNo}</strong></div><div><small>订单号</small><strong>{result.orderNo}</strong></div><div><small>核销时间</small><strong>{dateText(result.usedAt)} {timeText(result.usedAt)}</strong></div></div>
  </section>
}
