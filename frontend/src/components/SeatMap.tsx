import type { Seat } from '../types'

export function SeatMap({ seats, selectedIds, onToggle }: { seats: Seat[]; selectedIds: number[]; onToggle: (seat: Seat) => void }) {
  const rows = Array.from(new Set(seats.map((seat) => seat.rowNo))).sort((a, b) => a - b)
  return <div className="seat-map-wrap"><div className="screen-label">SCREEN</div><div className="screen-line" /><div className="seat-map">
    {rows.map((row) => <div className="seat-row" key={row}><span className="seat-row-label">{String.fromCharCode(64 + row)}</span>{seats.filter((seat) => seat.rowNo === row).sort((a, b) => a.colNo - b.colNo).map((seat) => {
      const selected = selectedIds.includes(seat.seatId)
      const disabled = seat.status !== 'available' && !selected
      return <button key={seat.seatId} className={`seat seat-${selected ? 'selected' : seat.status}`} disabled={disabled} title={`${seat.seatNo} · ${seat.type}`} onClick={() => onToggle(seat)}>{seat.colNo}</button>
    })}</div>)}
  </div></div>
}
