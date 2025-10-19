import { renderHook, act } from '@testing-library/react'
import { beforeAll, describe, expect, it } from 'vitest'

import { useStageSize } from './useStageSize'

describe('useStageSize', () => {
  const callbacks: ((entries: ResizeObserverEntry[]) => void)[] = []

  beforeAll(() => {
    class MockResizeObserver {
      private readonly callback: (entries: ResizeObserverEntry[]) => void
      constructor(callback: (entries: ResizeObserverEntry[]) => void) {
        this.callback = callback
        callbacks.push(callback)
      }
      observe(): void {
        // no-op for test shim
        void this.callback
      }
      disconnect(): void {
        void this.callback
      }
    }
    // @ts-expect-error - test shim
    global.ResizeObserver = MockResizeObserver
  })

  it('tracks container size', () => {
    const { result } = renderHook(() => useStageSize())
    const div = document.createElement('div')

    act(() => {
      result.current.containerRef(div)
    })

    act(() => {
      callbacks[0]([
        {
          contentRect: { width: 200, height: 150 }
        } as ResizeObserverEntry
      ])
    })

    expect(result.current.width).toBe(200)
    expect(result.current.height).toBe(150)
  })
})
