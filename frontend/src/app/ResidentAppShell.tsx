import type { ReactNode } from 'react'

import { Alert, AppShell, Burger, Center, Group, Loader, NavLink, Stack, Title } from '@mantine/core'
import { useDisclosure } from '@mantine/hooks'

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
    const [navbarOpened, { toggle: toggleNavbar, close: closeNavbar }] =
        useDisclosure(false)

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
            navbar={{
                width: 280,
                breakpoint: 'sm',
                collapsed: {
                    mobile: !navbarOpened,
                },
            }}
            header={{ height: 60 }}
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
                <Stack gap="xs">
                    {navigationItems.map((item) => (
                        <NavLink
                            key={item.href}
                            active={currentPath === item.href}
                            label={item.label}
                            onClick={() => {
                                closeNavbar()
                                window.location.assign(item.href)
                            }}
                            styles={{
                                root: {
                                    color: '#f9fafb',
                                    borderRadius: 8,
                                },
                                label: {
                                    color: '#f9fafb',
                                },
                            }}
                        />
                    ))}
                </Stack>
            </AppShell.Navbar>

            <AppShell.Header px="xl">
                <Group align="center" h="100%">
                    <Burger
                        opened={navbarOpened}
                        onClick={toggleNavbar}
                        hiddenFrom="sm"
                        size="sm"
                        aria-label="Открыть навигацию"
                    />

                    <Title order={3}>
                        Dorm
                    </Title>
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
