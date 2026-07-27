import type { ReactNode } from 'react'

import { Text } from '@mantine/core'

import classes from './SettingsAddAction.module.css'

type Props = {
    icon: ReactNode
    children: ReactNode
    onClick: () => void
    className?: string
}

export function SettingsAddAction({ icon, children, onClick, className }: Props) {
    return (
        <button type="button" className={[classes.button, className].filter(Boolean).join(' ')} onClick={onClick}>
            {icon}
            <Text className={classes.text}>{children}</Text>
        </button>
    )
}
