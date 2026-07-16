import { navigateTo, useAppPathname } from './navigation'
import { ResidentAppShell } from './ResidentAppShell'
import { AreasPage } from '../pages/areas/AreasPage'
import { useCurrentUserState } from '../features/current-user/model/useCurrentUser'
import { CurrentDutyPage } from '../pages/current-duty/CurrentDutyPage'
import { DormitoriesPage } from '../pages/dormitories/DormitoriesPage'
import { GroupsPage } from '../pages/groups/GroupsPage'
import { LoginPage } from '../pages/login/LoginPage'
import { ResidentsPage } from '../pages/residents/ResidentsPage'
import { TaskCatalogPage } from '../pages/task-catalog/TaskCatalogPage'
import { GroupTeamsPage } from '../pages/teams/GroupTeamsPage'

function matchGroupTeamsPath(pathname: string): string | null {
    const match = pathname.match(/^\/app\/groups\/([^/]+)\/teams$/)
    return match ? match[1] : null
}

function renderResidentPage(pathname: string, currentUser: ReturnType<typeof useCurrentUserState>['currentUser']) {
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

    return <CurrentDutyPage currentUser={currentUser} />
}

export function AppRouter() {
    const pathname = useAppPathname()

    if (pathname === '/app/login') {
        return <LoginPage />
    }

    const { currentUser, loading, error } = useCurrentUserState()

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
