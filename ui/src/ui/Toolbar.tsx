import Button from '@mui/material/Button'
import ButtonGroup from '@mui/material/ButtonGroup'
import Stack from '@mui/material/Stack'
import Tooltip from '@mui/material/Tooltip'
import ShareIcon from '@mui/icons-material/IosShare'
import BoltIcon from '@mui/icons-material/Bolt'

const tools = [
  { id: 'select', label: 'Select' },
  { id: 'player', label: 'Player' },
  { id: 'arrow', label: 'Arrow' },
  { id: 'zone', label: 'Zone' }
] as const

export interface TacticToolbarProps {
  activeTool: string
  onChangeTool: (tool: string) => void
  onShare: () => void
  onSnapToggle: () => void
  snapEnabled: boolean
}

export const TacticToolbar = ({ activeTool, onChangeTool, onShare, onSnapToggle, snapEnabled }: TacticToolbarProps) => (
  <Stack direction="row" spacing={1} alignItems="center">
    <ButtonGroup color="primary" size="small" aria-label="tool selection">
      {tools.map((tool) => (
        <Button
          key={tool.id}
          variant={tool.id === activeTool ? 'contained' : 'outlined'}
          onClick={() => onChangeTool(tool.id)}
          aria-pressed={tool.id === activeTool}
        >
          {tool.label}
        </Button>
      ))}
    </ButtonGroup>
    <Tooltip title={snapEnabled ? 'Disable snap-to-grid' : 'Enable snap-to-grid'}>
      <Button
        variant={snapEnabled ? 'contained' : 'outlined'}
        size="small"
        startIcon={<BoltIcon />}
        onClick={onSnapToggle}
        aria-pressed={snapEnabled}
      >
        Snap
      </Button>
    </Tooltip>
    <Button variant="contained" size="small" startIcon={<ShareIcon />} onClick={onShare}>
      Share
    </Button>
  </Stack>
)
