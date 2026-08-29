import { useCallback, useRef, useState } from 'react'

export function useAsyncLock() {
  const locked = useRef(false)
  const [isLocked, setLocked] = useState(false)

  const run = useCallback(async <T,>(task: () => Promise<T>) => {
    if (locked.current) return undefined
    locked.current = true
    setLocked(true)
    try {
      return await task()
    } finally {
      locked.current = false
      setLocked(false)
    }
  }, [])

  return { run, isLocked }
}
