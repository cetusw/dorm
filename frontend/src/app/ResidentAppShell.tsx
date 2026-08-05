import { type CSSProperties, type ReactNode, useEffect, useMemo, useRef } from 'react'

import { BellIcon, SignOutIcon } from '@phosphor-icons/react'
import {
    AppShell,
    Box,
    Burger,
    Button,
    Center,
    Loader,
    NavLink,
    Popover,
    ScrollArea,
    Select,
    Stack,
    Switch,
    Text,
    UnstyledButton,
    Alert,
} from '@mantine/core'
import { useClickOutside, useDisclosure } from '@mantine/hooks'

import { logoutResident } from '../features/auth/api/authApi'
import type { CurrentUser } from '../features/current-user/model/types'
import { useDormitorySelection } from '../features/dormitories/model/useDormitorySelection'
import { usePushNotifications } from '../features/manage-push-notifications/model/usePushNotifications'
import {
    HEADER_HEIGHT_PX,
} from '../shared/ui/mobileStickyThreshold'
import classes from './ResidentAppShell.module.css'

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
    const pathWithoutSearch = currentPath.split('?')[0] ?? currentPath

    if (href === '/app/tasks') {
        return pathWithoutSearch === '/app' || pathWithoutSearch === '/app/tasks'
    }

    if (href === '/app/groups') {
        return pathWithoutSearch === '/app/groups' || pathWithoutSearch.startsWith('/app/groups/')
    }

    return pathWithoutSearch === href
}

function getUserDisplayName(user: CurrentUser | null): string {
    if (!user) {
        return 'Аккаунт'
    }

    return user.first_name.trim() || user.last_name.trim() || 'Аккаунт'
}

function getUserInitial(user: CurrentUser | null): string {
    const source = user?.first_name.trim() || user?.last_name.trim() || 'A'
    return source.charAt(0).toUpperCase()
}

function getAvatarColor(user: CurrentUser | null): string {
    const source = `${user?.first_name ?? ''} ${user?.last_name ?? ''}`.trim() || 'Dorm User'
    let hash = 0

    for (const character of source) {
        hash = (hash * 31 + character.charCodeAt(0)) % 360
    }

    return `hsl(${hash} 78% 60%)`
}

async function handleResidentLogout() {
    try {
        await logoutResident()
    } finally {
        window.location.assign('/app/login')
    }
}

function getNotificationMessage(state: ReturnType<typeof usePushNotifications>['state'], error: string | null): string {
    switch (state) {
        case 'subscribed':
        case 'unsubscribing':
            return 'Уведомления включены'
        case 'permission-denied':
            return 'Разрешите уведомления в настройках браузера'
        case 'unsupported':
            return 'Ваш браузер не поддерживает уведомления'
        case 'insecure-context':
            return 'Для уведомлений требуется защищённое соединение'
        case 'disabled-by-server':
            return 'Уведомления временно недоступны'
        case 'installation-required':
            return 'Добавьте приложение на экран домой, чтобы включить уведомления'
        case 'error':
            return error ?? 'Не удалось обновить настройки уведомлений'
        default:
            return 'Включите уведомления, чтобы не забывать о дежурстве'
    }
}

function getNotificationChecked(state: ReturnType<typeof usePushNotifications>['state']): boolean {
    return state === 'subscribed' || state === 'unsubscribing'
}

function isNotificationToggleDisabled(state: ReturnType<typeof usePushNotifications>['state']): boolean {
    return state === 'loading'
        || state === 'subscribing'
        || state === 'unsubscribing'
        || state === 'unsupported'
        || state === 'insecure-context'
        || state === 'disabled-by-server'
        || state === 'installation-required'
        || state === 'permission-denied'
        || state === 'error'
}

type AccountButtonProps = {
    className?: string
    nameClassName?: string
    user: CurrentUser | null
}

function AccountButton({ className, nameClassName, user }: AccountButtonProps) {
    const avatarColor = useMemo(() => getAvatarColor(user), [user])
    const [opened, { toggle, close }] = useDisclosure(false)
    const rootRef = useClickOutside<HTMLDivElement>(() => close())

    return (
        <div ref={rootRef} className={classes.accountButtonWrap}>
            <UnstyledButton
                className={[classes.accountButton, className ?? ''].join(' ').trim()}
                aria-label="Открыть меню аккаунта"
                aria-expanded={opened}
                onClick={() => toggle()}
            >
                <span className={classes.avatar} style={{ backgroundColor: avatarColor }}>
                    {getUserInitial(user)}
                </span>
                <span className={[classes.accountName, nameClassName ?? ''].join(' ').trim()}>
                    {getUserDisplayName(user)}
                </span>
            </UnstyledButton>

            {opened ? (
                <div className={classes.accountDropdown}>
                    <button
                        type="button"
                        className={classes.logoutButton}
                        onClick={() => void handleResidentLogout()}
                    >
                        <SignOutIcon size={24} weight="regular" />
                        <span>Выйти</span>
                    </button>
                </div>
            ) : null}
        </div>
    )
}

function NotificationsButton() {
    const { state, error, enable, disable } = usePushNotifications()
    const checked = getNotificationChecked(state)
    const disabled = isNotificationToggleDisabled(state)
    const toggleLabel = checked ? 'Отключить уведомления' : 'Включить уведомления'

    return (
        <Popover position="bottom-end" offset={14} shadow="none" withArrow={false}>
            <Popover.Target>
                <UnstyledButton
                    className={classes.iconButton}
                    aria-label="Открыть настройки уведомлений"
                >
                    <BellIcon size={32} weight="regular" />
                </UnstyledButton>
            </Popover.Target>

            <Popover.Dropdown className={[classes.dropdown, classes.notificationDropdown].join(' ')}>
                <Stack gap={24}>
                    <Text className={classes.notificationMessage}>
                        {getNotificationMessage(state, error)}
                    </Text>

                    <div className={classes.notificationToggleRow}>
                        <Switch
                            className={classes.notificationSwitch}
                            checked={checked}
                            disabled={disabled}
                            withThumbIndicator={false}
                            aria-label={toggleLabel}
                            onChange={(event) => {
                                if (event.currentTarget.checked) {
                                    void enable()
                                    return
                                }

                                void disable()
                            }}
                            styles={{
                                root: {
                                    '--switch-height': '20px',
                                    '--switch-width': '40px',
                                    '--switch-thumb-size': '16px',
                                    '--switch-track-label-padding': '2px',
                                    '--switch-radius': '999px',
                                } as CSSProperties,
                                track: {
                                    backgroundColor: checked ? '#0F766E' : '#DDE4E2',
                                    borderColor: checked ? '#0F766E' : '#DDE4E2',
                                    height: 20,
                                    minHeight: 20,
                                },
                                thumb: {
                                    width: 16,
                                    minWidth: 16,
                                    height: 16,
                                    borderWidth: 0,
                                },
                            }}
                        />

                        <Text
                            className={[
                                classes.notificationToggleText,
                                disabled ? classes.notificationToggleTextDisabled : '',
                            ].join(' ').trim()}
                        >
                            {toggleLabel}
                        </Text>
                    </div>
                </Stack>
            </Popover.Dropdown>
        </Popover>
    )
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
    const showShellControls = currentUser != null && !currentUserLoading && currentUserError == null
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

    const sidebarDormitoryControls = hasManagementNavigation ? (
        <Stack gap="sm" className={classes.dormitoryControls}>
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
                    top: 0,
                    height: '100dvh',
                },
                header: {
                    backgroundColor: 'transparent',
                    borderBottom: 'none',
                    backdropFilter: 'none',
                },
            }}
        >
            {hasNavigation ? (
                <AppShell.Navbar p={0} className={classes.navbar}>
                    <Stack className={classes.navbarContent}>
                        {showShellControls ? (
                            <AccountButton
                                className={classes.accountButtonDark}
                                user={currentUser}
                            />
                        ) : null}

                        {sidebarDormitoryControls}

                        <div className={classes.navigationList}>
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
                        </div>
                    </Stack>
                </AppShell.Navbar>
            ) : null}

            <AppShell.Header px={{ base: 'md', md: 'xl' }} className={classes.header}>
                <div className={classes.headerInner}>
                    <div className={classes.headerLeft}>
                        {hasNavigation ? (
                            <>
                                <Burger
                                    opened={navbarOpened}
                                    onClick={toggleNavbar}
                                    hiddenFrom="sm"
                                    size="sm"
                                    aria-label="Открыть навигацию"
                                />

                                {showShellControls ? (
                                    <AccountButton
                                        className={classes.mobileHeaderAccount}
                                        nameClassName={classes.desktopAccountName}
                                        user={currentUser}
                                    />
                                ) : null}
                            </>
                        ) : showShellControls ? (
                            <AccountButton
                                nameClassName={classes.desktopAccountName}
                                user={currentUser}
                            />
                        ) : null}
                    </div>

                    {showShellControls ? <NotificationsButton /> : null}
                </div>
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
