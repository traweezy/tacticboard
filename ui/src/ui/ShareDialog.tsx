import { useState } from 'react'
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  MenuItem,
  Stack,
  TextField
} from '@mui/material'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import { useShareRoom } from '@net/api'

dayjs.extend(relativeTime)

export type ShareDialogProps = {
  roomId: string
  open: boolean
  onClose: () => void
}

export const ShareDialog = ({ roomId, open, onClose }: ShareDialogProps) => {
  const [role, setRole] = useState<'view' | 'edit'>('view')
  const [ttl, setTtl] = useState(60)
  const shareMutation = useShareRoom()

  const handleShare = async () => {
    const result = await shareMutation.mutateAsync({ roomId, role, ttlMinutes: ttl })
    await navigator.clipboard.writeText(result.link)
    onClose()
  }

  const expiresAt = shareMutation.data?.expiry ? dayjs(shareMutation.data.expiry).fromNow() : null

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="xs">
      <DialogTitle>Share room</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ mt: 1 }}>
          <DialogContentText>
            Generate a capability link for teammates. Share edit links carefully — anyone with the
            link can modify the board.
          </DialogContentText>
          <TextField select label="Role" value={role} onChange={(event) => setRole(event.target.value as 'view' | 'edit')}>
            <MenuItem value="view">View only</MenuItem>
            <MenuItem value="edit">Editor</MenuItem>
          </TextField>
          <TextField
            type="number"
            label="Expires in (minutes)"
            value={ttl}
            inputProps={{ min: 1, max: 1440 }}
            onChange={(event) => setTtl(Number(event.target.value))}
          />
          {shareMutation.data ? (
            <Stack spacing={1}>
              <TextField
                label="Share link"
                value={shareMutation.data.link}
                InputProps={{ readOnly: true }}
                multiline
                minRows={2}
              />
              {expiresAt ? <DialogContentText color="text.secondary">Expires {expiresAt}</DialogContentText> : null}
            </Stack>
          ) : null}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={handleShare} disabled={shareMutation.isPending} variant="contained">
          {shareMutation.isPending ? 'Creating...' : 'Copy link'}
        </Button>
      </DialogActions>
    </Dialog>
  )
}
