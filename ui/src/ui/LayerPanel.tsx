import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import Typography from '@mui/material/Typography'
import { useRoomStore } from '@state/store'

export const LayerPanel = () => {
  const nodes = useRoomStore((state) => state.nodes)
  const total = Object.keys(nodes).length
  return (
    <Card elevation={0} sx={{ backgroundColor: 'transparent', borderRadius: 0 }}>
      <CardContent>
        <Typography variant="subtitle1" gutterBottom>
          Layers
        </Typography>
        <List dense disablePadding>
          <ListItem disablePadding>
            <ListItemText primary="Offense" secondary="Visible" />
          </ListItem>
          <ListItem disablePadding>
            <ListItemText primary="Defense" secondary="Visible" />
          </ListItem>
        </List>
        <Typography variant="caption" color="text.secondary">
          {total} items on board
        </Typography>
      </CardContent>
    </Card>
  )
}
