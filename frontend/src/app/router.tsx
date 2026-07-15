import { navigateTo, useAppPathname } from './navigation'
import { ResidentAppShell } from './ResidentAppShell'
import { useCurrentUserState } from '../features/current-user/model/useCurrentUser'
import { CurrentDutyPage } from '../pages/current-duty/CurrentDutyPage'
import { DormitoriesPage } from '../pages/dormitories/DormitoriesPage'
import { GroupsPage } from '../pages/groups/GroupsPage'
import { LoginPage } from '../pages/login/LoginPage'
import { ResidentsPage } from '../pages/residents/ResidentsPage'

function renderResidentPage(pathname: string) {
    if (pathname === '/app/dormitories') {
        return <DormitoriesPage />
    }

    if (pathname === '/app/residents') {
        return <ResidentsPage />
    }

    if (pathname === '/app/groups') {
        return <GroupsPage />
    }

    return <CurrentDutyPage />
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
            {renderResidentPage(pathname)}
        </ResidentAppShell>
    )
}
