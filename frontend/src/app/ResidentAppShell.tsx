import type { ReactNode } from 'react'

import {
    Alert,
    AppShell,
    Burger,
    Button,
    Center,
    Group,
    Loader,
    NavLink,
    Select,
    Stack,
    Title,
} from '@mantine/core'
import { useDisclosure } from '@mantine/hooks'

import type { CurrentUser } from '../features/current-user/model/types'
import { useDormitorySelection } from '../features/dormitories/model/useDormitorySelection'

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

function isCurrentPathActive(currentPath: string, href: string): boolean {
    if (href === '/app/tasks') {
        return currentPath === '/app' || currentPath === '/app/tasks'
    }

    return currentPath === href
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
    const {
        dormitories,
        loading: dormitoriesLoading,
        selectedDormitoryId,
        setSelectedDormitoryId,
    } = useDormitorySelection(Boolean(currentUser?.can_manage_dormitories))

    const navigationItems: NavigationItem[] = [
        {
            href: '/app/residents',
            label: 'Жители',
        },
        {
            href: '/app/groups',
            label: 'Группы',
        },
        {
            href: '/app/tasks',
            label: 'Дежурство',
        },
    ]

    const dormitoryOptions = dormitories.map((dormitory) => ({
        value: String(dormitory.id),
        label: dormitory.name,
    }))

    const desktopDormitoryControls = currentUser?.can_manage_dormitories ? (
        <Group align="center" gap="sm" wrap="nowrap">
            <Select
                aria-label="Общежитие"
                placeholder="Общежитие"
                data={dormitoryOptions}
                value={selectedDormitoryId}
                onChange={(value) => setSelectedDormitoryId(value)}
                allowDeselect={false}
                disabled={dormitories.length === 0}
                w={{ base: '100%', sm: 240 }}
                loading={dormitoriesLoading}
            />

            <Button
                variant="default"
                onClick={() => window.location.assign('/app/dormitories')}
            >
                Управление общежитиями
            </Button>
        </Group>
    ) : null

    const mobileDormitoryControls = currentUser?.can_manage_dormitories ? (
        <Stack gap="sm">
            <Select
                aria-label="Общежитие"
                placeholder="Общежитие"
                data={dormitoryOptions}
                value={selectedDormitoryId}
                onChange={(value) => setSelectedDormitoryId(value)}
                allowDeselect={false}
                disabled={dormitories.length === 0}
                comboboxProps={{ withinPortal: false }}
                loading={dormitoriesLoading}
            />

            <Button
                variant="default"
                onClick={() => {
                    closeNavbar()
                    window.location.assign('/app/dormitories')
                }}
            >
                Управление общежитиями
            </Button>
        </Stack>
    ) : null

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
                <Stack gap="md">
                    {mobileDormitoryControls && (
                        <Stack gap="sm" hiddenFrom="sm">
                            {mobileDormitoryControls}
                        </Stack>
                    )}

                    {navigationItems.map((item) => (
                        <NavLink
                            key={item.href}
                            active={isCurrentPathActive(currentPath, item.href)}
                            label={item.label}
                            onClick={() => {
                                closeNavbar()
                                window.location.assign(item.href)
                            }}
                            styles={{
                                root: {
                                    color: '#f9fafb',
                                    borderRadius: 8,
                                    backgroundColor: isCurrentPathActive(currentPath, item.href)
                                        ? '#1f2937'
                                        : 'transparent',
                                },
                                label: {
                                    color: '#f9fafb',
                                    fontWeight: 500,
                                },
                            }}
                        />
                    ))}
                </Stack>
            </AppShell.Navbar>

            <AppShell.Header px={{ base: 'md', md: 'xl' }}>
                <Group align="center" h="100%" justify="space-between" wrap="nowrap">
                    <Group align="center" wrap="nowrap" gap="md">
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

                        {desktopDormitoryControls && (
                            <Group align="center" gap="sm" wrap="nowrap" visibleFrom="sm">
                                {desktopDormitoryControls}
                            </Group>
                        )}
                    </Group>
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
