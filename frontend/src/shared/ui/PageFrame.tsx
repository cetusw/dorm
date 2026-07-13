import type { ReactNode } from 'react'

import { Alert, Box, Stack, Title } from '@mantine/core'

type Props = {
    title: string
    error?: string | null
    notice?: string | null
    controls?: ReactNode
    children: ReactNode
}

export function PageFrame({ title, error, notice, controls, children }: Props) {
    return (
        <Box px={{ base: 'md', md: 'xl' }} py="xl">
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

                {controls}

                {children}
            </Stack>
        </Box>
    )
}
