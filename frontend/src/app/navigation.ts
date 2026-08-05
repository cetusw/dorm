import { useEffect, useState } from 'react'

const APP_NAVIGATION_EVENT = 'app:navigation'

type AppHistoryState = {
    appPath?: string
}

function readAppHistoryState(): AppHistoryState {
    const { state } = window.history

    if (typeof state !== 'object' || state == null) {
        return {}
    }

    return state as AppHistoryState
}

function buildAppHistoryState(path: string): AppHistoryState {
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

export function replaceTo(path: string) {
    const currentPath = window.location.pathname + window.location.search
    if (currentPath === path) {
        return
    }

    window.history.replaceState(buildAppHistoryState(path), '', path)
    window.dispatchEvent(new CustomEvent(APP_NAVIGATION_EVENT))
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
