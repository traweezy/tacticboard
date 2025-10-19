import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Stack from '@mui/material/Stack'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'
import { useRoomStore } from '@state/store'

export const ObjectPanel = () => {
  const selected = useRoomStore((state) => state.selectedIds)
  const node = useRoomStore((state) => (selected.length === 1 ? state.nodes[selected[0]] : null))

  if (!node) {
    return (
      <Card elevation={0} sx={{ backgroundColor: 'transparent', borderRadius: 0 }}>
        <CardContent>
          <Typography variant="subtitle1">Object</Typography>
          <Typography variant="body2" color="text.secondary">
            Select a node to view details.
          </Typography>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card elevation={0} sx={{ backgroundColor: 'transparent', borderRadius: 0 }}>
      <CardContent>
        <Typography variant="subtitle1" gutterBottom>
          {node.kind} node
        </Typography>
        <Stack spacing={1.5}>
          <TextField label="Label" value={node.label ?? ''} size="small" disabled fullWidth />
          <TextField label="Color" value={node.color ?? ''} size="small" disabled fullWidth />
        </Stack>
      </CardContent>
    </Card>
  )
}
