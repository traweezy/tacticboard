import { memo } from 'react'
import { Circle, Layer, Line, Rect } from 'react-konva'

export type FieldLayerProps = {
  width: number
  height: number
}

export const FieldLayer = memo(function FieldLayerComponent({ width, height }: FieldLayerProps) {
  const centerX = width / 2
  const centerY = height / 2
  const fieldMargin = 24
  return (
    <Layer listening={false}>
      <Rect
        x={fieldMargin}
        y={fieldMargin}
        width={width - fieldMargin * 2}
        height={height - fieldMargin * 2}
        cornerRadius={32}
        fill="#091019"
        stroke="rgba(255,255,255,0.08)"
        strokeWidth={4}
      />
      <Line
        points={[centerX, fieldMargin, centerX, height - fieldMargin]}
        stroke="rgba(255,255,255,0.12)"
        strokeWidth={2}
        dash={[16, 12]}
      />
      <Circle x={centerX} y={centerY} radius={80} stroke="rgba(255,255,255,0.12)" strokeWidth={2} />
    </Layer>
  )
})
