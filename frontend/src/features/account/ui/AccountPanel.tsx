import { useEffect, useMemo, useState } from 'react'

import {
    XIcon,
    CaretLeftIcon,
    ConfettiIcon,
    SignOutIcon,
    WarningIcon,
} from '@phosphor-icons/react'
import { ActionIcon, Alert, Box, Center, Loader, Stack, Text } from '@mantine/core'

import { navigateTo } from '../../../app/navigation'
import { logoutResident } from '../../auth/api/authApi'
import type { CurrentUser } from '../../current-user/model/types'
import { getCurrentUserPenalties } from '../../penalties/api/penaltiesApi'
import type { PenaltyItem } from '../../penalties/model/types'
import { formatPenaltyWeight } from '../../penalties/model/utils'
import { ResidentPenaltiesTable } from '../../penalties/ui/ResidentPenaltiesTable'
import classes from './AccountPanel.module.css'

type Props = {
    currentUser: CurrentUser
    variant: 'page' | 'drawer'
    onClose?: () => void
}

function getUserDisplayName(user: CurrentUser): string {
    const fullName = `${user.first_name} ${user.last_name}`.trim()
    return fullName !== '' ? fullName : 'Аккаунт'
}

function getUserInitial(user: CurrentUser): string {
    const source = user.first_name.trim() || user.last_name.trim() || 'A'
    return source.charAt(0).toUpperCase()
}

function getAvatarColor(user: CurrentUser): string {
    const source = `${user.first_name} ${user.last_name}`.trim() || 'Dorm User'
    let hash = 0

    for (const character of source) {
        hash = (hash * 31 + character.charCodeAt(0)) % 360
    }

    return `hsl(${hash} 78% 60%)`
}

async function handleResidentLogout() {
    try {
        await logoutResident()
    } finally {
        window.location.assign('/app/login')
    }
}

function toErrorMessage(error: unknown): string {
    if (error instanceof Error && error.message.trim() !== '') {
        return error.message
    }

    return 'Не удалось загрузить предупреждения'
}

function getPenaltyTotalWeight(penalties: PenaltyItem[]): number {
    return penalties.reduce((sum, penalty) => sum + penalty.weight, 0)
}

function getPenaltyNoun(value: number): string {
    if (!Number.isInteger(value)) {
        return 'предупреждения'
    }

    const absoluteValue = Math.abs(value)
    const lastTwoDigits = absoluteValue % 100
    const lastDigit = absoluteValue % 10

    if (lastTwoDigits >= 11 && lastTwoDigits <= 14) {
        return 'предупреждений'
    }

    if (lastDigit === 1) {
        return 'предупреждение'
    }

    if (lastDigit >= 2 && lastDigit <= 4) {
        return 'предупреждения'
    }

    return 'предупреждений'
}

function handleBack() {
    if (window.history.length > 1) {
        window.history.back()
        return
    }

    navigateTo('/app/tasks')
}

export function AccountPanel({ currentUser, variant, onClose }: Props) {
    const [penalties, setPenalties] = useState<PenaltyItem[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const avatarColor = useMemo(() => getAvatarColor(currentUser), [currentUser])
    const penaltyTotalWeight = useMemo(() => getPenaltyTotalWeight(penalties), [penalties])
    const hasPenalties = penaltyTotalWeight > 0

    useEffect(() => {
        let active = true

        async function loadPenalties() {
            setLoading(true)
            setError(null)

            try {
                const response = await getCurrentUserPenalties()
                if (!active) {
                    return
                }

                setPenalties(response.penalties)
            } catch (currentError) {
                if (!active) {
                    return
                }

                setError(toErrorMessage(currentError))
            } finally {
                if (active) {
                    setLoading(false)
                }
            }
        }

        void loadPenalties()

        return () => {
            active = false
        }
    }, [])

    const pageLayout = variant === 'page'

    return (
        <Box className={pageLayout ? classes.page : classes.drawerContent}>
            <Stack gap="lg" className={pageLayout ? classes.pageContent : undefined}>
                {pageLayout ? (
                    <button
                        type="button"
                        className={classes.backButton}
                        onClick={handleBack}
                    >
                        <CaretLeftIcon size={20} />
                        <span>Назад</span>
                    </button>
                ) : null}

                {pageLayout ? (
                    <div className={classes.pageHeader}>
                        <div className={classes.profile}>
                            <span className={classes.avatar} style={{ backgroundColor: avatarColor }}>
                                {getUserInitial(currentUser)}
                            </span>
                            <h1 className={classes.pageName}>
                                {getUserDisplayName(currentUser)}
                            </h1>
                        </div>

                        <button
                            type="button"
                            className={classes.logoutButton}
                            onClick={() => void handleResidentLogout()}
                        >
                            <SignOutIcon size={20} weight="regular" />
                            <span>Выйти</span>
                        </button>
                    </div>
                ) : (
                    <>
                        <div className={classes.drawerTopRow}>
                            <Text className={classes.drawerTitle}>Продуктивность</Text>
                            <ActionIcon
                                variant="subtle"
                                color="gray"
                                size={32}
                                aria-label="Закрыть"
                                className={classes.closeButton}
                                onClick={onClose}
                            >
                                <XIcon size={24} />
                            </ActionIcon>
                        </div>

                        <div className={classes.drawerHeader}>
                            <div className={classes.profile}>
                                <span className={classes.avatar} style={{ backgroundColor: avatarColor }}>
                                    {getUserInitial(currentUser)}
                                </span>
                                <h1 className={classes.drawerName}>
                                    {getUserDisplayName(currentUser)}
                                </h1>
                            </div>
                        </div>
                    </>
                )}

                {loading ? (
                    <Center py="xl">
                        <Loader />
                    </Center>
                ) : error ? (
                    <Alert color="red" title="Ошибка">
                        {error}
                    </Alert>
                ) : (
                    <>
                        <div
                            className={[
                                classes.notice,
                                hasPenalties ? classes.noticeWarning : classes.noticeInfo,
                            ].join(' ')}
                        >
                            <div className={classes.noticeContent}>
                                {hasPenalties ? (
                                    <WarningIcon size={32} weight="regular" />
                                ) : (
                                    <ConfettiIcon size={32} weight="regular" />
                                )}

                                <div className={classes.noticeText}>
                                    {hasPenalties
                                        ? `Внимание! У Вас ${formatPenaltyWeight(penaltyTotalWeight)} ${getPenaltyNoun(penaltyTotalWeight)}`
                                        : 'У Вас нет предупреждений'}
                                </div>
                            </div>
                        </div>

                        {hasPenalties ? (
                            <div className={pageLayout ? undefined : classes.mobilePenaltyTable}>
                                <ResidentPenaltiesTable
                                    penalties={penalties}
                                    dateLabel="Дата"
                                    minWidth={pageLayout ? 720 : 0}
                                />
                            </div>
                        ) : null}

                        {!pageLayout ? (
                            <Stack gap="sm">
                                <button
                                    type="button"
                                    className={[classes.logoutButton, classes.drawerLogoutButton].join(' ')}
                                    onClick={() => void handleResidentLogout()}
                                >
                                    <SignOutIcon size={20} weight="regular" />
                                    <span>Выйти</span>
                                </button>
                            </Stack>
                        ) : null}
                    </>
                )}
            </Stack>
        </Box>
    )
}
