import { memo } from 'react'
import { Circle, Group, Text } from 'react-konva'
import type { BoardNode } from '@state/store'

export interface PlayerNodeProps {
  node: BoardNode
  isSelected: boolean
}

export const PlayerNode = memo(function PlayerNodeComponent({ node, isSelected }: PlayerNodeProps) {
  const radius = 24
  const color = node.color ?? '#00e5a8'
  return (
    <Group x={node.x} y={node.y} listening>
      <Circle
        radius={radius}
        fill={color}
        opacity={isSelected ? 1 : 0.85}
        stroke="rgba(255,255,255,0.6)"
        strokeWidth={isSelected ? 4 : 2}
      />
      <Text
        text={node.label ?? 'P'}
        fontSize={16}
        fill="#021519"
        width={radius * 2}
        height={radius * 2}
        offset={{ x: radius, y: radius }}
        align="center"
        verticalAlign="middle"
        fontStyle="700"
      />
    </Group>
  )
})
