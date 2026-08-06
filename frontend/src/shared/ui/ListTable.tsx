import type { ReactNode } from 'react'

import { AppTable, appTableClasses } from './AppTable'
import classes from './ListTable.module.css'

type Props = {
    children: ReactNode
    minWidth: number
    verticalSpacing?: number | 'xs' | 'sm' | 'md' | 'lg' | 'xl'
}

export const listTableClasses = {
    interactiveRow: appTableClasses.interactiveRow,
    compactRow: appTableClasses.compactRow,
    bodyRow: classes.bodyRow,
    headerRow: classes.headerRow,
}

export function ListTable({ children, minWidth, verticalSpacing = 0 }: Props) {
    return (
        <AppTable minWidth={minWidth} verticalSpacing={verticalSpacing}>
            {children}
        </AppTable>
    )
}
