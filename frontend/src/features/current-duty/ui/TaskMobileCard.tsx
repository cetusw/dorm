import { useEffect, useRef, useState } from 'react'

import { Badge, Button, Checkbox, Paper, Stack, Text } from '@mantine/core'

import type { ResidentDutyTask } from '../model/types'
import { getStatusBadgeConfig } from './TaskStatusBadge'
import type { TaskRowActionMode } from './TaskRowActions'
import classes from './TaskMobileCard.module.css'

type Props = {
    activeSwipeTaskId: string | null
    isReadOnly?: boolean
    mode?: TaskRowActionMode
    pending: boolean
    task: ResidentDutyTask
    onTake: (taskId: string) => void | Promise<unknown>
    onReturn: (taskId: string) => void | Promise<unknown>
    onComplete: (taskId: string) => void | Promise<unknown>
    onOpen: (taskId: string) => void | Promise<unknown>
    onReopen?: (taskId: string) => void | Promise<unknown>
    onSwipeActiveChange: (taskId: string | null) => void
    onVerify?: (taskId: string) => void | Promise<unknown>
}

type StatusConfig = {
    backgroundColor: string
    label: string
}

type SwipeAction = {
    kind: 'danger' | 'default'
    label: string
    onClick: (taskId: string) => void | Promise<unknown>
}

const SWIPE_GAP = 6
const DEFAULT_SWIPE_ACTION_WIDTH = 128
const DANGER_SWIPE_ACTION_WIDTH = 136

function getMobileStatusConfig(task: ResidentDutyTask): StatusConfig | null {
    if (task.status === 'free') {
        return null
    }

    const config = getStatusBadgeConfig(task)

    if (!config) {
        return null
    }

    return {
        backgroundColor: config.backgroundColor,
        label: config.label,
    }
}

function MobileTaskStatus({ task }: { task: ResidentDutyTask }) {
    const statusConfig = getMobileStatusConfig(task)
    if (!statusConfig) {
        return null
    }

    return (
        <Badge
            radius="sm"
            variant="filled"
            styles={{
                root: {
                    backgroundColor: statusConfig.backgroundColor,
                    color: 'var(--app-color-text)',
                    fontWeight: 500,
                },
            }}
        >
            {statusConfig.label}
        </Badge>
    )
}

function getAssigneeLabel(task: ResidentDutyTask): string | null {
    if (!task.assignee_name) {
        return null
    }

    return task.is_mine ? 'Вы' : task.assignee_name
}

function getLeftSwipeAction(
    mode: TaskRowActionMode,
    task: ResidentDutyTask,
    isReadOnly: boolean,
    onReturn: (taskId: string) => void | Promise<unknown>,
    onReopen?: (taskId: string) => void | Promise<unknown>,
): SwipeAction | null {
    if (isReadOnly) {
        return null
    }

    if (mode === 'review') {
        if (task.status === 'completed' && task.can_review_open && onReopen) {
            return {
                kind: 'danger',
                label: 'Переоткрыть',
                onClick: onReopen,
            }
        }

        return null
    }

    if (task.can_return) {
        return {
            kind: 'default',
            label: 'Вернуть',
            onClick: onReturn,
        }
    }

    return null
}

function getRightSwipeAction(
    mode: TaskRowActionMode,
    task: ResidentDutyTask,
    isReadOnly: boolean,
    onTake: (taskId: string) => void | Promise<unknown>,
    onVerify?: (taskId: string) => void | Promise<unknown>,
): SwipeAction | null {
    if (isReadOnly) {
        return null
    }

    if (mode === 'review') {
        if (task.status === 'completed' && task.can_verify && onVerify) {
            return {
                kind: 'default',
                label: 'Подтвердить',
                onClick: onVerify,
            }
        }

        return null
    }

    if (task.can_take) {
        return {
            kind: 'default',
            label: 'Взять',
            onClick: onTake,
        }
    }

    return null
}

export function TaskMobileCard({
    activeSwipeTaskId,
    isReadOnly = false,
    mode = 'default',
    pending,
    task,
    onTake,
    onReturn,
    onComplete,
    onOpen,
    onReopen,
    onSwipeActiveChange,
    onVerify,
}: Props) {
    const [swipeOffset, setSwipeOffset] = useState(0)
    const [isDragging, setIsDragging] = useState(false)
    const touchStartXRef = useRef<number | null>(null)
    const touchStartYRef = useRef<number | null>(null)
    const touchStartOffsetRef = useRef(0)
    const swipeAxisRef = useRef<'x' | 'y' | null>(null)
    const assigneeLabel = getAssigneeLabel(task)
    const statusConfig = getMobileStatusConfig(task)
    const showCheckbox = !isReadOnly && task.is_mine && (task.can_complete || task.can_open)
    const leftSwipeAction = getLeftSwipeAction(mode, task, isReadOnly, onReturn, onReopen)
    const rightSwipeAction = getRightSwipeAction(mode, task, isReadOnly, onTake, onVerify)
    const swipeEnabled = Boolean(leftSwipeAction || rightSwipeAction)
    const leftSwipeButtonWidth = leftSwipeAction
        ? (leftSwipeAction.kind === 'danger' ? DANGER_SWIPE_ACTION_WIDTH : DEFAULT_SWIPE_ACTION_WIDTH)
        : 0
    const rightSwipeButtonWidth = rightSwipeAction ? DEFAULT_SWIPE_ACTION_WIDTH : 0
    const leftSwipeWidth = leftSwipeAction ? leftSwipeButtonWidth + SWIPE_GAP : 0
    const rightSwipeWidth = rightSwipeAction ? rightSwipeButtonWidth + SWIPE_GAP : 0

    useEffect(() => {
        setSwipeOffset(0)
    }, [task.id, task.status, pending])

    useEffect(() => {
        if (activeSwipeTaskId !== task.id && swipeOffset !== 0) {
            setSwipeOffset(0)
            setIsDragging(false)
        }
    }, [activeSwipeTaskId, swipeOffset, task.id])

    function handleTouchStart(clientX: number, clientY: number) {
        if (!swipeEnabled) {
            return
        }

        touchStartXRef.current = clientX
        touchStartYRef.current = clientY
        touchStartOffsetRef.current = swipeOffset
        swipeAxisRef.current = null
        setIsDragging(false)
    }

    function handleTouchMove(clientX: number, clientY: number) {
        const touchStartX = touchStartXRef.current
        const touchStartY = touchStartYRef.current
        if (!swipeEnabled || touchStartX === null || touchStartY === null) {
            return
        }

        const deltaX = clientX - touchStartX
        const deltaY = clientY - touchStartY
        const absDeltaX = Math.abs(deltaX)
        const absDeltaY = Math.abs(deltaY)
        const activationThreshold = 10

        if (swipeAxisRef.current === null) {
            if (absDeltaX < activationThreshold && absDeltaY < activationThreshold) {
                return
            }

            swipeAxisRef.current = absDeltaX > absDeltaY ? 'x' : 'y'

            if (swipeAxisRef.current === 'x') {
                onSwipeActiveChange(task.id)
                touchStartXRef.current = clientX
                touchStartOffsetRef.current = swipeOffset
                setIsDragging(true)
                return
            }
        }

        if (swipeAxisRef.current !== 'x') {
            return
        }

        const minOffset = rightSwipeAction ? -rightSwipeWidth : 0
        const maxOffset = leftSwipeAction ? leftSwipeWidth : 0
        const startOffset = touchStartOffsetRef.current
        let nextOffset = startOffset + deltaX

        if (startOffset < 0 && nextOffset > 0) {
            nextOffset = 0
        }

        if (startOffset > 0 && nextOffset < 0) {
            nextOffset = 0
        }

        nextOffset = Math.max(minOffset, Math.min(maxOffset, nextOffset))
        setSwipeOffset(nextOffset)
    }

    function handleTouchEnd() {
        if (!swipeEnabled) {
            return
        }

        const openThreshold = 56
        if (swipeOffset <= -openThreshold && rightSwipeAction) {
            setSwipeOffset(-rightSwipeWidth)
            onSwipeActiveChange(task.id)
        } else if (swipeOffset >= openThreshold && leftSwipeAction) {
            setSwipeOffset(leftSwipeWidth)
            onSwipeActiveChange(task.id)
        } else {
            setSwipeOffset(0)
            if (activeSwipeTaskId === task.id) {
                onSwipeActiveChange(null)
            }
        }

        touchStartXRef.current = null
        touchStartYRef.current = null
        touchStartOffsetRef.current = 0
        swipeAxisRef.current = null
        setIsDragging(false)
    }

    async function handleSwipeAction(
        action?: SwipeAction,
    ) {
        if (!action) {
            return
        }

        setSwipeOffset(0)
        if (activeSwipeTaskId === task.id) {
            onSwipeActiveChange(null)
        }
        await action.onClick(task.id)
    }

    return (
        <div className={classes.swipeRoot}>
            {swipeEnabled && (
                <div className={classes.actionLayer}>
                    <div className={`${classes.actionSide} ${classes.actionSideLeft}`}>
                        {leftSwipeAction && (
                            <Button
                                className={classes.actionButton}
                                radius="xl"
                                variant="default"
                                style={{ width: `${leftSwipeButtonWidth}px` }}
                                loading={pending && swipeOffset > 0}
                                styles={{
                                    root: leftSwipeAction.kind === 'danger'
                                        ? {
                                            backgroundColor: '#FEE2E2',
                                            borderColor: '#991B1B',
                                            color: 'var(--app-color-text)',
                                            fontWeight: 500,
                                        }
                                        : {
                                            backgroundColor: '#FFFFFF',
                                            borderColor: '#DDE4E2',
                                            color: 'var(--app-color-text)',
                                            fontWeight: 500,
                                        },
                                }}
                                onClick={() => void handleSwipeAction(leftSwipeAction)}
                            >
                                {leftSwipeAction.label}
                            </Button>
                        )}
                    </div>

                    <div className={`${classes.actionSide} ${classes.actionSideRight}`}>
                        {rightSwipeAction && (
                            <Button
                                className={classes.actionButton}
                                radius="xl"
                                variant="default"
                                style={{ width: `${rightSwipeButtonWidth}px` }}
                                loading={pending && swipeOffset < 0}
                                styles={{
                                    root: {
                                        backgroundColor: '#FFFFFF',
                                        borderColor: '#DDE4E2',
                                        color: 'var(--app-color-text)',
                                        fontWeight: 500,
                                    },
                                }}
                                onClick={() => void handleSwipeAction(rightSwipeAction)}
                            >
                                {rightSwipeAction.label}
                            </Button>
                        )}
                    </div>
                </div>
            )}

            <Paper
                withBorder
                radius="lg"
                p="md"
                bg="white"
                className={classes.surface}
                data-dragging={isDragging ? 'true' : 'false'}
                style={{
                    borderColor: 'var(--app-color-border)',
                    transform: `translateX(${swipeOffset}px)`,
                }}
                onTouchStart={(event) => handleTouchStart(
                    event.changedTouches[0]?.clientX ?? 0,
                    event.changedTouches[0]?.clientY ?? 0,
                )}
                onTouchMove={(event) => handleTouchMove(
                    event.changedTouches[0]?.clientX ?? 0,
                    event.changedTouches[0]?.clientY ?? 0,
                )}
                onTouchEnd={handleTouchEnd}
                onTouchCancel={handleTouchEnd}
            >
                <Stack gap={8}>
                    <div className={classes.headerRow}>
                        <div className={classes.titleWrap}>
                            {showCheckbox && (
                                <div className={classes.checkboxWrap}>
                                    <Checkbox
                                        checked={task.status === 'completed'}
                                        disabled={pending}
                                        size="20px"
                                        radius="xl"
                                        iconColor="#FFFFFF"
                                        styles={{
                                            input: task.status === 'completed'
                                                ? {
                                                    backgroundColor: '#8C8C8C',
                                                    borderColor: '#8C8C8C',
                                                }
                                                : undefined,
                                        }}
                                        aria-label={`${task.status === 'completed' ? 'Отменить выполнение' : 'Выполнить'} задачу ${task.title}`}
                                        onChange={(event) => {
                                            if (event.currentTarget.checked) {
                                                void onComplete(task.id)
                                                return
                                            }

                                            void onOpen(task.id)
                                        }}
                                    />
                                </div>
                            )}

                            <Text fw={600} className={classes.title}>
                                {task.title}
                            </Text>
                        </div>
                    </div>

                    <div
                        className={classes.metaRow}
                        data-has-checkbox={showCheckbox ? 'true' : 'false'}
                    >
                        <Badge
                            radius="sm"
                            variant="filled"
                            styles={{
                                root: {
                                    backgroundColor: '#EEF2F1',
                                    color: 'var(--app-color-text)',
                                    fontWeight: 500,
                                },
                            }}
                        >
                            {task.cost} баллов
                        </Badge>

                        {statusConfig && <MobileTaskStatus task={task} />}
                        {assigneeLabel && (
                            <Text className={classes.assigneeText}>
                                {assigneeLabel}
                            </Text>
                        )}
                    </div>
                </Stack>
            </Paper>
        </div>
    )
}
