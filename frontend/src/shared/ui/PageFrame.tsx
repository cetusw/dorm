import { useEffect, useRef, useState, type CSSProperties, type ReactNode } from 'react'

import { InfoIcon, WarningIcon } from '@phosphor-icons/react'
import { Alert, Box, Group, Stack, Title } from '@mantine/core'
import {
    MOBILE_BREAKPOINT_PX,
    SCROLL_DELTA_THRESHOLD_PX,
} from './mobileStickyThreshold'

type Props = {
    topContent?: ReactNode
    title: string
    titleStyle?: CSSProperties
    subtitle?: string
    titleActions?: ReactNode
    error?: string | null
    notice?: string | null
    noticeTone?: 'info' | 'warning' | ''
    controls?: ReactNode
    analytics?: ReactNode
    children: ReactNode
}

export function PageFrame({
    topContent,
    title,
    titleStyle,
    subtitle,
    titleActions,
    error,
    notice,
    noticeTone,
    controls,
    analytics,
    children,
}: Props) {
    const controlsBlockRef = useRef<HTMLDivElement | null>(null)
    const [isStickyCloneVisible, setIsStickyCloneVisible] = useState(false)
    const [isScrollingUp, setIsScrollingUp] = useState(false)

    useEffect(() => {
        if (!analytics && !controls) {
            setIsScrollingUp(false)
            return
        }

        let previousScrollY = window.scrollY

        function syncScrollDirection() {
            if (window.innerWidth >= MOBILE_BREAKPOINT_PX) {
                setIsScrollingUp(false)
                previousScrollY = window.scrollY
                return
            }

            const currentScrollY = Math.max(0, window.scrollY)
            const scrollDelta = currentScrollY - previousScrollY

            if (Math.abs(scrollDelta) < SCROLL_DELTA_THRESHOLD_PX) {
                return
            }

            setIsScrollingUp(scrollDelta < 0)
            previousScrollY = currentScrollY
        }

        syncScrollDirection()
        window.addEventListener('scroll', syncScrollDirection, { passive: true })
        window.addEventListener('resize', syncScrollDirection)

        return () => {
            window.removeEventListener('scroll', syncScrollDirection)
            window.removeEventListener('resize', syncScrollDirection)
        }
    }, [analytics, controls])

    useEffect(() => {
        if (!analytics && !controls) {
            setIsStickyCloneVisible(false)
            return
        }

        function readHeaderOffset(): number {
            const mainElement = document.querySelector('main')
            if (!(mainElement instanceof HTMLElement)) {
                return 0
            }

            const rawOffset = getComputedStyle(mainElement)
                .getPropertyValue('--app-shell-header-offset')
                .trim()
            const parsedOffset = Number.parseFloat(rawOffset)

            return Number.isFinite(parsedOffset) ? parsedOffset : 0
        }

        function syncStickyCloneVisibility() {
            if (window.innerWidth >= MOBILE_BREAKPOINT_PX) {
                setIsStickyCloneVisible(false)
                return
            }

            const block = controlsBlockRef.current
            if (!block) {
                setIsStickyCloneVisible(false)
                return
            }

            const headerOffset = readHeaderOffset()
            const rect = block.getBoundingClientRect()
            setIsStickyCloneVisible(rect.top < headerOffset)
        }

        syncStickyCloneVisibility()
        window.addEventListener('scroll', syncStickyCloneVisibility, { passive: true })
        window.addEventListener('resize', syncStickyCloneVisibility)

        return () => {
            window.removeEventListener('scroll', syncStickyCloneVisibility)
            window.removeEventListener('resize', syncStickyCloneVisibility)
        }
    }, [analytics, controls])

    function renderStickyContent() {
        return (
            <Stack
                gap="lg"
                py="sm"
                bg="var(--app-color-bg)"
            >
                {analytics}
                {controls}
            </Stack>
        )
    }

    const showStickyClone = isScrollingUp && isStickyCloneVisible
    const resolvedNoticeTone = noticeTone === 'warning' ? 'warning' : 'info'
    const noticeColor = resolvedNoticeTone === 'warning'
        ? { backgroundColor: '#FEF3C7', color: '#92400E' }
        : { backgroundColor: '#E0F2FE', color: '#0369A1' }
    const NoticeIcon = resolvedNoticeTone === 'warning' ? WarningIcon : InfoIcon

    return (
        <Box px={{ base: 'md', md: 'xl' }} py={{ base: 16, md: 'xl' }}>
            <Stack gap="lg" maw={1240} mx="auto">
                {error && (
                    <Alert color="red" title="Ошибка">
                        {error}
                    </Alert>
                )}

                {notice && (
                    <Box
                        style={{
                            minHeight: 80,
                            borderRadius: 16,
                            display: 'flex',
                            alignItems: 'center',
                            padding: 24,
                            ...noticeColor,
                        }}
                    >
                        <Group gap={15} wrap="nowrap">
                            <Box
                                style={{
                                    width: 32,
                                    minWidth: 32,
                                    height: 32,
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center',
                                    flexShrink: 0,
                                }}
                            >
                                <NoticeIcon size={32} weight="regular" />
                            </Box>
                            <Box
                                style={{
                                    fontSize: 16,
                                    lineHeight: '28px',
                                }}
                            >
                                {notice}
                            </Box>
                        </Group>
                    </Box>
                )}

                {topContent}

                <Group justify="space-between" align="flex-start" gap="md" wrap="wrap">
                    <Title order={1} style={titleStyle}>
                        <Box component="span">
                            {title}
                        </Box>

                        {subtitle && (
                            <>
                                <Box
                                    component="span"
                                    visibleFrom="sm"
                                >
                                    {' · '}
                                </Box>

                                <Box
                                    component="span"
                                    display={{ base: 'block', sm: 'inline' }}
                                    mt={{ base: 4, sm: 0 }}
                                    style={{
                                        whiteSpace: 'nowrap',
                                    }}
                                >
                                    {subtitle}
                                </Box>
                            </>
                        )}
                    </Title>

                    {titleActions}
                </Group>

                {showStickyClone && (
                    <Box
                        pos="fixed"
                        top="var(--app-shell-header-offset, 0px)"
                        style={{
                            zIndex: 10,
                            left: '50%',
                            width: 'min(1240px, calc(100vw - 32px))',
                            transform: 'translateX(-50%)',
                            transition: 'top 160ms ease, opacity 160ms ease',
                        }}
                    >
                        {renderStickyContent()}
                    </Box>
                )}

                {analytics || controls ? (
                    <Stack gap={0}>
                        <Box ref={controlsBlockRef}>
                            {renderStickyContent()}
                        </Box>

                        <Box data-sticky-hide-start h={0} />

                        {children}
                    </Stack>
                ) : (
                    children
                )}
            </Stack>
        </Box>
    )
}
