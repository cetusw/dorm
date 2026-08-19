import { createContext, useContext, useEffect, type ReactNode } from 'react'

type SetMobileHeaderContent = (content: ReactNode) => void

export const MobileHeaderContentContext = createContext<SetMobileHeaderContent | null>(null)

export function useMobileHeaderContent(content: ReactNode) {
    const setMobileHeaderContent = useContext(MobileHeaderContentContext)

    useEffect(() => {
        if (!setMobileHeaderContent) {
            return
        }

        setMobileHeaderContent(content)

        return () => {
            setMobileHeaderContent(null)
        }
    }, [content, setMobileHeaderContent])
}
