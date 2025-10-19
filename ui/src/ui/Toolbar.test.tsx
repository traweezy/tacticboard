import { render, screen, fireEvent } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { TacticToolbar } from './Toolbar'

describe('TacticToolbar', () => {
  const onChangeTool = vi.fn()
  const onShare = vi.fn()
  const onSnapToggle = vi.fn()

  beforeEach(() => {
    onChangeTool.mockReset()
    onShare.mockReset()
    onSnapToggle.mockReset()
  })

  it('changes active tool', () => {
    render(
      <TacticToolbar
        activeTool="select"
        onChangeTool={onChangeTool}
        onShare={onShare}
        onSnapToggle={onSnapToggle}
        snapEnabled
      />
    )

    fireEvent.click(screen.getByRole('button', { name: /player/i }))
    expect(onChangeTool).toHaveBeenCalled()
  })

  it('triggers share dialog', () => {
    render(
      <TacticToolbar
        activeTool="select"
        onChangeTool={onChangeTool}
        onShare={onShare}
        onSnapToggle={onSnapToggle}
        snapEnabled
      />
    )

    const shareButtons = screen.getAllByRole('button', { name: /share/i })
    fireEvent.click(shareButtons[0])
    expect(onShare).toHaveBeenCalled()
  })
})
