import type {ReactNode} from 'react'
import {AppShell, Burger, Group, NavLink, Title} from '@mantine/core'
import {useDisclosure} from '@mantine/hooks'

type Props = {
    children: ReactNode
}

export function ResidentAppShell({children}: Props) {
    const [navbarOpened, {toggle: toggleNavbar, close: closeNavbar}] =
        useDisclosure(false)

    return (
        <AppShell
            navbar={{
                width: 280,
                breakpoint: 'sm',
                collapsed: {
                    mobile: !navbarOpened,
                },
            }}
            header={{height: 60}}
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
                <NavLink
                    label="Дежурство"
                    onClick={closeNavbar}
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

            <AppShell.Main>{children}</AppShell.Main>
        </AppShell>
    )
}
