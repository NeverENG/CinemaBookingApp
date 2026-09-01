import { Edit3, Plus, Search, ToggleLeft, ToggleRight } from 'lucide-react'
import { useMemo, useState, type FormEvent } from 'react'
import { Button, DemoBadge, Field, Input, Modal, PageHeader, StatusBadge, TextArea } from '../../components/ui'
import { useAdminMoviesQuery, useCreateAdminMovieMutation, useSetMovieStatusMutation, useUpdateAdminMovieMutation } from '../../features/admin/hooks'
import type { Movie } from '../../types'

interface MovieForm {
  title: string
  cover_url: string
  trailer_url: string
  description: string
  duration_minutes: string
  genre: string
  release_date: string
  rating: string
}

const emptyMovieForm: MovieForm = { title: '', cover_url: '', trailer_url: '', description: '', duration_minutes: '120', genre: '', release_date: '', rating: '8.0' }

function movieToForm(movie: Movie): MovieForm {
  return {
    title: movie.title,
    cover_url: movie.coverUrl,
    trailer_url: movie.trailerUrl,
    description: movie.description,
    duration_minutes: String(movie.durationMinutes || 120),
    genre: movie.genre,
    release_date: movie.releaseDate?.slice(0, 10) || '',
    rating: String(movie.rating || 0),
  }
}

export function MoviesPage() {
  const query = useAdminMoviesQuery()
  const create = useCreateAdminMovieMutation()
  const update = useUpdateAdminMovieMutation()
  const status = useSetMovieStatusMutation()
  const [keyword, setKeyword] = useState('')
  const [open, setOpen] = useState(false)
  const [editingMovie, setEditingMovie] = useState<Movie | null>(null)
  const [form, setForm] = useState<MovieForm>(emptyMovieForm)
  const movies = useMemo(() => (query.data ?? []).filter((movie) => `${movie.title}${movie.genre}`.toLowerCase().includes(keyword.toLowerCase())), [keyword, query.data])

  function openCreate() {
    setEditingMovie(null)
    setForm(emptyMovieForm)
    setOpen(true)
  }

  function openEdit(movie: Movie) {
    setEditingMovie(movie)
    setForm(movieToForm(movie))
    setOpen(true)
  }

  async function submit(event: FormEvent) {
    event.preventDefault()
    const data = { ...form, duration_minutes: Number(form.duration_minutes), rating: Number(form.rating) }
    if (editingMovie) await update.mutateAsync({ id: editingMovie.id, data })
    else await create.mutateAsync(data)
    setForm(emptyMovieForm)
    setEditingMovie(null)
    setOpen(false)
  }

  const saving = create.isPending || update.isPending
  return <div className="admin-page"><PageHeader eyebrow="CONTENT" title="影片管理" description="维护影片信息，并控制用户端可见状态。" actions={<div className="dashboard-actions">{query.isDemo && <DemoBadge />}<Button onClick={openCreate}><Plus size={15} />新增影片</Button></div>} /><div className="admin-toolbar"><div className="toolbar-search"><Search size={16} /><input value={keyword} onChange={(event) => setKeyword(event.target.value)} placeholder="搜索片名或类型" /></div><span className="toolbar-count">{movies.length} 部影片</span></div><div className="admin-table-card"><table className="admin-table"><thead><tr><th>影片</th><th>类型</th><th>上映日期</th><th>评分</th><th>状态</th><th>操作</th></tr></thead><tbody>{movies.map((movie) => <MovieRow key={movie.id} movie={movie} onEdit={() => openEdit(movie)} onToggle={() => status.mutate({ id: movie.id, status: movie.status === 'ON_SALE' ? 'OFF_SALE' : 'ON_SALE' })} />)}</tbody></table></div><Modal open={open} title={editingMovie ? '编辑影片' : '新增影片'} onClose={() => setOpen(false)}><form className="modal-form" onSubmit={submit}><div className="form-grid"><Field label="影片名称"><Input value={form.title} onChange={(event) => setForm({ ...form, title: event.target.value })} required /></Field><Field label="类型"><Input value={form.genre} onChange={(event) => setForm({ ...form, genre: event.target.value })} placeholder="科幻 · 剧情" /></Field><Field label="时长（分钟）"><Input type="number" value={form.duration_minutes} onChange={(event) => setForm({ ...form, duration_minutes: event.target.value })} min={1} required /></Field><Field label="评分"><Input type="number" value={form.rating} onChange={(event) => setForm({ ...form, rating: event.target.value })} min={0} max={10} step={0.1} /></Field><Field label="上映日期"><Input type="date" value={form.release_date} onChange={(event) => setForm({ ...form, release_date: event.target.value })} /></Field><Field label="海报 URL"><Input value={form.cover_url} onChange={(event) => setForm({ ...form, cover_url: event.target.value })} placeholder="https://" /></Field></div><Field label="简介"><TextArea value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} rows={4} /></Field><Field label="预告片 URL"><Input value={form.trailer_url} onChange={(event) => setForm({ ...form, trailer_url: event.target.value })} placeholder="可选" /></Field><div className="modal-form-actions"><Button type="button" variant="secondary" onClick={() => setOpen(false)}>取消</Button><Button type="submit" disabled={saving}>{saving ? '保存中…' : editingMovie ? '保存修改' : '保存影片'}</Button></div></form></Modal></div>
}

function MovieRow({ movie, onEdit, onToggle }: { movie: Movie; onEdit: () => void; onToggle: () => void }) {
  return <tr><td><div className="table-movie"><div className="table-poster">{movie.coverUrl && <img src={movie.coverUrl} alt="" />}</div><div><strong>{movie.title}</strong><span>{movie.durationMinutes} 分钟</span></div></div></td><td>{movie.genre || '--'}</td><td>{movie.releaseDate?.slice(0, 10) || '--'}</td><td><span className="rating-inline">★ {movie.rating.toFixed(1)}</span></td><td><StatusBadge status={movie.status || 'ON_SALE'} /></td><td><div className="table-actions"><button onClick={onToggle} title="切换状态">{movie.status === 'ON_SALE' ? <ToggleRight size={18} /> : <ToggleLeft size={18} />}</button><button onClick={onEdit} title="编辑"><Edit3 size={16} /></button></div></td></tr>
}
