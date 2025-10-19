import { memo } from 'react'
import { RegularPolygon } from 'react-konva'
import type { BoardNode } from '@state/store'

export interface ConeNodeProps {
  node: BoardNode
  isSelected: boolean
}

export const ConeNode = memo(function ConeNodeComponent({ node, isSelected }: ConeNodeProps) {
  const fill = node.color ?? '#ffb74d'
  return (
    <RegularPolygon
      x={node.x}
      y={node.y}
      sides={3}
      radius={22}
      rotation={-90}
      fill={fill}
      stroke="rgba(0,0,0,0.3)"
      strokeWidth={isSelected ? 3 : 1}
    />
  )
})
