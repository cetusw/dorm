import type {ReactNode} from 'react'

import {Alert, Box, Stack, Title} from '@mantine/core'

type Props = {
    title: string
    subtitle?: string
    error?: string | null
    notice?: string | null
    controls?: ReactNode
    children: ReactNode
}

export function PageFrame({title, subtitle, error, notice, controls, children}: Props) {
    return (
        <Box px={{base: 'md', md: 'xl'}} py="xl">
            <Stack gap="lg" maw={1240} mx="auto">
                {error && (
                    <Alert color="red" title="Ошибка">
                        {error}
                    </Alert>
                )}

                {notice && (
                    <Alert color="gray">{notice}</Alert>
                )}

                <Title order={1}>{title}</Title>
                <Title order={2}>{subtitle}</Title>

                {controls}

                {children}
            </Stack>
        </Box>
    )
}
