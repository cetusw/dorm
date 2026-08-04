import { useEffect, useRef, useState } from 'react'

import { Button, Checkbox, Paper, Stack, Text } from '@mantine/core'

import { DutyTaskStatusBadge } from '../../../entities/duty-task'
import {
    getLeftSwipeActionSpec,
    getRightSwipeActionSpec,
    type TaskActionHandler,
    type TaskActionSpec,
    type TaskRowActionMode,
} from '../model/taskActions'
import { getMobileTaskCardPresentation } from '../model/taskMobilePresentation'
import type { ResidentDutyTask } from '../model/types'
import { TaskCostBadge } from './TaskCostBadge'
import classes from './TaskMobileCard.module.css'

type Props = {
    activeSwipeTaskId: string | null
    isReadOnly?: boolean
    mode?: TaskRowActionMode
    pending: boolean
    task: ResidentDutyTask
    onTake: TaskActionHandler
    onReturn: TaskActionHandler
    onComplete: TaskActionHandler
    onOpen: TaskActionHandler
    onOpenDetails: (taskId: string) => void
    onReopen?: TaskActionHandler
    onSwipeActiveChange: (taskId: string | null) => void
    onVerify?: TaskActionHandler
}

type SwipeAction = {
    label: string
    onClick: TaskActionHandler
    tone: TaskActionSpec['tone']
}

const SWIPE_GAP = 6
const DEFAULT_SWIPE_ACTION_WIDTH = 128
const DANGER_SWIPE_ACTION_WIDTH = 136

function getLeftSwipeAction(
    mode: TaskRowActionMode,
    task: ResidentDutyTask,
    isReadOnly: boolean,
    onReturn: TaskActionHandler,
    onReopen?: TaskActionHandler,
): SwipeAction | null {
    const action = getLeftSwipeActionSpec({
        isReadOnly,
        mode,
        task,
    })

    if (!action) {
        return null
    }

    if (action.kind === 'return') {
        return {
            label: action.label,
            onClick: onReturn,
            tone: action.tone,
        }
    }

    if (action.kind === 'reopen' && onReopen) {
        return {
            label: action.label,
            onClick: onReopen,
            tone: action.tone,
        }
    }

    return null
}

function getRightSwipeAction(
    mode: TaskRowActionMode,
    task: ResidentDutyTask,
    isReadOnly: boolean,
    onTake: TaskActionHandler,
    onVerify?: TaskActionHandler,
): SwipeAction | null {
    const action = getRightSwipeActionSpec({
        isReadOnly,
        mode,
        task,
    })

    if (!action) {
        return null
    }

    if (action.kind === 'take') {
        return {
            label: action.label,
            onClick: onTake,
            tone: action.tone,
        }
    }

    if (action.kind === 'verify' && onVerify) {
        return {
            label: action.label,
            onClick: onVerify,
            tone: action.tone,
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
    onOpenDetails,
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
    const suppressClickTimeoutRef = useRef<number | null>(null)
    const suppressNextClickRef = useRef(false)
    const presentation = getMobileTaskCardPresentation({
        isReadOnly,
        mode,
        task,
    })
    const leftSwipeAction = getLeftSwipeAction(mode, task, isReadOnly, onReturn, onReopen)
    const rightSwipeAction = getRightSwipeAction(mode, task, isReadOnly, onTake, onVerify)
    const swipeEnabled = Boolean(leftSwipeAction || rightSwipeAction)
    const leftSwipeButtonWidth = leftSwipeAction
        ? (leftSwipeAction.tone === 'danger' ? DANGER_SWIPE_ACTION_WIDTH : DEFAULT_SWIPE_ACTION_WIDTH)
        : 0
    const rightSwipeButtonWidth = rightSwipeAction ? DEFAULT_SWIPE_ACTION_WIDTH : 0
    const leftSwipeWidth = leftSwipeAction ? leftSwipeButtonWidth + SWIPE_GAP : 0
    const rightSwipeWidth = rightSwipeAction ? rightSwipeButtonWidth + SWIPE_GAP : 0
    const shouldRenderActionLayer = swipeEnabled && (isDragging || swipeOffset !== 0)

    useEffect(() => {
        setSwipeOffset(0)
    }, [task.id, task.status, pending])

    useEffect(() => {
        if (activeSwipeTaskId !== task.id && swipeOffset !== 0) {
            setSwipeOffset(0)
            setIsDragging(false)
        }
    }, [activeSwipeTaskId, swipeOffset, task.id])

    useEffect(() => () => {
        if (suppressClickTimeoutRef.current !== null) {
            window.clearTimeout(suppressClickTimeoutRef.current)
        }
    }, [])

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

        if (swipeAxisRef.current === 'x') {
            suppressNextClickRef.current = true
            if (suppressClickTimeoutRef.current !== null) {
                window.clearTimeout(suppressClickTimeoutRef.current)
            }
            suppressClickTimeoutRef.current = window.setTimeout(() => {
                suppressNextClickRef.current = false
            }, 250)
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

    function handleSurfaceClick() {
        if (suppressNextClickRef.current || isDragging) {
            return
        }

        if (swipeOffset !== 0) {
            setSwipeOffset(0)
            if (activeSwipeTaskId === task.id) {
                onSwipeActiveChange(null)
            }
            return
        }

        onSwipeActiveChange(null)
        onOpenDetails(task.id)
    }

    return (
        <div className={classes.swipeRoot}>
            {shouldRenderActionLayer && (
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
                                    root: leftSwipeAction.tone === 'danger'
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
                onClick={handleSurfaceClick}
            >
                <Stack gap={8}>
                    <div className={classes.headerRow}>
                        <div className={classes.titleWrap}>
                            {presentation.showCheckbox && (
                                <div className={classes.checkboxWrap}>
                                    <Checkbox
                                        checked={task.status === 'completed'}
                                        disabled={pending}
                                        size="25px"
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
                                        onClick={(event) => {
                                            event.stopPropagation()
                                        }}
                                        onChange={async (event) => {
                                            event.stopPropagation()
                                            if (event.currentTarget.checked) {
                                                await onComplete(task.id)
                                                return
                                            }

                                            await onOpen(task.id)
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
                        data-has-checkbox={presentation.showCheckbox ? 'true' : 'false'}
                    >
                        <TaskCostBadge cost={task.cost} />

                        {presentation.showStatus && <DutyTaskStatusBadge status={task.status} justify="flex-start" />}
                        {presentation.showAssignee && presentation.assigneeLabel && (
                            <Text className={classes.assigneeText}>
                                {presentation.assigneeLabel}
                            </Text>
                        )}
                    </div>
                </Stack>
            </Paper>
        </div>
    )
}
