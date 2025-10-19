import { memo } from 'react'
import { Rect } from 'react-konva'
import type { BoardNode } from '@state/store'

export interface ZoneNodeProps {
  node: BoardNode
  isSelected: boolean
}

export const ZoneNode = memo(function ZoneNodeComponent({ node, isSelected }: ZoneNodeProps) {
  const width = node.points?.[0] ?? 120
  const height = node.points?.[1] ?? 80
  const stroke = node.color ?? 'rgba(144,164,174,0.8)'
  return (
    <Rect
      x={node.x - width / 2}
      y={node.y - height / 2}
      width={width}
      height={height}
      stroke={stroke}
      strokeWidth={isSelected ? 4 : 2}
      dash={[12, 8]}
      cornerRadius={12}
      listening
    />
  )
})
