import type { ButtonHTMLAttributes, InputHTMLAttributes, PropsWithChildren, ReactNode, SelectHTMLAttributes } from 'react'
import type { LucideIcon } from 'lucide-react'
import { X } from 'lucide-react'
import { statusText } from '../lib/format'

type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger' | 'quiet'
type ButtonSize = 'sm' | 'md' | 'lg'

export function Button({ variant = 'primary', size = 'md', className = '', children, ...props }: ButtonHTMLAttributes<HTMLButtonElement> & { variant?: ButtonVariant; size?: ButtonSize }) {
  return <button className={`button button-${variant} button-${size} ${className}`} {...props}>{children}</button>
}

export function IconButton({ label, icon: Icon, className = '', ...props }: ButtonHTMLAttributes<HTMLButtonElement> & { label: string; icon: LucideIcon }) {
  return <button className={`icon-button ${className}`} aria-label={label} title={label} {...props}><Icon size={17} strokeWidth={1.8} /></button>
}

export function StatusBadge({ status }: { status: string }) {
  const key = status.toLowerCase().replaceAll('_', '-')
  return <span className={`status-badge status-${key}`}>{statusText(status)}</span>
}

export function DemoBadge() {
  return <span className="demo-badge">演示数据</span>
}

export function Panel({ children, className = '', title, action }: PropsWithChildren<{ className?: string; title?: string; action?: ReactNode }>) {
  return <section className={`panel ${className}`}>
    {(title || action) && <div className="panel-heading"><div>{title && <h2>{title}</h2>}</div>{action}</div>}
    {children}
  </section>
}

export function EmptyState({ icon: Icon, title, description, action }: { icon?: LucideIcon; title: string; description?: string; action?: ReactNode }) {
  return <div className="empty-state">{Icon && <div className="empty-icon"><Icon size={22} /></div>}<strong>{title}</strong>{description && <p>{description}</p>}{action}</div>
}

export function LoadingBlock({ lines = 3 }: { lines?: number }) {
  return <div className="loading-block" aria-label="加载中">{Array.from({ length: lines }, (_, index) => <span key={index} />)}</div>
}

export function Field({ label, hint, children }: { label: string; hint?: string; children: ReactNode }) {
  return <label className="field"><span className="field-label">{label}</span>{children}{hint && <small>{hint}</small>}</label>
}

export function Input(props: InputHTMLAttributes<HTMLInputElement>) {
  return <input className="input" {...props} />
}

export function Select(props: SelectHTMLAttributes<HTMLSelectElement>) {
  return <select className="input select" {...props} />
}

export function TextArea(props: React.TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return <textarea className="input textarea" {...props} />
}

export function Modal({ open, title, onClose, children, footer }: { open: boolean; title: string; onClose: () => void; children: ReactNode; footer?: ReactNode }) {
  if (!open) return null
  return <div className="modal-backdrop" role="presentation" onMouseDown={onClose}>
    <div className="modal" onMouseDown={(event) => event.stopPropagation()}>
      <div className="modal-header"><h2>{title}</h2><IconButton label="关闭" icon={X} onClick={onClose} type="button" /></div>
      <div className="modal-body">{children}</div>
      {footer && <div className="modal-footer">{footer}</div>}
    </div>
  </div>
}

export function MetricCard({ label, value, note, icon: Icon, tone = 'default' }: { label: string; value: string; note?: string; icon?: LucideIcon; tone?: 'default' | 'gold' | 'green' | 'red' }) {
  return <div className={`metric-card metric-${tone}`}><div className="metric-top"><span>{label}</span>{Icon && <Icon size={17} />}</div><strong>{value}</strong>{note && <small>{note}</small>}</div>
}

export function PageHeader({ eyebrow, title, description, actions }: { eyebrow?: string; title: string; description?: string; actions?: ReactNode }) {
  return <div className="page-header"><div>{eyebrow && <span className="eyebrow">{eyebrow}</span>}<h1>{title}</h1>{description && <p>{description}</p>}</div>{actions && <div className="page-actions">{actions}</div>}</div>
}

export function Avatar({ name = 'L', size = 'md' }: { name?: string; size?: 'sm' | 'md' | 'lg' }) {
  const label = name.trim().slice(0, 1).toUpperCase() || 'L'
  return <span className={`avatar avatar-${size}`}>{label}</span>
}
