import { useMutation, useQuery } from '@tanstack/react-query'
import { z } from 'zod'

const roomSchema = z.object({
  id: z.string(),
  currentSeq: z.number().int(),
  snapshot: z
    .object({
      seq: z.number().int(),
      state: z.object({ nodes: z.array(z.any()).optional() }).passthrough()
    })
    .optional()
})

const shareResponseSchema = z.object({
  token: z.string(),
  role: z.enum(['view', 'edit']),
  expiry: z.string().or(z.date()).optional(),
  link: z.string()
})

const jsonFetch = async <T>(input: RequestInfo, init?: RequestInit) => {
  const response = await fetch(input, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...(init?.headers ?? {}) }
  })

  if (!response.ok) {
    const errorBody = await response.json().catch(() => ({}))
    throw new Error(errorBody.error ?? response.statusText)
  }
  return (await response.json()) as T
}

export const useRoom = (roomId: string | null) =>
  useQuery({
    queryKey: ['room', roomId],
    enabled: Boolean(roomId),
    queryFn: async () => {
      const data = await jsonFetch(`/api/rooms/${roomId}`)
      return roomSchema.parse(data)
    },
    staleTime: 3000
  })

export const useShareRoom = () =>
  useMutation({
    mutationFn: async (payload: { roomId: string; role: 'view' | 'edit'; ttlMinutes: number }) => {
      const body = JSON.stringify({ role: payload.role, ttlMinutes: payload.ttlMinutes })
      const data = await jsonFetch(`/api/rooms/${payload.roomId}/share`, { method: 'POST', body })
      return shareResponseSchema.parse(data)
    }
  })

export type RoomResponse = z.infer<typeof roomSchema>
export type ShareResponse = z.infer<typeof shareResponseSchema>
