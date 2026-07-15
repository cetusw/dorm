import { ResidentAppShell } from './ResidentAppShell'
import { useCurrentUserState } from '../features/current-user/model/useCurrentUser'
import { CurrentDutyPage } from '../pages/current-duty/CurrentDutyPage'
import { DormitoriesPage } from '../pages/dormitories/DormitoriesPage'
import { LoginPage } from '../pages/login/LoginPage'
import { ResidentSectionPlaceholderPage } from '../pages/placeholder/ResidentSectionPlaceholderPage'

function renderResidentPage(pathname: string) {
    if (pathname === '/app/dormitories') {
        return <DormitoriesPage />
    }

    if (pathname === '/app/residents') {
        return <ResidentSectionPlaceholderPage title="Жители" />
    }

    if (pathname === '/app/groups') {
        return <ResidentSectionPlaceholderPage title="Группы" />
    }

    return <CurrentDutyPage />
}

export function AppRouter() {
    const pathname = window.location.pathname

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
        >
            {renderResidentPage(pathname)}
        </ResidentAppShell>
    )
}
