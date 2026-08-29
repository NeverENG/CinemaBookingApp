import { ArrowLeft, ShieldX } from 'lucide-react'
import { Link } from 'react-router-dom'

export function ForbiddenPage() {
  return <div className="feedback-page"><ShieldX size={34} /><span className="eyebrow">403 / FORBIDDEN</span><h1>这个页面暂时不对你开放</h1><p>你的角色没有访问该工作空间的权限。</p><div><Link className="button button-primary" to="/">返回热映</Link><Link className="back-link" to="/admin/dashboard"><ArrowLeft size={15} />返回工作台</Link></div></div>
}

export function NotFoundPage() {
  return <div className="feedback-page"><span className="eyebrow">404 / NOT FOUND</span><h1>页面走丢了</h1><p>回到热映列表，继续寻找下一场电影。</p><Link className="button button-primary" to="/">返回热映</Link></div>
}
