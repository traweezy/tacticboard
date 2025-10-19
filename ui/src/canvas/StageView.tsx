import { useMemo } from 'react'
import Box from '@mui/material/Box'
import Typography from '@mui/material/Typography'
import { Stage } from 'react-konva'
import { useRoomStore, selectNodesArray } from '@state/store'

import { useStageSize } from './hooks/useStageSize'
import { useCanvasShortcuts } from './hooks/useCanvasShortcuts'
import { FieldLayer } from './layers/FieldLayer'
import { GuideLayer } from './layers/GuideLayer'
import { NodeLayer } from './layers/NodeLayer'
import { CursorLayer } from './layers/CursorLayer'

export const StageView = () => {
  const nodes = useRoomStore(selectNodesArray)
  const selectedIds = useRoomStore((state) => state.selectedIds)
  const presence = useRoomStore((state) => Object.values(state.presence))
  const connected = useRoomStore((state) => state.connected)
  const { containerRef, width, height } = useStageSize()
  useCanvasShortcuts()

  const statusMessage = useMemo(() => (connected ? 'Live' : 'Reconnecting…'), [connected])

  return (
    <Box ref={containerRef} sx={{ position: 'relative', flex: 1, minHeight: 0 }}>
      {width === 0 || height === 0 ? null : (
        <Stage width={width} height={height} listening>
          <GuideLayer width={width} height={height} />
          <FieldLayer width={width} height={height} />
          <NodeLayer nodes={nodes} selectedIds={selectedIds} />
          <CursorLayer presence={presence} />
        </Stage>
      )}
      <Typography
        variant="caption"
        sx={{ position: 'absolute', bottom: 12, right: 16, color: 'rgba(255,255,255,0.6)' }}
      >
        {statusMessage}
      </Typography>
    </Box>
  )
}
