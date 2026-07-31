import { useEffect, useState } from 'react'

const APP_NAVIGATION_EVENT = 'app:navigation'
const DEFAULT_SETTINGS_RETURN_PATH = '/app/tasks'

type AppHistoryState = {
    appPathname?: string
    previousAppPathname?: string
}

function isSettingsPath(pathname: string): boolean {
    return pathname === '/app/settings' || pathname === '/app/notifications'
}

function readAppHistoryState(): AppHistoryState {
    const { state } = window.history

    if (typeof state !== 'object' || state == null) {
        return {}
    }

    return state as AppHistoryState
}

function buildAppHistoryState(pathname: string): AppHistoryState {
    const currentPathname = window.location.pathname
    const currentState = readAppHistoryState()

    if (isSettingsPath(pathname)) {
        return {
            appPathname: pathname,
            previousAppPathname: isSettingsPath(currentPathname)
                ? currentState.previousAppPathname ?? DEFAULT_SETTINGS_RETURN_PATH
                : currentPathname,
        }
    }

    return {
        appPathname: pathname,
    }
}

export function navigateTo(pathname: string) {
    if (window.location.pathname === pathname) {
        return
    }

    window.history.pushState(buildAppHistoryState(pathname), '', pathname)
    window.dispatchEvent(new CustomEvent(APP_NAVIGATION_EVENT))
}

export function getSettingsReturnPath(): string {
    const previousAppPathname = readAppHistoryState().previousAppPathname

    if (typeof previousAppPathname === 'string' && previousAppPathname.startsWith('/app/')) {
        return previousAppPathname
    }

    return DEFAULT_SETTINGS_RETURN_PATH
}

export function useAppPathname() {
    const [pathname, setPathname] = useState(window.location.pathname)

    useEffect(() => {
        const currentState = readAppHistoryState()
        if (currentState.appPathname !== window.location.pathname) {
            window.history.replaceState(
                {
                    ...currentState,
                    appPathname: window.location.pathname,
                } satisfies AppHistoryState,
                '',
                window.location.pathname,
            )
        }

        function syncPathname() {
            setPathname(window.location.pathname)
        }

        window.addEventListener('popstate', syncPathname)
        window.addEventListener(APP_NAVIGATION_EVENT, syncPathname)

        return () => {
            window.removeEventListener('popstate', syncPathname)
            window.removeEventListener(APP_NAVIGATION_EVENT, syncPathname)
        }
    }, [])

    return pathname
}
