import { memo, useMemo } from 'react'
import { Layer, Line } from 'react-konva'

export type GuideLayerProps = {
  width: number
  height: number
  gap?: number
}

export const GuideLayer = memo(function GuideLayerComponent({ width, height, gap = 30 }: GuideLayerProps) {
  const lines = useMemo(() => {
    const vertical: number[][] = []
    const horizontal: number[][] = []
    for (let x = 0; x <= width; x += gap) {
      vertical.push([x, 0, x, height])
    }
    for (let y = 0; y <= height; y += gap) {
      horizontal.push([0, y, width, y])
    }
    return { vertical, horizontal }
  }, [width, height, gap])

  return (
    <Layer listening={false}>
      {lines.vertical.map((points, index) => (
        <Line key={`v-${index}`} points={points} stroke="rgba(255,255,255,0.03)" strokeWidth={1} />
      ))}
      {lines.horizontal.map((points, index) => (
        <Line key={`h-${index}`} points={points} stroke="rgba(255,255,255,0.03)" strokeWidth={1} />
      ))}
    </Layer>
  )
})
