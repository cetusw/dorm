import { ResidentAppShell } from './app/ResidentAppShell'
import { LoginPage } from './pages/LoginPage/LoginPage'
import { CurrentDutyTasksPage } from './pages/CurrentDutyTasksPage'

function App() {
    if (window.location.pathname === '/app/login') {
        return <LoginPage />
    }

    if (window.location.pathname === '/app/tasks' || window.location.pathname === '/app') {
        return (
            <ResidentAppShell>
                <CurrentDutyTasksPage />
            </ResidentAppShell>
        )
    }

    return (
        <ResidentAppShell>
            <CurrentDutyTasksPage />
        </ResidentAppShell>
    )
}

export default App
