import { createContext, useCallback, useContext, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import Snackbar from '@mui/material/Snackbar'
import Alert from '@mui/material/Alert'

export type Toast = {
  message: string
  severity?: 'success' | 'info' | 'warning' | 'error'
}

const ToastContext = createContext<{
  push: (toast: Toast) => void
} | null>(null)

export const ToastProvider = ({ children }: { children: ReactNode }) => {
  const [toast, setToast] = useState<Toast | null>(null)

  const push = useCallback((next: Toast) => {
    setToast(next)
  }, [])

  const handleClose = useCallback(() => setToast(null), [])

  const value = useMemo(() => ({ push }), [push])

  return (
    <ToastContext.Provider value={value}>
      {children}
      <Snackbar open={Boolean(toast)} autoHideDuration={4000} onClose={handleClose}>
        <Alert
          onClose={handleClose}
          severity={toast?.severity ?? 'info'}
          variant="filled"
          sx={{ width: '100%' }}
        >
          {toast?.message}
        </Alert>
      </Snackbar>
    </ToastContext.Provider>
  )
}

export const useToasts = () => {
  const ctx = useContext(ToastContext)
  if (!ctx) {
    throw new Error('useToasts must be used within ToastProvider')
  }
  return ctx
}
