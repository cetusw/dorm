import { useEffect, useState } from 'react'

const APP_NAVIGATION_EVENT = 'app:navigation'
const DEFAULT_SETTINGS_RETURN_PATH = '/app/tasks'

type AppHistoryState = {
    appPath?: string
    previousAppPath?: string
}

function normalizeAppPath(path: string): string {
    const [pathname] = path.split('?')
    return pathname ?? path
}

function isSettingsPath(path: string): boolean {
    const pathname = normalizeAppPath(path)
    return pathname === '/app/settings' || pathname === '/app/notifications'
}

function readAppHistoryState(): AppHistoryState {
    const { state } = window.history

    if (typeof state !== 'object' || state == null) {
        return {}
    }

    return state as AppHistoryState
}

function buildAppHistoryState(path: string): AppHistoryState {
    const currentPath = window.location.pathname + window.location.search
    const currentState = readAppHistoryState()

    if (isSettingsPath(path)) {
        return {
            appPath: path,
            previousAppPath: isSettingsPath(currentPath)
                ? currentState.previousAppPath ?? DEFAULT_SETTINGS_RETURN_PATH
                : currentPath,
        }
    }

    return {
        appPath: path,
    }
}

export function navigateTo(path: string) {
    const currentPath = window.location.pathname + window.location.search
    if (currentPath === path) {
        return
    }

    window.history.pushState(buildAppHistoryState(path), '', path)
    window.dispatchEvent(new CustomEvent(APP_NAVIGATION_EVENT))
}

export function getSettingsReturnPath(): string {
    const previousAppPath = readAppHistoryState().previousAppPath

    if (typeof previousAppPath === 'string' && previousAppPath.startsWith('/app/')) {
        return previousAppPath
    }

    return DEFAULT_SETTINGS_RETURN_PATH
}

export function useAppPathname() {
    const [path, setPath] = useState(window.location.pathname + window.location.search)

    useEffect(() => {
        const currentState = readAppHistoryState()
        const currentPath = window.location.pathname + window.location.search
        if (currentState.appPath !== currentPath) {
            window.history.replaceState(
                {
                    ...currentState,
                    appPath: currentPath,
                } satisfies AppHistoryState,
                '',
                currentPath,
            )
        }

        function syncPath() {
            setPath(window.location.pathname + window.location.search)
        }

        window.addEventListener('popstate', syncPath)
        window.addEventListener(APP_NAVIGATION_EVENT, syncPath)

        return () => {
            window.removeEventListener('popstate', syncPath)
            window.removeEventListener(APP_NAVIGATION_EVENT, syncPath)
        }
    }, [])

    return path
}
