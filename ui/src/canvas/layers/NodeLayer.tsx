import { memo } from 'react'
import { Layer } from 'react-konva'
import type { BoardNode } from '@state/store'

import { ArrowNode } from '../nodes/ArrowNode'
import { ConeNode } from '../nodes/ConeNode'
import { FreehandNode } from '../nodes/FreehandNode'
import { PlayerNode } from '../nodes/PlayerNode'
import { ZoneNode } from '../nodes/ZoneNode'

export type NodeLayerProps = {
  nodes: BoardNode[]
  selectedIds: string[]
}

export const NodeLayer = memo(function NodeLayerComponent({ nodes, selectedIds }: NodeLayerProps) {
  const selectedSet = new Set(selectedIds)

  return (
    <Layer>
      {nodes.map((node) => {
        const isSelected = selectedSet.has(node.id)
        switch (node.kind) {
          case 'player':
            return <PlayerNode key={node.id} node={node} isSelected={isSelected} />
          case 'arrow':
            return <ArrowNode key={node.id} node={node} isSelected={isSelected} />
          case 'zone':
            return <ZoneNode key={node.id} node={node} isSelected={isSelected} />
          case 'cone':
            return <ConeNode key={node.id} node={node} isSelected={isSelected} />
          case 'freehand':
          default:
            return <FreehandNode key={node.id} node={node} isSelected={isSelected} />
        }
      })}
    </Layer>
  )
})
