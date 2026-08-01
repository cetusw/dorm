import { XIcon } from '@phosphor-icons/react'
import { ActionIcon, Button, Drawer, FocusTrap, Group, Text } from '@mantine/core'

import { DutyTaskStatusBadge } from '../../../entities/duty-task'
import { formatEstimatedDuration } from '../model/formatEstimatedDuration'
import { getDrawerActionSpecs, type TaskActionHandler, type TaskRowActionMode } from '../model/taskActions'
import type { ResidentDutyTask } from '../model/types'
import { TaskCostBadge } from './TaskCostBadge'
import classes from './TaskMobileDetailsDrawer.module.css'

type Props = {
    isReadOnly?: boolean
    mode?: TaskRowActionMode
    opened: boolean
    pendingTaskId: string | null
    task: ResidentDutyTask | null
    onClose: () => void
    onTake: TaskActionHandler
    onReturn: TaskActionHandler
    onReopen: TaskActionHandler
    onVerify: TaskActionHandler
}

function getActionHandler(
    task: ResidentDutyTask,
    actionKind: 'take' | 'return' | 'reopen' | 'verify',
    handlers: Pick<Props, 'onTake' | 'onReturn' | 'onReopen' | 'onVerify'>,
): (() => Promise<boolean>) | null {
    switch (actionKind) {
        case 'take':
            return () => handlers.onTake(task.id)
        case 'return':
            return () => handlers.onReturn(task.id)
        case 'reopen':
            return () => handlers.onReopen(task.id)
        case 'verify':
            return () => handlers.onVerify(task.id)
        default:
            return null
    }
}

function getAssigneeLabel(task: ResidentDutyTask): string | null {
    if (!task.assignee_name) {
        return null
    }

    return task.is_mine ? 'Вы' : task.assignee_name
}

export function TaskMobileDetailsDrawer({
    isReadOnly = false,
    mode = 'default',
    opened,
    pendingTaskId,
    task,
    onClose,
    onTake,
    onReturn,
    onReopen,
    onVerify,
}: Props) {
    const actionSpecs = task
        ? getDrawerActionSpecs({
            isReadOnly,
            mode,
            task,
        })
        : []
    const assigneeLabel = task ? getAssigneeLabel(task) : null

    return (
        <Drawer
            opened={opened}
            onClose={onClose}
            position="bottom"
            hiddenFrom="md"
            withCloseButton={false}
            classNames={{
                body: classes.content,
                content: classes.drawerContent,
            }}
        >
            <FocusTrap.InitialFocus />

            {task ? (
                <>
                    <div className={classes.header}>
                        <Text className={classes.title}>{task.title}</Text>
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

                    <Text className={classes.infoRow}>
                        {task.area_floor} этаж · {task.area_name}
                    </Text>

                    <Group gap="sm" wrap="wrap" className={classes.infoRow}>
                        <TaskCostBadge cost={task.cost} />
                        <Text size="sm">{formatEstimatedDuration(task.cost)}</Text>
                    </Group>

                    <div className={classes.statusRow}>
                        <DutyTaskStatusBadge status={task.status} justify="flex-start" />
                        {assigneeLabel ? <Text size="sm">{assigneeLabel}</Text> : null}
                    </div>

                    {actionSpecs.length > 0 ? (
                        <div className={classes.actions}>
                            {actionSpecs.length === 1 ? (
                                <Button
                                    className={classes.singleActionButton}
                                    variant="default"
                                    loading={pendingTaskId === task.id}
                                    styles={{
                                        root: actionSpecs[0].tone === 'danger'
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
                                    onClick={async () => {
                                        const handler = getActionHandler(task, actionSpecs[0].kind, {
                                            onTake,
                                            onReturn,
                                            onReopen,
                                            onVerify,
                                        })
                                        const success = handler ? await handler() : false

                                        if (success) {
                                            onClose()
                                        }
                                    }}
                                >
                                    {actionSpecs[0].label}
                                </Button>
                            ) : (
                                <div className={classes.splitActions}>
                                    {actionSpecs.map((action) => (
                                        <div key={action.kind} className={classes.splitActionWrap}>
                                            <Button
                                                className={classes.splitActionButton}
                                                variant="default"
                                                loading={pendingTaskId === task.id}
                                                styles={{
                                                    root: action.tone === 'danger'
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
                                                onClick={async () => {
                                                    const handler = getActionHandler(task, action.kind, {
                                                        onTake,
                                                        onReturn,
                                                        onReopen,
                                                        onVerify,
                                                    })
                                                    const success = handler ? await handler() : false

                                                    if (success) {
                                                        onClose()
                                                    }
                                                }}
                                            >
                                                {action.label}
                                            </Button>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>
                    ) : null}
                </>
            ) : null}
        </Drawer>
    )
}
