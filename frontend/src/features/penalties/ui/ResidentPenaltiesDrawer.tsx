import { useCallback, useEffect, useState } from 'react'

import { RowsPlusBottomIcon } from '@phosphor-icons/react'
import { Alert, Center, Drawer, FocusTrap, Loader, Stack } from '@mantine/core'
import { Text } from '@mantine/core'

import { ApiError } from '../../../shared/api/ApiError'
import { getResidentPenalties, resolvePenalty } from '../api/penaltiesApi'
import type { PenaltyResidentDetailsResponse } from '../model/types'
import { CreatePenaltyDrawer } from './CreatePenaltyDrawer'
import { ResidentPenaltiesTable } from './ResidentPenaltiesTable'
import classes from './ResidentPenaltiesDrawer.module.css'

type Props = {
    residentId: string | null
    opened: boolean
    onClose: () => void
    onChanged: () => Promise<void> | void
}

function toErrorMessage(error: unknown): string {
    if (error instanceof ApiError) {
        if (error.status === 500) {
            return 'Не удалось выполнить запрос'
        }
        if (error.status === 403 && error.message.trim() === '') {
            return 'Недостаточно прав для выполнения действия'
        }
        return error.message
    }

    if (error instanceof Error && error.message.trim() !== '') {
        return error.message
    }

    return 'Не удалось выполнить запрос'
}

export function ResidentPenaltiesDrawer({
    residentId,
    opened,
    onClose,
    onChanged,
}: Props) {
    const [data, setData] = useState<PenaltyResidentDetailsResponse | null>(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [pendingPenaltyId, setPendingPenaltyId] = useState<string | null>(null)
    const [createDrawerOpened, setCreateDrawerOpened] = useState(false)
    const [hasChanges, setHasChanges] = useState(false)

    const resetState = useCallback(() => {
        setData(null)
        setLoading(false)
        setError(null)
        setPendingPenaltyId(null)
        setCreateDrawerOpened(false)
        setHasChanges(false)
    }, [])

    const handleClose = useCallback(() => {
        if (hasChanges) {
            void onChanged()
        }
        resetState()
        onClose()
    }, [hasChanges, onChanged, onClose, resetState])

    useEffect(() => {
        if (!opened || !residentId) {
            return
        }

        let cancelled = false
        const timeoutId = window.setTimeout(() => {
            setLoading(true)
            setError(null)
            setData(null)

            void getResidentPenalties(residentId)
                .then((response) => {
                    if (!cancelled) {
                        setData(response)
                    }
                })
                .catch(async (currentError) => {
                    if (cancelled) {
                        return
                    }

                    if (currentError instanceof ApiError && currentError.status === 404) {
                        setHasChanges(true)
                        handleClose()
                        return
                    }

                    setError(toErrorMessage(currentError))
                })
                .finally(() => {
                    if (!cancelled) {
                        setLoading(false)
                    }
                })
        }, 0)

        return () => {
            cancelled = true
            window.clearTimeout(timeoutId)
        }
    }, [handleClose, onChanged, opened, residentId])

    async function handleDelete(penaltyId: string) {
        setPendingPenaltyId(penaltyId)
        setError(null)

        try {
            await resolvePenalty(penaltyId)
            let shouldClose = false

            setData((currentData) => {
                if (currentData == null) {
                    return currentData
                }

                const penalties = currentData.penalties.filter((penalty) => penalty.id !== penaltyId)
                shouldClose = penalties.length === 0

                return {
                    ...currentData,
                    penalties,
                }
            })
            setHasChanges(true)

            if (shouldClose) {
                handleClose()
            }
        } catch (currentError) {
            if (currentError instanceof ApiError && currentError.status === 404) {
                setHasChanges(true)
                handleClose()
                return
            }

            setError(toErrorMessage(currentError))
        } finally {
            setPendingPenaltyId(null)
        }
    }

    return (
        <Drawer
            opened={opened}
            onClose={handleClose}
            position="right"
            size={760}
            title={data?.full_name ?? 'Предупреждения'}
        >
            <FocusTrap.InitialFocus />

            <Stack gap="md" className={classes.content}>
                {error ? (
                    <Alert color="red">{error}</Alert>
                ) : null}

                {loading ? (
                    <Center py="xl">
                        <Loader />
                    </Center>
                ) : data ? (
                    <ResidentPenaltiesTable
                        penalties={data.penalties}
                        pendingPenaltyId={pendingPenaltyId}
                        onDelete={handleDelete}
                    />
                ) : null}

                {!loading && data ? (
                    <div className={classes.actions}>
                        <button
                            type="button"
                            className={classes.addPenaltyLink}
                            onClick={() => setCreateDrawerOpened(true)}
                        >
                            <RowsPlusBottomIcon size={25} />
                            <Text className={classes.addPenaltyText}>Выдать предупреждение</Text>
                        </button>
                    </div>
                ) : null}
            </Stack>

            {data ? (
                <CreatePenaltyDrawer
                    opened={createDrawerOpened}
                    onClose={() => setCreateDrawerOpened(false)}
                    onCreated={async () => {
                        if (!residentId) {
                            return
                        }

                        const response = await getResidentPenalties(residentId)
                        setData(response)
                        setHasChanges(true)
                    }}
                    residentPreset={{
                        id: data.user_id,
                        name: data.full_name,
                    }}
                    lockResident
                />
            ) : null}
        </Drawer>
    )
}
