import type { DeltaMessage, SnapshotMessage } from './schema'

type PongMessage = { type: 'pong'; ts: number }
type ErrorMessage = { type: 'error'; code: string; msg: string }
type UnknownMessage = { type: string; [key: string]: unknown }

export type WSMessage = SnapshotMessage | DeltaMessage | PongMessage | ErrorMessage | UnknownMessage

export type SocketCallbacks = {
  onOpen?: () => void
  onClose?: (event: CloseEvent) => void
  onError?: (event: Event) => void
  onMessage?: (message: WSMessage) => void
}

export type ConnectParams = {
  roomId: string
  token: string
  capability: 'view' | 'edit'
  since: number
  callbacks?: SocketCallbacks
}

const serializeHello = (roomId: string, capability: 'view' | 'edit', token: string, since: number) =>
  JSON.stringify({ type: 'hello', roomId, cap: capability, token, since })

const createUrl = (roomId: string) => {
  const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const host = window.location.host
  return `${protocol}://${host}/ws/room/${roomId}`
}

export const connectWS = ({ roomId, token, capability, since, callbacks }: ConnectParams) => {
  let socket: WebSocket | null = null
  let attempt = 0
  let closed = false

  const connect = () => {
    if (closed) return
    socket = new WebSocket(createUrl(roomId))
    const localSocket = socket

    localSocket.addEventListener('open', () => {
      attempt = 0
      localSocket.send(serializeHello(roomId, capability, token, since))
      callbacks?.onOpen?.()
    })

    localSocket.addEventListener('message', (event) => {
      try {
        const data = JSON.parse(event.data) as WSMessage
        callbacks?.onMessage?.(data)
      } catch {
        callbacks?.onError?.(new Event('parse-error'))
      }
    })

    localSocket.addEventListener('error', (event) => {
      callbacks?.onError?.(event)
    })

    localSocket.addEventListener('close', (event) => {
      callbacks?.onClose?.(event)
      if (!closed) {
        attempt += 1
        const backoff = Math.min(1000 * 2 ** attempt, 8000)
        window.setTimeout(connect, backoff)
      }
    })
  }

  connect()

  const send = (payload: unknown) => {
    if (!socket || socket.readyState !== WebSocket.OPEN) return false
    socket.send(JSON.stringify(payload))
    return true
  }

  const close = () => {
    closed = true
    socket?.close()
  }

  return { send, close }
}
