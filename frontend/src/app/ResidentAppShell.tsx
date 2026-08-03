import { type ReactNode, useEffect, useRef } from 'react'

import { GearIcon } from '@phosphor-icons/react'
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
    ScrollArea,
    Select,
    Stack,
    Tooltip,
} from '@mantine/core'
import { useDisclosure } from '@mantine/hooks'

import logo from '../assets/logo.svg'
import { getSettingsReturnPath } from './navigation'
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

function isSettingsPath(pathname: string): boolean {
    return pathname === '/app/settings' || pathname === '/app/notifications'
}

export function ResidentAppShell({
    currentPath,
    currentUser,
    currentUserError,
    currentUserLoading,
    onNavigate,
    children,
}: Props) {
    const pageViewportRef = useRef<HTMLDivElement | null>(null)
    const [navbarOpened, { toggle: toggleNavbar, close: closeNavbar }] =
        useDisclosure(false)
    const hasManagementNavigation = Boolean(currentUser?.can_manage_dormitories)
    const canManagePenalties = currentUser?.can_manage_penalties === true
    const hasNavigation = hasManagementNavigation || canManagePenalties
    const {
        dormitories,
        loading: dormitoriesLoading,
        selectedDormitoryId,
        setSelectedDormitoryId,
    } = useDormitorySelection(hasManagementNavigation)

    const navigationItems: NavigationItem[] = hasManagementNavigation
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
            {
                href: '/app/penalties',
                label: 'Предупреждения',
            },
        ]
        : canManagePenalties ? [
            {
                href: '/app/tasks',
                label: 'Дежурство',
            },
            {
                href: '/app/penalties',
                label: 'Предупреждения',
            },
        ] : []

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

    const desktopDormitoryControls = hasManagementNavigation ? (
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

    const mobileDormitoryControls = hasManagementNavigation ? (
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

    function handleLogoClick() {
        closeNavbar()
        onNavigate('/app/tasks')
    }

    function handleSettingsClick() {
        closeNavbar()

        if (isSettingsPath(currentPath)) {
            onNavigate(getSettingsReturnPath())
            return
        }

        onNavigate('/app/settings')
    }

    useEffect(() => {
        pageViewportRef.current?.scrollTo({
            top: 0,
            left: 0,
            behavior: 'auto',
        })
    }, [currentPath])

    const pageContent = currentUserLoading ? (
        <Center h="100%">
            <Loader />
        </Center>
    ) : currentUserError ? (
        <Center h="100%" px="md">
            <Alert color="red" title="Ошибка">
                {currentUserError}
            </Alert>
        </Center>
    ) : (
        children
    )

    return (
        <AppShell
            style={{
                height: '100dvh',
                minHeight: '100vh',
            }}
            navbar={hasNavigation ? {
                width: 280,
                breakpoint: 'sm',
                collapsed: {
                    desktop: false,
                    mobile: !navbarOpened,
                },
            } : undefined}
            header={{ height: 60 }}
            padding={0}
            styles={{
                main: {
                    backgroundColor: 'var(--app-color-bg)',
                    height: '100dvh',
                    minHeight: '100vh',
                    overflow: 'hidden',
                    display: 'flex',
                    flexDirection: 'column',
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
            {hasNavigation ? (
                <AppShell.Navbar p="md">
                    <Stack gap="md">
                        {mobileDormitoryControls ? (
                            <Stack gap="sm" hiddenFrom="sm">
                                {mobileDormitoryControls}
                            </Stack>
                        ) : null}

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
                </AppShell.Navbar>
            ) : null}

            <AppShell.Header px={{ base: 'md', md: 'xl' }}>
                <Group align="center" h="100%" justify="space-between" wrap="nowrap">
                    <Group align="center" wrap="nowrap" gap="md">
                        {hasNavigation ? (
                            <Burger
                                opened={navbarOpened}
                                onClick={toggleNavbar}
                                hiddenFrom="sm"
                                size="sm"
                                aria-label="Открыть навигацию"
                            />
                        ) : null}

                        <Box
                            component="img"
                            src={logo}
                            alt="Dorm"
                            h={32}
                            w="auto"
                            style={{
                                display: 'block',
                                cursor: 'pointer',
                            }}
                            onClick={handleLogoClick}
                        />

                        {desktopDormitoryControls && (
                            <Group align="center" gap="sm" wrap="nowrap" visibleFrom="sm">
                                {desktopDormitoryControls}
                            </Group>
                        )}
                    </Group>

                    <Tooltip label="Настройки" withArrow>
                        <ActionIcon
                            aria-label="Открыть настройки"
                            variant="subtle"
                            color="gray"
                            size="lg"
                            radius="xl"
                            onClick={handleSettingsClick}
                        >
                            <GearIcon size={25} />
                        </ActionIcon>
                    </Tooltip>
                </Group>
            </AppShell.Header>

            <AppShell.Main>
                <ScrollArea
                    viewportRef={pageViewportRef}
                    type="auto"
                    scrollbars="y"
                    h="100%"
                    viewportProps={{
                        style: {
                            overflowX: 'hidden',
                        },
                    }}
                    styles={{
                        root: {
                            flex: 1,
                            minHeight: 0,
                        },
                        viewport: {
                            height: '100%',
                        },
                        content: {
                            minHeight: '100%',
                        },
                    }}
                >
                    <Box mih="100%">
                        {pageContent}
                    </Box>
                </ScrollArea>
            </AppShell.Main>
        </AppShell>
    )
}
