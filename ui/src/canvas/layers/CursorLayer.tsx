import { memo } from 'react'
import { Circle, Group, Layer, Text } from 'react-konva'
import type { Presence } from '@state/store'

export type CursorLayerProps = {
  presence: Presence[]
}

export const CursorLayer = memo(function CursorLayerComponent({ presence }: CursorLayerProps) {
  return (
    <Layer listening={false}>
      {presence.map((cursor) => (
        <Group key={cursor.clientId} x={cursor.x} y={cursor.y}>
          <Circle radius={10} fill={cursor.color} shadowBlur={12} shadowColor={cursor.color} />
          <Text
            text={cursor.name.slice(0, 2).toUpperCase()}
            fontSize={10}
            fill="#000"
            offsetX={10}
            offsetY={5}
            align="center"
            width={20}
          />
        </Group>
      ))}
    </Layer>
  )
})
