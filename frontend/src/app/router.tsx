import { Alert } from '@mantine/core'

import { navigateTo, useAppPathname } from './navigation'
import { ResidentAppShell } from './ResidentAppShell'
import { AreasPage } from '../pages/areas/AreasPage'
import { useCurrentUserState } from '../features/current-user/model/useCurrentUser'
import { CurrentDutyPage } from '../pages/current-duty/CurrentDutyPage'
import { DutySettingsPage } from '../pages/duty-settings/DutySettingsPage'
import { DormitoriesPage } from '../pages/dormitories/DormitoriesPage'
import { GroupsPage } from '../pages/groups/GroupsPage'
import { LoginPage } from '../pages/login/LoginPage'
import { PenaltiesPage } from '../pages/penalties/PenaltiesPage'
import { ResidentsPage } from '../pages/residents/ResidentsPage'
import { SettingsPage } from '../pages/settings/SettingsPage'
import { TaskCatalogPage } from '../pages/task-catalog/TaskCatalogPage'
import { GroupTeamsPage } from '../pages/teams/GroupTeamsPage'
import { PageFrame } from '../shared/ui/PageFrame'

function matchGroupTeamsPath(pathname: string): string | null {
    const match = pathname.match(/^\/app\/groups\/([^/]+)\/teams$/)
    return match ? match[1] : null
}

function matchDutySettingsPath(pathname: string): string | null {
    const match = pathname.match(/^\/app\/groups\/([^/]+)\/duty-settings$/)
    return match ? match[1] : null
}

function renderResidentPage(pathname: string, currentUser: ReturnType<typeof useCurrentUserState>['currentUser']) {
    const canManageDormitories = Boolean(currentUser?.can_manage_dormitories)
    const canManagePenalties = currentUser?.can_manage_penalties === true
    const dutySettingsGroupId = matchDutySettingsPath(pathname)

    if (dutySettingsGroupId) {
        return <DutySettingsPage groupId={dutySettingsGroupId} />
    }

    if (pathname === '/app/settings' || pathname === '/app/notifications') {
        return <SettingsPage />
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

    if (!canManageDormitories) {
        return <CurrentDutyPage />
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

    return <CurrentDutyPage />
}

export function AppRouter() {
    const pathname = useAppPathname()
    const { currentUser, loading, error } = useCurrentUserState(pathname !== '/app/login')

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
