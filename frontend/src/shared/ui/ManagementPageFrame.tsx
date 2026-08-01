import type { ReactNode } from 'react'

import { Box, Group, Stack, Title } from '@mantine/core'

type Props = {
    title: string
    titleActions?: ReactNode
    children: ReactNode
}

export function ManagementPageFrame({ title, titleActions, children }: Props) {
    return (
        <Box px={{ base: 'md', md: 'xl' }} py="xl">
            <Stack gap="lg" maw={1240} mx="auto">
                <Group justify="space-between" align="center" wrap="wrap" gap="md">
                    <Title order={1}>{title}</Title>
                    {titleActions}
                </Group>

                {children}
            </Stack>
        </Box>
    )
}
