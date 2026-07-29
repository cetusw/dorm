import { type ReactNode } from 'react'

import { BellIcon } from '@phosphor-icons/react'
import {
    Alert,
    AppShell,
    Box,
    Burger,
    Button,
    Center,
    Group,
    Loader,
    NavLink,
    ActionIcon,
    Select,
    Stack,
    Tooltip,
} from '@mantine/core'
import { useDisclosure } from '@mantine/hooks'

import logo from '../assets/logo.svg'
import { logoutResident } from '../features/auth/api/authApi'
import type { CurrentUser } from '../features/current-user/model/types'
import { useDormitorySelection } from '../features/dormitories/model/useDormitorySelection'
import {
    HEADER_HEIGHT_PX,
} from '../shared/ui/mobileStickyThreshold'

type Props = {
    currentPath: string
    currentUser: CurrentUser | null
    currentUserError: string | null
    currentUserLoading: boolean
    onNavigate: (pathname: string) => void
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

    if (href === '/app/groups') {
        return currentPath === '/app/groups' || currentPath.startsWith('/app/groups/')
    }

    return currentPath === href
}

export function ResidentAppShell({
    currentPath,
    currentUser,
    currentUserError,
    currentUserLoading,
    onNavigate,
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

    const navigationItems: NavigationItem[] = currentUser?.can_manage_dormitories
        ? [
            {
                href: '/app/residents',
                label: 'Жители',
            },
            {
                href: '/app/groups',
                label: 'Группы',
            },
            {
                href: '/app/areas',
                label: 'Территории',
            },
            {
                href: '/app/task-definitions',
                label: 'Задачи',
            },
            {
                href: '/app/tasks',
                label: 'Дежурство',
            },
        ]
        : [
            {
                href: '/app/tasks',
                label: 'Дежурство',
            },
        ]

    const dormitoryOptions = dormitories.map((dormitory) => ({
        value: String(dormitory.id),
        label: dormitory.name,
    }))

    function getNavigationItemStyles(active: boolean) {
        return {
            root: {
                color: '#f9fafb',
                borderRadius: 8,
                backgroundColor: active
                    ? 'rgba(204, 251, 241, 0.16)'
                    : 'transparent',
                border: active
                    ? '1px solid rgba(204, 251, 241, 0.28)'
                    : '1px solid transparent',
            },
            label: {
                color: '#f9fafb',
                fontWeight: 500,
            },
        }
    }

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
                onClick={() => onNavigate('/app/dormitories')}
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
                    onNavigate('/app/dormitories')
                }}
            >
                Управление общежитиями
            </Button>
        </Stack>
    ) : null

    async function handleLogout() {
        try {
            await logoutResident()
        } finally {
            window.location.assign('/app/login')
        }
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
                    backgroundColor: 'var(--app-color-bg)',
                    minHeight: '100vh',
                    '--app-shell-header-offset': `${HEADER_HEIGHT_PX}px`,
                },
                navbar: {
                    backgroundColor: '#1F2927',
                    borderRight: '1px solid #31403D',
                },
                header: {
                    backgroundColor: 'var(--app-color-surface)',
                    borderBottom: '1px solid var(--app-color-border)',
                },
            }}
        >
            <AppShell.Navbar p="md">
                <Stack justify="space-between" h="100%">
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
                                    onNavigate(item.href)
                                }}
                                styles={getNavigationItemStyles(
                                    isCurrentPathActive(currentPath, item.href),
                                )}
                            />
                        ))}
                    </Stack>

                    <NavLink
                        label="Выйти"
                        onClick={() => void handleLogout()}
                        styles={getNavigationItemStyles(false)}
                    />
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

                        <Box
                            component="img"
                            src={logo}
                            alt="Dorm"
                            h={32}
                            w="auto"
                            style={{ display: 'block' }}
                        />

                        {desktopDormitoryControls && (
                            <Group align="center" gap="sm" wrap="nowrap" visibleFrom="sm">
                                {desktopDormitoryControls}
                            </Group>
                        )}
                    </Group>

                    <Tooltip label="Уведомления" withArrow>
                        <ActionIcon
                            aria-label="Открыть уведомления"
                            variant="subtle"
                            color="gray"
                            size="lg"
                            radius="xl"
                            onClick={() => onNavigate('/app/notifications')}
                        >
                            <BellIcon size={25} />
                        </ActionIcon>
                    </Tooltip>
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
