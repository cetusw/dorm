import { useCallback, useEffect, useState } from 'react'

import { RowsPlusBottomIcon, ScalesIcon } from '@phosphor-icons/react'
import { Alert, Center, Drawer, FocusTrap, Loader, Stack, Tooltip } from '@mantine/core'
import { Text } from '@mantine/core'

import { ApiError } from '../../../shared/api/ApiError'
import { deletePenaltyEntry, getResidentPenalties } from '../api/penaltiesApi'
import { formatPenaltyWeight } from '../model/utils'
import type { PenaltyEntryItem, PenaltyResidentDetailsResponse } from '../model/types'
import { CreatePenaltyDrawer } from './CreatePenaltyDrawer'
import { ResidentPenaltiesTable } from './ResidentPenaltiesTable'
import classes from './ResidentPenaltiesDrawer.module.css'

type Props = {
    residentId: string | null
    opened: boolean
    onClose: () => void
    onChanged: () => Promise<void> | void
}

type ActionModalState =
    | { mode: 'create'; entryType: 'issue' | 'resolve'; entry: null }
    | { mode: 'edit'; entryType: 'issue' | 'resolve'; entry: PenaltyEntryItem }
    | null

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
    const [actionModalState, setActionModalState] = useState<ActionModalState>(null)
    const [pendingEntryId, setPendingEntryId] = useState<string | null>(null)

    const reloadResidentPenalties = useCallback(async () => {
        if (!residentId) {
            return
        }

        const response = await getResidentPenalties(residentId)
        setData(response)
    }, [residentId])

    const resetState = useCallback(() => {
        setData(null)
        setLoading(false)
        setError(null)
        setActionModalState(null)
        setPendingEntryId(null)
    }, [])

    const handleClose = useCallback(() => {
        resetState()
        onClose()
    }, [onClose, resetState])

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
                .catch((currentError) => {
                    if (!cancelled) {
                        setError(toErrorMessage(currentError))
                    }
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
    }, [opened, residentId])

    async function handleDelete(entry: PenaltyEntryItem) {
        setPendingEntryId(entry.id)
        setError(null)

        try {
            await deletePenaltyEntry(entry.id)
            await reloadResidentPenalties()
            await onChanged()
        } catch (currentError) {
            setError(toErrorMessage(currentError))
        } finally {
            setPendingEntryId(null)
        }
    }

    function handleEdit(entry: PenaltyEntryItem) {
        setActionModalState({
            mode: 'edit',
            entryType: entry.type,
            entry,
        })
    }

    return (
        <Drawer
            opened={opened}
            onClose={handleClose}
            position="right"
            size={760}
            title={data?.full_name ?? 'Предупреждения'}
            closeOnEscape={actionModalState === null}
        >
            <FocusTrap.InitialFocus />

            <Stack gap={0} className={classes.content}>
                {error ? (
                    <Alert color="red" mb="md">{error}</Alert>
                ) : null}

                {loading ? (
                    <Center py="xl">
                        <Loader />
                    </Center>
                ) : data ? (
                    <>
                        <Text
                            mb="md"
                            className={[
                                classes.balance,
                                data.total_weight >= 6 ? classes.balanceCritical : '',
                            ].join(' ').trim()}
                        >
                            Сумма предупреждений: {formatPenaltyWeight(data.total_weight)}
                        </Text>
                        <ResidentPenaltiesTable
                            entries={data.entries}
                            pendingEntryId={pendingEntryId}
                            onDelete={handleDelete}
                            onEdit={handleEdit}
                        />
                    </>
                ) : null}

                {!loading && data ? (
                    <div className={classes.actions}>
                        <button
                            type="button"
                            className={classes.addPenaltyLink}
                            onClick={() => {
                                setActionModalState({
                                    mode: 'create',
                                    entryType: 'issue',
                                    entry: null,
                                })
                            }}
                        >
                            <RowsPlusBottomIcon size={25} />
                            <Text className={classes.addPenaltyText}>Выдать предупреждение</Text>
                        </button>

                        <Tooltip
                            label="У жителя погашены все предупреждения"
                            disabled={data.total_weight > 0}
                            withArrow
                        >
                            <span className={classes.resolveAction}>
                                <button
                                    type="button"
                                    className={classes.addPenaltyLink}
                                    onClick={() => {
                                        setActionModalState({
                                            mode: 'create',
                                            entryType: 'resolve',
                                            entry: null,
                                        })
                                    }}
                                    disabled={data.total_weight <= 0}
                                >
                                    <ScalesIcon size={25} />
                                    <Text className={classes.addPenaltyText}>Погасить предупреждение</Text>
                                </button>
                            </span>
                        </Tooltip>
                    </div>
                ) : null}
            </Stack>

            {data ? (
                <CreatePenaltyDrawer
                    opened={actionModalState !== null}
                    onClose={() => {
                        setActionModalState(null)
                    }}
                    onCreated={async () => {
                        await reloadResidentPenalties()
                        await onChanged()
                    }}
                    residentPreset={{
                        id: data.user_id,
                        name: data.full_name,
                    }}
                    lockResident={actionModalState?.mode === 'edit'}
                    mode={actionModalState?.mode ?? 'create'}
                    entryType={actionModalState?.entryType ?? 'issue'}
                    entryId={actionModalState?.entry?.id ?? null}
                    maxWeight={data.total_weight}
                    initialWeight={actionModalState?.entry?.weight ?? null}
                    initialReason={actionModalState?.entry?.reason ?? ''}
                />
            ) : null}
        </Drawer>
    )
}
