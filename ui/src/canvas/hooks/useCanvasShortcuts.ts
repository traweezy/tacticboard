import { useEffect } from 'react'
import { useRoomStore } from '@state/store'

export const useCanvasShortcuts = () => {
  const toggleSnap = useRoomStore((state) => state.toggleSnap)
  const setTool = useRoomStore((state) => state.setTool)

  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      if (event.defaultPrevented) return
      switch (event.key.toLowerCase()) {
        case 'v':
          setTool('select')
          break
        case 'p':
          setTool('player')
          break
        case 'z':
          if (event.ctrlKey || event.metaKey) {
            event.preventDefault()
            // undo placeholder
          }
          break
        case 'g':
          if (event.shiftKey) {
            toggleSnap()
          }
          break
        default:
          break
      }
    }

    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [setTool, toggleSnap])
}
