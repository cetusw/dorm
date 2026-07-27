import type { ReactNode } from 'react'

import { Paper } from '@mantine/core'

import classes from './SettingsCardSurface.module.css'

type Props = {
    children: ReactNode
    className?: string
    dragging?: boolean
    menuOpen?: boolean
    muted?: boolean
}

export function SettingsCardSurface({ children, className, dragging = false, menuOpen = false, muted = false }: Props) {
    return (
        <Paper
            withBorder
            p="md"
            className={[classes.card, className].filter(Boolean).join(' ')}
            data-dragging={dragging ? 'true' : undefined}
            data-menu-open={menuOpen ? 'true' : undefined}
            data-muted={muted ? 'true' : undefined}
        >
            {children}
        </Paper>
    )
}
