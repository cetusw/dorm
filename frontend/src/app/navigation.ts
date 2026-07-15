import { useEffect, useState } from 'react'

const APP_NAVIGATION_EVENT = 'app:navigation'

export function navigateTo(pathname: string) {
    if (window.location.pathname === pathname) {
        return
    }

    window.history.pushState(null, '', pathname)
    window.dispatchEvent(new CustomEvent(APP_NAVIGATION_EVENT))
}

export function useAppPathname() {
    const [pathname, setPathname] = useState(window.location.pathname)

    useEffect(() => {
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
