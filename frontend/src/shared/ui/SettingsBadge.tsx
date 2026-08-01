import type { ReactNode } from 'react'

import { Box, Text } from '@mantine/core'

import classes from './SettingsBadge.module.css'

type Props = {
    children: ReactNode
    color?: 'default' | 'success' | 'leader'
}

export function SettingsBadge({ children, color = 'default' }: Props) {
    return (
        <Box
            className={[
                classes.badge,
                color === 'success' ? classes.success : '',
                color === 'leader' ? classes.leader : '',
            ].filter(Boolean).join(' ')}
        >
            <Text size="sm">{children}</Text>
        </Box>
    )
}
