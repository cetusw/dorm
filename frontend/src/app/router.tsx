import { ResidentAppShell } from './ResidentAppShell'
import { CurrentDutyPage } from '../pages/current-duty/CurrentDutyPage'
import { LoginPage } from '../pages/login/LoginPage'

function renderResidentRoute(pathname: string) {
    if (pathname === '/app' || pathname === '/app/tasks') {
        return (
            <ResidentAppShell>
                <CurrentDutyPage />
            </ResidentAppShell>
        )
    }

    return (
        <ResidentAppShell>
            <CurrentDutyPage />
        </ResidentAppShell>
    )
}

export function AppRouter() {
    const pathname = window.location.pathname

    if (pathname === '/app/login') {
        return <LoginPage />
    }

    return renderResidentRoute(pathname)
}
