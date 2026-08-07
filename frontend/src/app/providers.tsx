import type { ReactNode } from 'react'

import { createTheme, MantineProvider } from '@mantine/core'

type Props = {
    children: ReactNode
}

const theme = createTheme({
    primaryColor: 'brand',
    primaryShade: { light: 6, dark: 7 },
    defaultRadius: 'md',
    fontFamily: 'Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
    headings: {
        fontFamily: 'Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
    },
    black: '#1F2927',
    colors: {
        brand: [
            '#F0FDFA',
            '#DCFDF7',
            '#CCFBF1',
            '#99F6E4',
            '#5EEAD4',
            '#2DD4BF',
            '#0F766E',
            '#115E59',
            '#134E4A',
            '#0B3D3A',
        ],
        gray: [
            '#F6F8F7',
            '#EEF2F1',
            '#DDE4E2',
            '#C6D1CE',
            '#A2B0AC',
            '#7E8E89',
            '#66736F',
            '#4E5B57',
            '#36423F',
            '#1F2927',
        ],
    },
})

export function AppProviders({ children }: Props) {
    return (
        <MantineProvider theme={theme} defaultColorScheme="light">
            {children}
        </MantineProvider>
    )
}
