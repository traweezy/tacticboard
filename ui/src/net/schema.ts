import { z } from 'zod'

export const nodeSchema = z.object({
  id: z.string(),
  kind: z.enum(['player', 'arrow', 'zone', 'cone', 'freehand']).or(z.string()),
  x: z.number(),
  y: z.number(),
  rotation: z.number().optional(),
  color: z.string().optional(),
  label: z.string().optional(),
  points: z.array(z.number()).optional()
})

export const operationSchema = z.discriminatedUnion('k', [
  z.object({ k: z.literal('add'), node: nodeSchema }),
  z.object({ k: z.literal('move'), id: z.string(), x: z.number(), y: z.number() }),
  z.object({ k: z.literal('patch'), id: z.string(), changes: nodeSchema.partial() }),
  z.object({ k: z.literal('remove'), id: z.string() })
])

export const snapshotMessageSchema = z.object({
  type: z.literal('snapshot'),
  roomId: z.string(),
  seq: z.number(),
  state: z.object({ nodes: z.array(nodeSchema).optional() }).passthrough()
})

export const deltaMessageSchema = z.object({
  type: z.literal('delta'),
  roomId: z.string(),
  from: z.number(),
  to: z.number(),
  ops: z.array(operationSchema)
})

export const errorMessageSchema = z.object({
  type: z.literal('error'),
  code: z.string(),
  msg: z.string()
})

export type NodePayload = z.infer<typeof nodeSchema>
export type OperationPayload = z.infer<typeof operationSchema>
export type SnapshotMessage = z.infer<typeof snapshotMessageSchema>
export type DeltaMessage = z.infer<typeof deltaMessageSchema>
