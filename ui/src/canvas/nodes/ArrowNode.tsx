import { memo } from 'react'
import { Arrow } from 'react-konva'
import type { BoardNode } from '@state/store'

export interface ArrowNodeProps {
  node: BoardNode
  isSelected: boolean
}

export const ArrowNode = memo(function ArrowNodeComponent({ node, isSelected }: ArrowNodeProps) {
  const points = node.points?.length ? node.points : [0, 0, 60, 0]
  const stroke = node.color ?? '#7c4dff'
  return (
    <Arrow
      x={node.x}
      y={node.y}
      points={points}
      pointerLength={16}
      pointerWidth={16}
      stroke={stroke}
      fill={stroke}
      strokeWidth={4}
      opacity={isSelected ? 1 : 0.9}
    />
  )
})
