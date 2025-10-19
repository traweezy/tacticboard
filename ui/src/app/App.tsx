import { useEffect, useMemo, useRef, useState } from 'react'
import Box from '@mui/material/Box'
import CircularProgress from '@mui/material/CircularProgress'
import Stack from '@mui/material/Stack'
import { ThemeProvider } from '@mui/material/styles'
import Typography from '@mui/material/Typography'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { StageView } from '@canvas/StageView'
import { useRoom } from '@net/api'
import { deltaMessageSchema, errorMessageSchema, nodeSchema, snapshotMessageSchema } from '@net/schema'
import { connectWS } from '@net/ws'
import { useRoomStore } from '@state/store'
import type { BoardNode, Operation, Presence } from '@state/store'
import { LayerPanel } from '@ui/LayerPanel'
import { ObjectPanel } from '@ui/ObjectPanel'
import { Palette } from '@ui/Palette'
import { ShareDialog } from '@ui/ShareDialog'
import { Shell } from '@ui/Shell'
import { ToastProvider, useToasts } from '@ui/Toasts'
import { TacticToolbar } from '@ui/Toolbar'

import { getInitialRoomContext } from './room-context'
import { theme } from './theme'

type PresenceMessage = {
  type: 'presence'
  clients: Presence[]
}

const isPresenceMessage = (value: unknown): value is PresenceMessage => {
  if (!value || typeof value !== 'object') {
    return false
  }
  const message = value as Partial<PresenceMessage>
  return message.type === 'presence' && Array.isArray(message.clients)
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1
    }
  }
})

const AppInner = () => {
  const [{ roomId, token, capability }] = useState(() => getInitialRoomContext())
  const setRoom = useRoomStore((state) => state.setRoom)
  const applySnapshot = useRoomStore((state) => state.applySnapshot)
  const applyOperations = useRoomStore((state) => state.applyOperations)
  const setConnected = useRoomStore((state) => state.setConnected)
  const latestSeq = useRoomStore((state) => state.latestSeq)
  const tool = useRoomStore((state) => state.tool)
  const setTool = useRoomStore((state) => state.setTool)
  const snap = useRoomStore((state) => state.snap)
  const toggleSnap = useRoomStore((state) => state.toggleSnap)
  const updatePresence = useRoomStore((state) => state.updatePresence)
  const { push } = useToasts()
  const [shareOpen, setShareOpen] = useState(false)
  const latestSeqRef = useRef(latestSeq)
  const roomQuery = useRoom(roomId)

  useEffect(() => {
    setRoom(roomId, capability)
  }, [roomId, capability, setRoom])

  useEffect(() => {
    if (!roomQuery.data?.snapshot?.state?.nodes) {
      return
    }
    const nodes: BoardNode[] = (roomQuery.data.snapshot.state.nodes ?? []).map((node) => {
      const parsed = nodeSchema.parse(node)
      return {
        id: parsed.id,
        kind: (parsed.kind as BoardNode['kind']) ?? 'player',
        x: parsed.x,
        y: parsed.y,
        rotation: parsed.rotation,
        color: parsed.color,
        label: parsed.label,
        points: parsed.points
      }
    })
    applySnapshot({
      seq: roomQuery.data.snapshot.seq,
      nodes
    })
  }, [roomQuery.data, applySnapshot])

  useEffect(() => {
    latestSeqRef.current = latestSeq
  }, [latestSeq])

  useEffect(() => {
    if (!roomId || !token) {
      return
    }

    const connection = connectWS({
      roomId,
      token,
      capability,
      since: latestSeqRef.current,
      callbacks: {
        onOpen: () => setConnected(true),
        onClose: () => setConnected(false),
        onMessage: (raw) => {
          const snapshot = snapshotMessageSchema.safeParse(raw)
          if (snapshot.success) {
            const nodes: BoardNode[] = (snapshot.data.state.nodes ?? []).map((node) => ({
              id: node.id,
              kind: (node.kind as BoardNode['kind']) ?? 'player',
              x: node.x,
              y: node.y,
              rotation: node.rotation,
              color: node.color,
              label: node.label,
              points: node.points
            }))
            applySnapshot({ seq: snapshot.data.seq, nodes })
            return
          }

          const delta = deltaMessageSchema.safeParse(raw)
          if (delta.success) {
            const ops: Operation[] = delta.data.ops.map((op) => {
              switch (op.k) {
                case 'add':
                  return {
                    k: 'add' as const,
                    node: {
                      id: op.node.id,
                      kind: (op.node.kind as BoardNode['kind']) ?? 'player',
                      x: op.node.x,
                      y: op.node.y,
                      rotation: op.node.rotation,
                      color: op.node.color,
                      label: op.node.label,
                      points: op.node.points
                    }
                  }
                case 'move':
                  return {
                    k: 'move',
                    id: op.id,
                    x: op.x,
                    y: op.y
                  }
                case 'patch': {
                  const changes = { ...op.changes } as Partial<BoardNode>
                  if (changes.kind) {
                    changes.kind = changes.kind as BoardNode['kind']
                  }
                  return {
                    k: 'patch',
                    id: op.id,
                    changes
                  }
                }
                case 'remove':
                default:
                  return { k: 'remove', id: op.id }
              }
            })
            applyOperations(ops, delta.data.to)
            return
          }

          const error = errorMessageSchema.safeParse(raw)
          if (error.success) {
            push({ message: error.data.msg, severity: 'error' })
            return
          }

          if (isPresenceMessage(raw)) {
            updatePresence(raw.clients)
          }
        },
        onError: () => {
          push({ message: 'Connection interrupted', severity: 'warning' })
        }
      }
    })

    return () => {
      connection.close()
    }
  }, [roomId, token, capability, applySnapshot, applyOperations, updatePresence, push, setConnected])

  const toolbar = useMemo(
    () => (
      <TacticToolbar
        activeTool={tool}
        onChangeTool={setTool}
        onShare={() => setShareOpen(true)}
        onSnapToggle={toggleSnap}
        snapEnabled={snap}
      />
    ),
    [tool, setTool, toggleSnap, snap]
  )

  const sidePanel = useMemo(
    () => (
      <Stack spacing={2} sx={{ p: 2, overflowY: 'auto' }}>
        <Palette />
        <LayerPanel />
        <ObjectPanel />
      </Stack>
    ),
    []
  )

  if (roomQuery.isLoading) {
    return (
      <Box sx={{ display: 'grid', placeItems: 'center', height: '100%' }}>
        <CircularProgress />
      </Box>
    )
  }

  if (roomQuery.isError) {
    return (
      <Box sx={{ display: 'grid', placeItems: 'center', height: '100%', textAlign: 'center' }}>
        <Typography variant="h6" gutterBottom>
          Failed to load room
        </Typography>
        <Typography color="text.secondary">{(roomQuery.error as Error).message}</Typography>
      </Box>
    )
  }

  return (
    <>
      <Shell title="TacticBoard" toolbar={toolbar} sidePanel={sidePanel}>
        <StageView />
      </Shell>
      <ShareDialog roomId={roomId} open={shareOpen} onClose={() => setShareOpen(false)} />
    </>
  )
}

export const App = () => (
  <ThemeProvider theme={theme}>
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <AppInner />
      </ToastProvider>
    </QueryClientProvider>
  </ThemeProvider>
)
