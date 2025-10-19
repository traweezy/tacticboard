import { useCallback, useRef } from 'react'

import type { Operation } from './store'

export const useOpBuffer = (flush: (ops: Operation[]) => void, interval = 200) => {
  const bufferRef = useRef<Operation[]>([])
  const timer = useRef<number | null>(null)

  const schedule = useCallback(() => {
    if (timer.current !== null) {
      return
    }
    timer.current = window.setTimeout(() => {
      const ops = bufferRef.current
      bufferRef.current = []
      timer.current = null
      if (ops.length > 0) {
        flush(ops)
      }
    }, interval)
  }, [flush, interval])

  const push = useCallback(
    (op: Operation) => {
      bufferRef.current = [...bufferRef.current, op]
      schedule()
    },
    [schedule]
  )

  const flushImmediate = useCallback(() => {
    const ops = bufferRef.current
    bufferRef.current = []
    if (timer.current !== null) {
      window.clearTimeout(timer.current)
      timer.current = null
    }
    if (ops.length > 0) {
      flush(ops)
    }
  }, [flush])

  return { push, flush: flushImmediate }
}
