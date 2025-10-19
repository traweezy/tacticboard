import type { PropsWithChildren, ReactNode } from 'react'
import AppBar from '@mui/material/AppBar'
import Box from '@mui/material/Box'
import Container from '@mui/material/Container'
import CssBaseline from '@mui/material/CssBaseline'
import Divider from '@mui/material/Divider'
import IconButton from '@mui/material/IconButton'
import Toolbar from '@mui/material/Toolbar'
import Typography from '@mui/material/Typography'
import MenuIcon from '@mui/icons-material/Menu'

export type ShellProps = PropsWithChildren<{
  title: string
  toolbar?: ReactNode
  sidePanel?: ReactNode
}>

export const Shell = ({ title, toolbar, sidePanel, children }: ShellProps) => (
  <Box sx={{ display: 'grid', gridTemplateRows: 'auto 1fr', height: '100%' }}>
    <CssBaseline />
    <AppBar position="sticky" elevation={0}>
      <Toolbar sx={{ gap: 2 }}>
        <IconButton edge="start" color="inherit" sx={{ display: { sm: 'none' } }} aria-label="open navigation">
          <MenuIcon />
        </IconButton>
        <Typography variant="h6" component="h1" sx={{ fontWeight: 700 }}>
          {title}
        </Typography>
        <Box sx={{ flexGrow: 1 }} />
        {toolbar}
      </Toolbar>
      <Divider sx={{ borderColor: 'rgba(255,255,255,0.08)' }} />
    </AppBar>
    <Box sx={{ display: 'flex', flex: 1, minHeight: 0 }}>
      {sidePanel ? (
        <Box
          component="aside"
          sx={{
            display: { xs: 'none', md: 'flex' },
            width: 320,
            borderRight: '1px solid rgba(255,255,255,0.08)',
            backgroundColor: 'background.paper',
            flexDirection: 'column'
          }}
        >
          {sidePanel}
        </Box>
      ) : null}
      <Container
        component="main"
        disableGutters
        sx={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0, position: 'relative' }}
      >
        {children}
      </Container>
    </Box>
  </Box>
)
