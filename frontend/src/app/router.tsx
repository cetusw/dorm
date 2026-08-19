import { useEffect } from 'react'

import { Alert } from '@mantine/core'

import { navigateTo, replaceTo, useAppPathname } from './navigation'
import { ResidentAppShell } from './ResidentAppShell'
import { AreasPage } from '../pages/areas/AreasPage'
import { AccountPage } from '../pages/account/AccountPage'
import { useCurrentUserState } from '../features/current-user/model/useCurrentUser'
import { CurrentDutyPage } from '../pages/current-duty/CurrentDutyPage'
import { DutyHistoryDetailsPage } from '../pages/duty-history/DutyHistoryDetailsPage'
import { DutyHistoryPage } from '../pages/duty-history/DutyHistoryPage'
import { DutySettingsPage } from '../pages/duty-settings/DutySettingsPage'
import { DormitoriesPage } from '../pages/dormitories/DormitoriesPage'
import { GroupsPage } from '../pages/groups/GroupsPage'
import { LoginPage } from '../pages/login/LoginPage'
import { PenaltiesPage } from '../pages/penalties/PenaltiesPage'
import { ResidentsPage } from '../pages/residents/ResidentsPage'
import { TaskCatalogPage } from '../pages/task-catalog/TaskCatalogPage'
import { GroupTeamsPage } from '../pages/teams/GroupTeamsPage'
import { WarehousePage } from '../pages/warehouse/WarehousePage'
import { PageFrame } from '../shared/ui/PageFrame'

function matchGroupTeamsPath(pathname: string): string | null {
    const match = pathname.match(/^\/app\/groups\/([^/]+)\/teams$/)
    return match ? match[1] : null
}

function matchDutySettingsPath(pathname: string): string | null {
    const match = pathname.match(/^\/app\/groups\/([^/]+)\/duty-settings$/)
    return match ? match[1] : null
}

function matchDutyHistoryDetailsPath(pathname: string): string | null {
    const match = pathname.match(/^\/app\/duties\/([^/]+)$/)
    return match ? match[1] : null
}

function renderResidentPage(path: string, currentUser: ReturnType<typeof useCurrentUserState>['currentUser']) {
    const [pathname, search = ''] = path.split('?')
    const searchParams = new URLSearchParams(search)
    const selectedGroupId = searchParams.get('group_id') ?? undefined
    const canManageDormitories = Boolean(currentUser?.can_manage_dormitories)
    const canManagePenalties = currentUser?.can_manage_penalties === true
    const canManageWarehouse = currentUser?.can_manage_warehouse === true
    const dutySettingsGroupId = matchDutySettingsPath(pathname)

    if (dutySettingsGroupId) {
        return <DutySettingsPage groupId={dutySettingsGroupId} />
    }

    if (pathname === '/app/penalties') {
        if (!canManagePenalties) {
            return (
                <PageFrame title="Предупреждения">
                    <Alert color="red">
                        Недостаточно прав для управления предупреждениями.
                    </Alert>
                </PageFrame>
            )
        }

        return <PenaltiesPage />
    }

    if (pathname === '/app/warehouse') {
        if (!canManageWarehouse) {
            return (
                <PageFrame title="Склад">
                    <Alert color="red">
                        Недостаточно прав для управления складом.
                    </Alert>
                </PageFrame>
            )
        }

        return <WarehousePage currentUser={currentUser} />
    }

    if (pathname === '/app/account') {
        if (!currentUser) {
            return (
                <PageFrame title="Аккаунт">
                    <Alert color="red">Не удалось загрузить текущего пользователя.</Alert>
                </PageFrame>
            )
        }

        return <AccountPage currentUser={currentUser} />
    }

    if (!canManageDormitories) {
        if (pathname === '/app/duties/history') {
            return <DutyHistoryPage />
        }

        const dutyId = matchDutyHistoryDetailsPath(pathname)
        if (dutyId) {
            return <DutyHistoryDetailsPage dutyId={dutyId} />
        }

        return <CurrentDutyPage currentUser={currentUser} selectedGroupId={selectedGroupId} />
    }

    if (pathname === '/app/duties/history') {
        return <DutyHistoryPage />
    }

    const dutyId = matchDutyHistoryDetailsPath(pathname)
    if (dutyId) {
        return <DutyHistoryDetailsPage dutyId={dutyId} />
    }

    const teamGroupId = matchGroupTeamsPath(pathname)
    if (teamGroupId) {
        return <GroupTeamsPage groupId={teamGroupId} />
    }

    if (pathname === '/app/dormitories') {
        return <DormitoriesPage />
    }

    if (pathname === '/app/residents') {
        return <ResidentsPage />
    }

    if (pathname === '/app/groups') {
        return <GroupsPage />
    }

    if (pathname === '/app/areas') {
        return <AreasPage />
    }

    if (pathname === '/app/task-definitions') {
        return <TaskCatalogPage />
    }

    return <CurrentDutyPage currentUser={currentUser} selectedGroupId={selectedGroupId} />
}

export function AppRouter() {
    const pathname = useAppPathname()
    const shouldRedirectToLogin =
        pathname === '/' || pathname === '/app' || pathname === '/app/'
    const shouldRedirectLegacySettings =
        pathname === '/app/settings' || pathname === '/app/notifications'
    const { currentUser, loading, error } = useCurrentUserState(
        pathname !== '/app/login' && !shouldRedirectToLogin && !shouldRedirectLegacySettings,
    )

    useEffect(() => {
        if (shouldRedirectToLogin) {
            window.location.replace('/app/login')
        }
    }, [shouldRedirectToLogin])

    useEffect(() => {
        if (shouldRedirectLegacySettings) {
            replaceTo('/app/tasks')
        }
    }, [shouldRedirectLegacySettings])

    if (shouldRedirectToLogin) {
        return null
    }

    if (shouldRedirectLegacySettings) {
        return null
    }

    if (pathname === '/app/login') {
        return <LoginPage />
    }

    return (
        <ResidentAppShell
            currentPath={pathname}
            currentUser={currentUser}
            currentUserError={error}
            currentUserLoading={loading}
            onNavigate={navigateTo}
        >
            {renderResidentPage(pathname, currentUser)}
        </ResidentAppShell>
    )
}
