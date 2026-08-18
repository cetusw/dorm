import { useEffect } from 'react'
import type { RefObject } from 'react'

type Options = {
    enabled: boolean
    defer?: boolean
}

export function useOverlayAutofocus(
    rootRef: RefObject<HTMLElement | null>,
    { enabled, defer = false }: Options,
) {
    useEffect(() => {
        if (!enabled) {
            return
        }

        let timeoutID: number | null = null
        let frameID = 0

        const focusFirstField = () => {
            const root = rootRef.current
            if (!root) {
                return
            }

            const explicitTarget = root.querySelector<HTMLElement>('[data-autofocus]')
            const fallbackTarget = root.querySelector<HTMLElement>(
                'input:not([type="hidden"]):not([disabled]), textarea:not([disabled]), select:not([disabled])',
            )
            const target = explicitTarget ?? fallbackTarget

            target?.focus()
        }

        const scheduleFocus = () => {
            frameID = window.requestAnimationFrame(() => {
                frameID = window.requestAnimationFrame(focusFirstField)
            })
        }

        if (defer) {
            timeoutID = window.setTimeout(scheduleFocus, 0)
        } else {
            scheduleFocus()
        }

        return () => {
            if (timeoutID !== null) {
                window.clearTimeout(timeoutID)
            }
            window.cancelAnimationFrame(frameID)
        }
    }, [defer, enabled, rootRef])
}
