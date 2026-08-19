import { ArrowBendRightDownIcon, ArrowBendRightUpIcon, CheckIcon, XIcon } from '@phosphor-icons/react'
import { ActionIcon, Checkbox, Table, Text, Tooltip } from '@mantine/core'

import { DutyTaskStatusBadge } from '../../../entities/duty-task'
import { getTaskCardPresentation } from '../model/taskCardPresentation'
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
    actionKind: 'take' | 'return' | 'reopen' | 'verify'
    label: string
    tone?: 'danger' | 'default'
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
    }).flatMap<HoverAction>((action) => {
        switch (action.kind) {
            case 'take':
                return [{
                    label: action.label,
                    actionKind: action.kind,
                    tone: action.tone,
                    loading: pending,
                    onClick: () => onTake(task.id),
                }]
            case 'return':
                return [{
                    label: action.label,
                    actionKind: action.kind,
                    tone: action.tone,
                    loading: pending,
                    onClick: () => onReturn(task.id),
                }]
            case 'reopen':
                return onReopen ? [{
                    label: action.label,
                    actionKind: action.kind,
                    tone: action.tone,
                    loading: pending,
                    onClick: () => onReopen(task.id),
                }] : []
            case 'verify':
                return onVerify ? [{
                    label: action.label,
                    actionKind: action.kind,
                    tone: action.tone,
                    loading: pending,
                    onClick: () => onVerify(task.id),
                }] : []
            default:
                return []
        }
    })
}

function getActionIcon(actionKind: HoverAction['actionKind']) {
    switch (actionKind) {
        case 'take':
            return <ArrowBendRightDownIcon size={20} />
        case 'return':
            return <ArrowBendRightUpIcon size={20} />
        case 'verify':
            return <CheckIcon size={20} />
        case 'reopen':
            return <XIcon size={20} />
        default:
            return null
    }
}

function getActionTooltip(actionKind: HoverAction['actionKind']) {
    switch (actionKind) {
        case 'take':
            return 'Взять задачу'
        case 'return':
            return 'Вернуть задачу'
        case 'verify':
            return 'Подтвердить задачу'
        case 'reopen':
            return 'Переоткрыть задачу'
        default:
            return ''
    }
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
    const hoverActions = isReadOnly
        ? []
        : getHoverActions(mode, task, pending, onTake, onReturn, onReopen, onVerify)
            .sort((left, right) => {
                const order: Record<HoverAction['actionKind'], number> = {
                    verify: 0,
                    reopen: 1,
                    take: 2,
                    return: 3,
                }

                return order[left.actionKind] - order[right.actionKind]
            })
    const presentation = getTaskCardPresentation({
        isReadOnly,
        task,
    })
    const showCheckbox = presentation.showCheckbox
    const checkboxChecked = task.status === 'completed' || task.status === 'verified'
    const isCheckboxInteractive = !pending && task.is_mine && (task.can_complete || task.can_open)

    function handleCheckboxToggle() {
        if (!isCheckboxInteractive) {
            return
        }

        if (checkboxChecked) {
            void onOpen(task.id)
            return
        }

        void onComplete(task.id)
    }

    return (
        <Table.Tr className={classes.row}>
            <Table.Td className={classes.cell} colSpan={1}>
                <div
                    className={classes.surface}
                    data-has-checkbox={!isReadOnly && showCheckbox ? 'true' : 'false'}
                    data-layout={layout}
                >
                    <div className={classes.main}>
                        {showCheckbox && (
                            <div
                                className={classes.checkboxSlot}
                                onClick={(event) => {
                                    event.stopPropagation()
                                    handleCheckboxToggle()
                                }}
                            >
                                <Checkbox
                                    className={classes.checkbox}
                                    checked={checkboxChecked}
                                    disabled={pending}
                                    readOnly={!isCheckboxInteractive}
                                    size="25px"
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
                                    onClick={(event) => {
                                        event.stopPropagation()
                                    }}
                                    onChange={(event) => {
                                        event.stopPropagation()
                                        handleCheckboxToggle()
                                    }}
                                />
                            </div>
                        )}

                        <div className={classes.content}>
                            <div className={classes.titleCell}>
                                <Text fw={400} className={classes.title}>
                                    {task.title}
                                </Text>

                                {hoverActions.length > 0 && (
                                    <div className={[classes.titleActions, classes.titleActionsVisible].join(' ')}>
                                        {hoverActions.map((action) => (
                                            <Tooltip key={action.label} label={getActionTooltip(action.actionKind)} position="top">
                                                <ActionIcon
                                                    size={30}
                                                    radius="md"
                                                    variant="subtle"
                                                    color="gray"
                                                    className={classes.titleActionButton}
                                                    aria-label={getActionTooltip(action.actionKind)}
                                                    disabled={action.loading}
                                                    onClick={() => {
                                                        void action.onClick()
                                                    }}
                                                >
                                                    {getActionIcon(action.actionKind)}
                                                </ActionIcon>
                                            </Tooltip>
                                        ))}
                                    </div>
                                )}
                            </div>

                            <div className={classes.metaRow}>
                                <TaskCostBadge cost={task.cost} />
                                <div className={classes.metaRight}>
                                    {presentation.showStatus && <DutyTaskStatusBadge status={task.status} justify="flex-start" />}
                                    {presentation.showAssignee && presentation.assigneeLabel ? (
                                        <Text className={classes.assigneeText}>
                                            {presentation.assigneeLabel}
                                        </Text>
                                    ) : null}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </Table.Td>
        </Table.Tr>
    )
}
