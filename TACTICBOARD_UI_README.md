
# TacticBoard Frontend Guide

React 19 + Vite + TypeScript + MUI 7 + Konva. Crisp UI, smooth canvas, real-time sync.

---

## Goals

- Fast canvas interactions at 60 fps.
- Clean, modern UI with MUI. Dark by default.
- Real-time collaboration: cursors, selection, deltas.
- Zero jank on resize and zoom.

---

## Stack

- React 19, Vite, TypeScript
- MUI v7 (Material)
- Konva + react-konva
- Zustand for local state
- Zod for runtime validation
- Dayjs for time
- ESLint + Prettier + Vitest + Testing Library

---

## Setup

```bash
cd ui
pnpm create vite@latest tacticboard-ui -- --template react-ts
cd tacticboard-ui
pnpm add @mui/material @emotion/react @emotion/styled @mui/icons-material          konva react-konva zustand zod dayjs          @tanstack/react-query
pnpm add -D vite-tsconfig-paths eslint @types/konva @testing-library/react @testing-library/user-event vitest jsdom eslint-plugin-react eslint-plugin-react-hooks eslint-plugin-import prettier eslint-config-prettier
```

Vite config:
```ts
// vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tsconfigPaths from 'vite-tsconfig-paths'

export default defineConfig({
  plugins: [react(), tsconfigPaths()],
  server: { port: 5173, strictPort: true, proxy: { '/ws': 'http://localhost:8080', '/api': 'http://localhost:8080' } },
  build: { sourcemap: true },
  test: { environment: 'jsdom', setupFiles: ['./src/test/setup.ts'] }
})
```

TS paths:
```jsonc
// tsconfig.json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": { "@/*": ["src/*"] }
  }
}
```

---

## File tree

```
src/
  app/
    App.tsx
    theme.ts
    routes.tsx
  canvas/
    StageView.tsx
    layers/
      FieldLayer.tsx
      NodeLayer.tsx
      CursorLayer.tsx
      GuideLayer.tsx
    nodes/
      PlayerNode.tsx
      ArrowNode.tsx
      ZoneNode.tsx
      ConeNode.tsx
      FreehandNode.tsx
    hooks/
      useStageSize.ts
      useCanvasShortcuts.ts
  state/
    store.ts               # Zustand store
    selectors.ts
    ops.ts                 # op batching and versioning
  net/
    ws.ts                  # WebSocket client with auto-retry
    sse.ts                 # EventSource fallback
    api.ts                 # REST helpers
  ui/
    Shell.tsx              # App shell with AppBar and Drawer
    Toolbar.tsx
    Palette.tsx            # colors and presets
    LayerPanel.tsx
    ObjectPanel.tsx
    ShareDialog.tsx
    Toasts.tsx
  icons/
    Sports.tsx
  styles/
    theme-overrides.ts
    globals.css
  main.tsx
  index.css
  test/
    setup.ts
```

---

## Theme and style

Create a bold but clean dark theme.

```ts
// src/app/theme.ts
import { createTheme } from '@mui/material/styles'

export const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: { main: '#00E5A8' },
    secondary: { main: '#7C4DFF' },
    background: { default: '#0B0F12', paper: '#12181D' },
    divider: 'rgba(255,255,255,0.08)',
  },
  shape: { borderRadius: 14 },
  typography: {
    fontFamily: 'Inter, ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto',
    h1: { fontWeight: 700, letterSpacing: -0.5 },
    h2: { fontWeight: 700, letterSpacing: -0.25 },
    button: { textTransform: 'none', fontWeight: 600 }
  },
  components: {
    MuiButton: { styleOverrides: { root: { borderRadius: 14 } } },
    MuiPaper: { styleOverrides: { root: { backgroundImage: 'none' } } }
  }
})
```

Wrap app:

```tsx
// src/main.tsx
import React from 'react'
import ReactDOM from 'react-dom/client'
import { ThemeProvider, CssBaseline } from '@mui/material'
import { theme } from '@/app/theme'
import App from '@/app/App'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <ThemeProvider theme={theme}>
    <CssBaseline />
    <App />
  </ThemeProvider>
)
```

Global CSS:

```css
/* src/styles/globals.css */
html, body, #root { height: 100%; }
:root { --app-shell-height: 64px; }
```

---

## Shell layout

```tsx
// src/ui/Shell.tsx
import { AppBar, Toolbar, Typography, Box, Stack } from '@mui/material'
import SportsIcon from '@mui/icons-material/SportsSoccer'

export const Shell = ({ title, toolbar, children }:{ title:string; toolbar?:React.ReactNode; children:React.ReactNode }) => (
  <Stack sx={{ height: '100vh' }}>
    <AppBar position="static" elevation={0}>
      <Toolbar sx={{ gap: 2 }}>
        <SportsIcon />
        <Typography variant="h6">{title}</Typography>
        <Box sx={{ flex: 1 }} />
        {toolbar}
      </Toolbar>
    </AppBar>
    <Box sx={{ flex: 1, minHeight: 0 }}>{children}</Box>
  </Stack>
)
```

---

## Canvas sizing and layers

```tsx
// src/canvas/StageView.tsx
import { Stage, Layer } from 'react-konva'
import { useStageSize } from './hooks/useStageSize'
import { FieldLayer } from './layers/FieldLayer'
import { NodeLayer } from './layers/NodeLayer'
import { CursorLayer } from './layers/CursorLayer'
import { GuideLayer } from './layers/GuideLayer'

export const StageView = () => {
  const { width, height, containerRef } = useStageSize()

  return (
    <div ref={containerRef} style={{ width: '100%', height: '100%' }}>
      <Stage width={width} height={height} listening>
        <Layer><FieldLayer /></Layer>
        <Layer><GuideLayer /></Layer>
        <Layer><NodeLayer /></Layer>
        <Layer listening={false}><CursorLayer /></Layer>
      </Stage>
    </div>
  )
}
```

Adaptive size hook:

```ts
// src/canvas/hooks/useStageSize.ts
import { useLayoutEffect, useRef, useState } from 'react'
export const useStageSize = () => {
  const ref = useRef<HTMLDivElement>(null)
  const [size, set] = useState({ width: 0, height: 0 })
  useLayoutEffect(() => {
    const ro = new ResizeObserver(() => {
      const r = ref.current?.getBoundingClientRect()
      if (r) set({ width: r.width, height: r.height })
    })
    if (ref.current) ro.observe(ref.current)
    return () => ro.disconnect()
  }, [])
  return { ...size, containerRef: ref }
}
```

---

## Field presets

```tsx
// src/canvas/layers/FieldLayer.tsx
import { Group, Rect, Line } from 'react-konva'

export const FieldLayer = () => (
  <Group>
    <Rect x={0} y={0} width={5000} height={3000} fill="#0D1B12" />
    <Line points={[0,1500,5000,1500]} stroke="#2E7D32" strokeWidth={2} />
  </Group>
)
```

Use a scale transform to implement zoom: keep world coordinates large, scale Stage.

---

## Nodes

Player node with drag, selection, and cursor response.

```tsx
// src/canvas/nodes/PlayerNode.tsx
import { Circle, Text, Group } from 'react-konva'
type Props = { id:string; x:number; y:number; label?:string; color:string; selected:boolean; onDragEnd:(x:number,y:number)=>void }
export const PlayerNode = ({ id, x, y, label='P', color, selected, onDragEnd }: Props) => (
  <Group x={x} y={y} draggable onDragEnd={e => onDragEnd(e.target.x(), e.target.y())}>
    <Circle radius={20} fill={color} stroke={selected ? 'white' : 'transparent'} strokeWidth={selected ? 3 : 0} />
    <Text text={label} fontSize={14} offsetX={5} offsetY={8} />
  </Group>
)
```

---

## Toolbar

```tsx
// src/ui/Toolbar.tsx
import { Stack, ToggleButton, ToggleButtonGroup, Button, Divider } from '@mui/material'
import SportsFootballIcon from '@mui/icons-material/SportsFootball'
import AddCircleIcon from '@mui/icons-material/AddCircle'

export const TacticToolbar = ({ tool, setTool, onShare }:{ tool:string; setTool:(t:string)=>void; onShare:()=>void }) => (
  <Stack direction="row" spacing={1} alignItems="center" sx={{ p: 1 }}>
    <ToggleButtonGroup exclusive value={tool} onChange={(_,v)=>v && setTool(v)}>
      <ToggleButton value="select">Select</ToggleButton>
      <ToggleButton value="player"><SportsFootballIcon fontSize="small" /></ToggleButton>
      <ToggleButton value="arrow">Arrow</ToggleButton>
      <ToggleButton value="zone">Zone</ToggleButton>
      <ToggleButton value="freehand">Freehand</ToggleButton>
    </ToggleButtonGroup>
    <Divider flexItem orientation="vertical" />
    <Button variant="contained" startIcon={<AddCircleIcon />} onClick={onShare}>Share</Button>
  </Stack>
)
```

---

## State model

Minimal client state with Zustand. Server state is authoritative.

```ts
// src/state/store.ts
import { create } from 'zustand'

export type Node = { id:string; t:'player'|'arrow'|'zone'|'freehand'; x:number; y:number; color:string; label?:string }
type State = {
  nodes: Record<string, Node>
  selected: string[]
  zoom: number
  pan: { x:number; y:number }
  applyOps: (ops:any[]) => void
}

export const useStore = create<State>((set, get) => ({
  nodes: {}, selected: [], zoom: 1, pan: { x:0, y:0 },
  applyOps: (ops) => set(s => {
    const next = { ...s.nodes }
    for (const op of ops) {
      if (op.k === 'add') next[op.id] = op
      if (op.k === 'move' && next[op.id]) { next[op.id] = { ...next[op.id], x: op.x, y: op.y } }
      if (op.k === 'remove') { delete next[op.id] }
    }
    return { nodes: next }
  })
}))
```

---

## Realtime client

```ts
// src/net/ws.ts
export const connectWS = (roomId: string, onMsg: (msg:any)=>void) => {
  const url = `${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/ws/room/${roomId}`
  let ws = new WebSocket(url)
  ws.onopen = () => ws.send(JSON.stringify({ type: 'hello', roomId, cap: 'edit', since: 0 }))
  ws.onmessage = (e) => onMsg(JSON.parse(e.data))
  ws.onclose = () => setTimeout(()=>connectWS(roomId, onMsg), 1000) // simple backoff
  return ws
}
```

Message handling:

```tsx
// src/app/App.tsx
import { Shell } from '@/ui/Shell'
import { StageView } from '@/canvas/StageView'
import { TacticToolbar } from '@/ui/Toolbar'
import { useEffect, useState } from 'react'
import { useStore } from '@/state/store'
import { connectWS } from '@/net/ws'

export default () => {
  const [tool, setTool] = useState('select')
  const applyOps = useStore(s => s.applyOps)

  useEffect(() => {
    const roomId = new URLSearchParams(location.search).get('room') || 'demo'
    const ws = connectWS(roomId, msg => {
      if (msg.type === 'snapshot') { /* set full state if needed */ }
      if (msg.type === 'delta') applyOps(msg.ops)
    })
    return () => ws.close()
  }, [])

  return (
    <Shell title="TacticBoard" toolbar={<TacticToolbar tool={tool} setTool={setTool} onShare={()=>{}} />}>
      <StageView />
    </Shell>
  )
}
```

---

## Visual rules

- Flat AppBar. No gradients. Primary as accent.
- Border radius 14. Spacing x8.
- Konva.Tween 120 ms for cursor and selection.
- Grid overlay at 30 px. Toggle enabled by default.
- Colors: home `#00E5A8`, away `#7C4DFF`, neutral `#90A4AE`.
- Presence cursors with initials and soft shadow.

---

## Shortcuts

- Space pan
- Wheel zoom to cursor (0.5 to 2)
- Ctrl Z undo, Ctrl Shift Z redo
- Delete remove
- Shift snap to 15°

---

## Accessibility

- `aria-label` on all toolbar buttons
- High-contrast toggle
- Visible focus rings
- `prefers-reduced-motion` respected

---

## Testing

```bash
pnpm test
```

- `useStageSize` unit test
- Toolbar snapshot
- Node drag event test

---

## Build and deploy

```bash
pnpm build
pnpm preview
```

Copy `dist/` to the Go backend `web/` dir or serve standalone. For Go-serve, route unknown paths to `/index.html`.

---

## Polish backlog

- Mini-map viewport
- Rulers, snap guides
- Team libraries
- SVG and PNG export
- Comment pins
- Timeline replay
