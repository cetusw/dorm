import { Button, Checkbox, Table, Text } from '@mantine/core'

import { DutyTaskStatusBadge } from '../../../entities/duty-task'
import {
    getDrawerActionSpecs,
    type TaskActionHandler,
    type TaskRowActionMode,
} from '../model/taskActions'
import type { ResidentDutyTask } from '../model/types'
import { TaskCostBadge } from './TaskCostBadge'
import classes from './TaskTableRow.module.css'

type Props = {
    isReadOnly?: boolean
    layout?: 'default' | 'drawer'
    mode?: TaskRowActionMode
    pending: boolean
    task: ResidentDutyTask
    onTake: TaskActionHandler
    onReturn: TaskActionHandler
    onComplete: TaskActionHandler
    onOpen: TaskActionHandler
    onReopen?: TaskActionHandler
    onVerify?: TaskActionHandler
}

type HoverAction = {
    label: string
    kind?: 'danger' | 'default'
    loading: boolean
    onClick: () => Promise<boolean>
}

function getHoverActions(
    mode: TaskRowActionMode,
    task: ResidentDutyTask,
    pending: boolean,
    onTake: Props['onTake'],
    onReturn: Props['onReturn'],
    onReopen: Props['onReopen'],
    onVerify: Props['onVerify'],
) : HoverAction[] {
    return getDrawerActionSpecs({
        isReadOnly: false,
        mode,
        task,
    }).flatMap((action) => {
        switch (action.kind) {
            case 'take':
                return [{
                    label: action.label,
                    kind: action.tone,
                    loading: pending,
                    onClick: () => onTake(task.id),
                }]
            case 'return':
                return [{
                    label: action.label,
                    kind: action.tone,
                    loading: pending,
                    onClick: () => onReturn(task.id),
                }]
            case 'reopen':
                return onReopen ? [{
                    label: action.label,
                    kind: action.tone,
                    loading: pending,
                    onClick: () => onReopen(task.id),
                }] : []
            case 'verify':
                return onVerify ? [{
                    label: action.label,
                    kind: action.tone,
                    loading: pending,
                    onClick: () => onVerify(task.id),
                }] : []
            default:
                return []
        }
    })
}

export function TaskTableRow({
    isReadOnly = false,
    layout = 'default',
    mode = 'default',
    pending,
    task,
    onTake,
    onReturn,
    onComplete,
    onOpen,
    onReopen,
    onVerify,
}: Props) {
    const hoverActions = isReadOnly ? [] : getHoverActions(mode, task, pending, onTake, onReturn, onReopen, onVerify)
    const showCheckbox = task.is_mine && (task.can_complete || task.can_open)
    const checkboxChecked = task.status === 'completed'

    return (
        <Table.Tr className={classes.row}>
            <Table.Td className={classes.cell} colSpan={1}>
                <div
                    className={classes.surface}
                    data-has-checkbox={!isReadOnly && showCheckbox ? 'true' : 'false'}
                    data-layout={layout}
                >
                    <div className={classes.checkboxSlot}>
                        {!isReadOnly && showCheckbox && (
                            <Checkbox
                                className={classes.checkbox}
                                checked={checkboxChecked}
                                disabled={pending}
                                size="20px"
                                radius="xl"
                                iconColor="#FFFFFF"
                                styles={{
                                    input: checkboxChecked
                                        ? {
                                            backgroundColor: '#8C8C8C',
                                            borderColor: '#8C8C8C',
                                        }
                                        : undefined,
                                }}
                                aria-label={`${checkboxChecked ? 'Отменить выполнение' : 'Выполнить'} задачу ${task.title}`}
                                onChange={(event) => {
                                    if (event.currentTarget.checked) {
                                        void onComplete(task.id)
                                        return
                                    }

                                    void onOpen(task.id)
                                }}
                            />
                        )}
                    </div>

                    <div className={classes.titleCell}>
                        <Text fw={600} className={classes.title}>
                            {task.title}
                        </Text>
                    </div>

                    <div className={classes.scoreCell}>
                        <TaskCostBadge cost={task.cost} />
                    </div>

                    <div className={classes.assigneeCell}>
                        {task.assignee_name ? (
                            <Text className={classes.assigneeText}>
                                {task.is_mine ? 'Вы' : task.assignee_name}
                            </Text>
                        ) : null}
                    </div>

                    <div className={classes.right}>
                        <div className={[classes.persistent, hoverActions.length > 0 ? classes.persistentHiddenOnHover : ''].join(' ').trim()}>
                            <DutyTaskStatusBadge status={task.status} />
                        </div>

                        {hoverActions.length > 0 && (
                            <div className={[classes.hoverAction, classes.hoverActionVisible].join(' ')}>
                                <div className={classes.hoverActionsGroup}>
                                    {hoverActions.map((action) => (
                                        <Button
                                            key={action.label}
                                            size="sm"
                                            radius="xl"
                                            variant="default"
                                            loading={action.loading}
                                            styles={{
                                                root: action.kind === 'danger'
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
                                            onClick={action.onClick}
                                        >
                                            {action.label}
                                        </Button>
                                    ))}
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </Table.Td>
        </Table.Tr>
    )
}
