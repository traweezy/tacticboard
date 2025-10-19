# TacticBoard UI

React 19 + Vite + TypeScript frontend for the TacticBoard experience. The app renders the Konva stage, syncs real-time operations over WebSocket, and wraps Material UI controls to manage rooms.

## Scripts

```bash
pnpm install    # install dependencies
pnpm dev        # start Vite dev server (http://localhost:5173)
pnpm lint       # run eslint
pnpm test       # run vitest suite
pnpm build      # type-check and build production bundle to dist/
```

The dev server proxies `/api` and `/ws` to the Go backend on `localhost:8080`. Provide `room` and `token` query params in the browser URL to join a board, e.g. `http://localhost:5173/room/demo?token=...&cap=edit`.

## Structure

- `src/app/` — top-level app composition, theming, query client
- `src/canvas/` — Konva-based stage, layers, and hooks
- `src/state/` — Zustand store and op batching helper
- `src/net/` — REST client, WebSocket connector, and zod schemas
- `src/ui/` — Material UI shell, toolbar, dialogs, palette, toasts
- `src/test/` — shared testing utilities

## Testing

The vitest suite covers toolbar interactions and the resize observer hook. Extend with additional canvas and network tests as behaviours evolve.

## Build Output

`pnpm build` emits a production bundle under `dist/`. Copy the contents into the Go server's `web/` directory or serve with a static host.
