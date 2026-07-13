import type { ReactNode } from 'react'

import { MantineProvider } from '@mantine/core'

type Props = {
    children: ReactNode
}

export function AppProviders({ children }: Props) {
    return (
        <MantineProvider defaultColorScheme="light">
            {children}
        </MantineProvider>
    )
}
