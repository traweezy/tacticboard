import { alpha, createTheme } from '@mui/material/styles'

const primary = '#00e5a8'
const secondary = '#7c4dff'
const background = {
  default: '#07090d',
  paper: '#0f121a'
}

export const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: primary
    },
    secondary: {
      main: secondary
    },
    background,
    text: {
      primary: 'rgba(255,255,255,0.92)',
      secondary: 'rgba(255,255,255,0.64)'
    }
  },
  typography: {
    fontFamily: "'Inter', system-ui, sans-serif",
    h1: { fontWeight: 700 },
    h2: { fontWeight: 600 },
    button: { textTransform: 'none', fontWeight: 600 }
  },
  shape: {
    borderRadius: 14
  },
  components: {
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: 'none'
        }
      }
    },
    MuiButton: {
      defaultProps: {
        disableElevation: true
      },
      styleOverrides: {
        root: ({ ownerState }) => ({
          borderRadius: 12,
          paddingInline: 16,
          paddingBlock: 10,
          ...(ownerState.variant === 'outlined' && {
            borderColor: alpha(primary, 0.4)
          })
        })
      }
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          backgroundColor: background.paper,
          borderBottom: `1px solid ${alpha('#ffffff', 0.08)}`
        }
      }
    }
  }
})
