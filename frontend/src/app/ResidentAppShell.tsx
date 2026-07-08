import type {ReactNode} from 'react'
import {AppShell, Group, Stack, Text, ThemeIcon} from '@mantine/core'

type Props = {
    children: ReactNode
}

export function ResidentAppShell({children}: Props) {
    return (
        <AppShell
            navbar={{width: 88, breakpoint: 0}}
            header={{height: 72}}
            padding={0}
            styles={{
                main: {
                    backgroundColor: '#f5f7fb',
                    minHeight: '100vh',
                },
                navbar: {
                    backgroundColor: '#111827',
                    borderRight: '1px solid #1f2937',
                },
                header: {
                    backgroundColor: '#ffffff',
                    borderBottom: '1px solid #e5e7eb',
                },
            }}
        >
            <AppShell.Navbar p="md">
                <Stack justify="space-between" h="100%">
                    <Stack align="center" gap="md">
                        <ThemeIcon size={48} radius="md" color="blue">
                            Д
                        </ThemeIcon>
                    </Stack>
                </Stack>
            </AppShell.Navbar>

            <AppShell.Header px="xl">
                <Group align="center" h="100%">
                    <Text size="lg" c="dimmed">
                        Dorm
                    </Text>
                </Group>
            </AppShell.Header>

            <AppShell.Main>{children}</AppShell.Main>
        </AppShell>
    )
}
