import type { ReactNode } from 'react'

import { Table } from '@mantine/core'

import classes from './AppTable.module.css'

type Props = {
    children: ReactNode
    minWidth: number
    verticalSpacing?: number | 'xs' | 'sm' | 'md' | 'lg' | 'xl'
}

export { classes as appTableClasses }

export function AppTable({
    children,
    minWidth,
    verticalSpacing = 0,
}: Props) {
    return (
        <Table.ScrollContainer minWidth={minWidth} className={classes.table}>
            <Table horizontalSpacing="lg" verticalSpacing={verticalSpacing} withRowBorders>
                {children}
            </Table>
        </Table.ScrollContainer>
    )
}
