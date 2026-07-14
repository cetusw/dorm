import type { ReactNode } from 'react'

import { Alert, AppShell, Center, Group, Loader, NavLink, Stack, Text } from '@mantine/core'

import type { CurrentUser } from '../features/current-user/model/types'

type Props = {
    currentPath: string
    currentUser: CurrentUser | null
    currentUserError: string | null
    currentUserLoading: boolean
    children: ReactNode
}

type NavigationItem = {
    href: string
    label: string
}

export function ResidentAppShell({
    currentPath,
    currentUser,
    currentUserError,
    currentUserLoading,
    children,
}: Props) {
    const navigationItems: NavigationItem[] = [
        {
            href: '/app/tasks',
            label: 'Дежурство',
        },
    ]

    if (currentUser?.can_manage_dormitories) {
        navigationItems.push({
            href: '/app/dormitories',
            label: 'Общежития',
        })
    }

    return (
        <AppShell
            navbar={{ width: 240, breakpoint: 0 }}
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
                    <Stack gap="md">
                        <Stack gap="xs" w="100%">
                            {navigationItems.map((item) => (
                                <NavLink
                                    key={item.href}
                                    active={currentPath === item.href}
                                    label={item.label}
                                    onClick={() => {
                                        window.location.assign(item.href)
                                    }}
                                />
                            ))}
                        </Stack>
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

            <AppShell.Main>
                {currentUserLoading ? (
                    <Center h="100vh">
                        <Loader />
                    </Center>
                ) : currentUserError ? (
                    <Center h="100vh" px="md">
                        <Alert color="red" title="Ошибка">
                            {currentUserError}
                        </Alert>
                    </Center>
                ) : (
                    children
                )}
            </AppShell.Main>
        </AppShell>
    )
}
