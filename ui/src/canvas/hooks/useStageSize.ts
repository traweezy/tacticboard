import { useCallback, useEffect, useState } from 'react'

export const useStageSize = () => {
  const [element, setElement] = useState<HTMLDivElement | null>(null)
  const [size, setSize] = useState({ width: 0, height: 0 })

  useEffect(() => {
    if (!element) return
    const node = element
    const obs = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width, height } = entry.contentRect
        setSize({ width, height })
      }
    })

    obs.observe(node)
    setSize({ width: node.clientWidth, height: node.clientHeight })

    return () => obs.disconnect()
  }, [element])

  const ref = useCallback((node: HTMLDivElement | null) => {
    setElement(node)
  }, [])

  return { containerRef: ref, width: size.width, height: size.height }
}
