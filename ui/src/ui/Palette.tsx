import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import IconButton from '@mui/material/IconButton'
import Tooltip from '@mui/material/Tooltip'

const colors = [
  '#00e5a8',
  '#7c4dff',
  '#f06292',
  '#ffb74d',
  '#90a4ae'
]

export const Palette = () => (
  <Stack spacing={1} sx={{ p: 2 }}>
    <Typography variant="subtitle1">Palette</Typography>
    <Stack direction="row" spacing={1}>
      {colors.map((color) => (
        <Tooltip title={color} key={color}>
          <IconButton
            size="small"
            sx={{ backgroundColor: color, '&:hover': { backgroundColor: color } }}
            aria-label={`Select ${color}`}
          />
        </Tooltip>
      ))}
    </Stack>
  </Stack>
)
