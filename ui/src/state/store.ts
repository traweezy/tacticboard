import { create } from 'zustand'
import { devtools } from 'zustand/middleware'

export type NodeKind = 'player' | 'arrow' | 'zone' | 'cone' | 'freehand'

export type BoardNode = {
  id: string
  kind: NodeKind
  x: number
  y: number
  rotation?: number
  color?: string
  label?: string
  points?: number[]
}

export type Operation =
  | { k: 'add'; node: BoardNode }
  | { k: 'move'; id: string; x: number; y: number }
  | { k: 'patch'; id: string; changes: Partial<BoardNode> }
  | { k: 'remove'; id: string }

export type Presence = {
  clientId: string
  name: string
  x: number
  y: number
  color: string
  updatedAt: number
}

export type SnapshotPayload = {
  seq: number
  nodes: BoardNode[]
}

export type RoomState = {
  roomId: string | null
  capability: 'view' | 'edit'
  connected: boolean
  latestSeq: number
  nodes: Record<string, BoardNode>
  selectedIds: string[]
  tool: string
  snap: boolean
  presence: Record<string, Presence>
}

export type RoomActions = {
  setRoom: (roomId: string, capability: 'view' | 'edit') => void
  setConnected: (connected: boolean) => void
  applySnapshot: (payload: SnapshotPayload) => void
  applyOperations: (ops: Operation[], toSeq: number) => void
  setTool: (tool: string) => void
  toggleSnap: () => void
  updatePresence: (presence: Presence[]) => void
  selectNodes: (ids: string[]) => void
}

export const useRoomStore = create<RoomState & RoomActions>()(
  devtools((set) => ({
    roomId: null,
    capability: 'view',
    connected: false,
    latestSeq: 0,
    nodes: {},
    selectedIds: [],
    tool: 'select',
    snap: true,
    presence: {},

    setRoom: (roomId, capability) =>
      set({
        roomId,
        capability,
        nodes: {},
        latestSeq: 0,
        selectedIds: [],
        presence: {}
      }),

    setConnected: (connected) => set({ connected }),

    applySnapshot: (payload) =>
      set((state) => {
        if (payload.seq <= state.latestSeq) {
          return state
        }
        const nodes: Record<string, BoardNode> = {}
        for (const node of payload.nodes) {
          nodes[node.id] = node
        }
        return {
          nodes,
          latestSeq: payload.seq
        }
      }),

    applyOperations: (ops, toSeq) =>
      set((state) => {
        if (toSeq <= state.latestSeq) {
          return state
        }
        const nodes = { ...state.nodes }
        for (const op of ops) {
          switch (op.k) {
            case 'add':
              nodes[op.node.id] = op.node
              break
            case 'move':
              if (nodes[op.id]) {
                nodes[op.id] = { ...nodes[op.id], x: op.x, y: op.y }
              }
              break
            case 'patch':
              if (nodes[op.id]) {
                nodes[op.id] = { ...nodes[op.id], ...op.changes }
              }
              break
            case 'remove':
              delete nodes[op.id]
              break
            default:
              break
          }
        }
        return {
          nodes,
          latestSeq: toSeq
        }
      }),

    setTool: (tool) => set({ tool }),

    toggleSnap: () => set((state) => ({ snap: !state.snap })),

    updatePresence: (presenceList) =>
      set(() => {
        const next: Record<string, Presence> = {}
        for (const presence of presenceList) {
          next[presence.clientId] = presence
        }
        return { presence: next }
      }),

    selectNodes: (ids) => set({ selectedIds: ids })
  }))
)

export const selectNodesArray = (state: RoomState) => Object.values(state.nodes)
