import { useMemo, useState } from 'react'

import {
    CheckIcon,
    DotsThreeVerticalIcon,
    PencilIcon,
    TrashIcon,
} from '@phosphor-icons/react'
import { ActionIcon, Menu, Text, Tooltip } from '@mantine/core'

import { TaskStatusBadge } from '../../current-duty/ui/TaskStatusBadge'
import type { ResidentDutyTask } from '../../current-duty/model/types'
import { formatLastCompletedAt } from '../model/utils'
import type { DutySettingsTask } from '../model/types'
import { SettingsBadge } from '../../../shared/ui/SettingsBadge'
import { SettingsCardSurface } from '../../../shared/ui/SettingsCardSurface'
import classes from './DutySettingsTaskCard.module.css'

type Props = {
    task: DutySettingsTask
    onEdit: (taskId: string) => void
    onInclude: (taskId: string) => void
    onExclude: (taskId: string) => void
}

function toResidentDutyTask(task: DutySettingsTask): ResidentDutyTask {
    return {
        id: task.id,
        area_id: 0,
        area_name: '',
        area_floor: 0,
        title: task.title,
        cost: task.cost,
        status: task.status === '' ? 'free' : task.status,
        is_mine: false,
        can_take: false,
        can_return: false,
        can_complete: false,
        can_open: false,
        can_verify: false,
        can_review_open: false,
    }
}

export function DutySettingsTaskCard({ task, onEdit, onInclude, onExclude }: Props) {
    const [menuOpened, setMenuOpened] = useState(false)
    const badgeTask = useMemo(() => toResidentDutyTask(task), [task])
    const canExclude = task.is_included && (task.status === 'free' || task.status === 'assigned' || task.status === '')
    const shouldShowStatus = task.is_included && (task.status === 'completed' || task.status === 'verified')

    return (
        <SettingsCardSurface
            className={classes.cardRoot}
            menuOpen={menuOpened}
            muted={!task.is_included}
        >
            <div className={classes.content}>
                <div className={classes.titleCell}>
                    <Text fw={500} className={classes.title}>{task.title}</Text>
                </div>

                <div className={classes.scoreCell}>
                    <SettingsBadge>{task.cost} баллов</SettingsBadge>
                </div>

                <div className={classes.dateCell}>
                    {task.frequency === 0 ? (
                        <SettingsBadge>Одноразовая задача</SettingsBadge>
                    ) : (
                        <Tooltip label="Дата последнего выполнения">
                            <Text fw={500}>{formatLastCompletedAt(task.last_completed_at)}</Text>
                        </Tooltip>
                    )}
                </div>

                <div className={classes.assigneeCell}>
                    {task.is_included && task.assignee_name ? (
                        <Text fw={500} className={classes.assigneeText}>{task.assignee_name}</Text>
                    ) : null}
                </div>

                <div className={classes.statusCell}>
                    {shouldShowStatus ? <TaskStatusBadge task={badgeTask} justify="flex-start" /> : null}
                </div>

                <div className={classes.actions}>
                    <Menu opened={menuOpened} onChange={setMenuOpened} withinPortal position="bottom-end">
                        <Menu.Target>
                            <ActionIcon
                                variant="subtle"
                                color="gray"
                                aria-label="Действия с задачей"
                                className={classes.iconButton}
                            >
                                <DotsThreeVerticalIcon size={25} />
                            </ActionIcon>
                        </Menu.Target>
                        <Menu.Dropdown>
                            <Menu.Item leftSection={<PencilIcon size={25} />} onClick={() => onEdit(task.id)}>
                                Редактировать
                            </Menu.Item>
                            {task.is_included ? (
                                canExclude ? (
                                    <Menu.Item color="red" leftSection={<TrashIcon size={25} />} onClick={() => onExclude(task.id)}>
                                        Исключить из дежурства
                                    </Menu.Item>
                                ) : null
                            ) : (
                                <Menu.Item leftSection={<CheckIcon size={25} />} onClick={() => onInclude(task.id)}>
                                    Включить в дежурство
                                </Menu.Item>
                            )}
                        </Menu.Dropdown>
                    </Menu>
                </div>
            </div>
        </SettingsCardSurface>
    )
}
