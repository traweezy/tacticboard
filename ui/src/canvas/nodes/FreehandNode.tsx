import { memo } from 'react'
import { Line } from 'react-konva'
import type { BoardNode } from '@state/store'

export interface FreehandNodeProps {
  node: BoardNode
  isSelected: boolean
}

export const FreehandNode = memo(function FreehandNodeComponent({ node, isSelected }: FreehandNodeProps) {
  const points = node.points?.length ? node.points : [0, 0, 0, 0]
  const stroke = node.color ?? '#90a4ae'
  return (
    <Line
      x={node.x}
      y={node.y}
      points={points}
      stroke={stroke}
      strokeWidth={isSelected ? 6 : 4}
      lineCap="round"
      lineJoin="round"
      tension={0.4}
    />
  )
})
