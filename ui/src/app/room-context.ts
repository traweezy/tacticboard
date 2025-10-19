export const getInitialRoomContext = () => {
  const search = new URLSearchParams(window.location.search)
  const pathParts = window.location.pathname.split('/').filter(Boolean)
  const roomFromPath = pathParts.length >= 2 && pathParts[0] === 'room' ? pathParts[1] : null
  const roomId = search.get('room') ?? roomFromPath ?? 'demo'
  const token = search.get('token') ?? ''
  const capability = (search.get('cap') as 'view' | 'edit' | null) ?? 'view'
  return { roomId, token, capability }
}
