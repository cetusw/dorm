import type { ReactNode } from 'react'

import { Alert, Box, Group, Stack, Title } from '@mantine/core'

type Props = {
    title: string
    error?: string | null
    notice?: string | null
    titleActions?: ReactNode
    controls?: ReactNode
    children: ReactNode
}

export function PageFrame({ title, error, notice, titleActions, controls, children }: Props) {
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

                <Group justify="space-between" align="center" wrap="wrap" gap="md">
                    <Title order={1}>{title}</Title>
                    {titleActions}
                </Group>

                {controls}

                {children}
            </Stack>
        </Box>
    )
}
