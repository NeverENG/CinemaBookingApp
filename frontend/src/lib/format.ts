export function money(cents = 0) {
  return `¥${(cents / 100).toFixed(2)}`
}

export function shortMoney(cents = 0) {
  return `¥${(cents / 100).toFixed(0)}`
}

export function dateText(value?: string | Date) {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return String(value).slice(0, 10)
  return new Intl.DateTimeFormat('zh-CN', { month: 'long', day: 'numeric', weekday: 'short' }).format(date)
}

export function timeText(value?: string | Date) {
  if (!value) return '--:--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return String(value).slice(11, 16)
  return new Intl.DateTimeFormat('zh-CN', { hour: '2-digit', minute: '2-digit', hour12: false }).format(date)
}

export function statusText(status: string) {
  const key = status.toUpperCase()
  const map: Record<string, string> = {
    OPEN: '可购',
    SOLD_OUT: '售罄',
    CLOSED: '已结束',
    CANCELED: '已取消',
    PENDING_PAYMENT: '待支付',
    PAID: '已支付',
    COMPLETED: '已完成',
    REFUNDING: '退款中',
    REFUNDED: '已退款',
    EXPIRED: '已过期',
    ACTIVE: '启用中',
    INACTIVE: '已停用',
    ON_SALE: '上映中',
    OFF_SALE: '已下架',
    SUCCESS: '成功',
    PENDING: '处理中',
  }
  return map[key] ?? status
}
