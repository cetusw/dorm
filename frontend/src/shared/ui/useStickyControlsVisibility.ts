import { useEffect, useState } from 'react'

const MOBILE_BREAKPOINT_PX = 48 * 16
const MOBILE_CONTROLS_HIDE_SCROLL_Y = 24

type Params = {
    enabled: boolean
}

function shouldHideStickyControls(): boolean {
    if (window.innerWidth >= MOBILE_BREAKPOINT_PX) {
        return false
    }

    return window.scrollY > MOBILE_CONTROLS_HIDE_SCROLL_Y
}

export function useStickyControlsVisibility({ enabled }: Params): boolean {
    const [isVisible, setIsVisible] = useState(true)

    useEffect(() => {
        if (!enabled) {
            setIsVisible(true)
            return
        }

        function syncVisibility() {
            setIsVisible(!shouldHideStickyControls())
        }

        syncVisibility()
        window.addEventListener('scroll', syncVisibility, { passive: true })
        window.addEventListener('resize', syncVisibility)

        return () => {
            window.removeEventListener('scroll', syncVisibility)
            window.removeEventListener('resize', syncVisibility)
        }
    }, [enabled])

    return isVisible
}
