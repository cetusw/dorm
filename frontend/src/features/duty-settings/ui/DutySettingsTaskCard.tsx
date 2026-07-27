import { useState } from 'react'

import { DotsThreeVerticalIcon, PencilIcon, TrashIcon } from '@phosphor-icons/react'
import { ActionIcon, Menu, Text, Tooltip } from '@mantine/core'

import type { DutySettingsTask } from '../model/types'
import { formatLastCompletedAt } from '../model/utils'
import { SettingsBadge } from '../../../shared/ui/SettingsBadge'
import { SettingsCardSurface } from '../../../shared/ui/SettingsCardSurface'
import classes from './DutySettingsTaskCard.module.css'

type Props = {
    task: DutySettingsTask
    onEdit: (taskId: string) => void
    onDelete: (taskId: string) => void
}

export function DutySettingsTaskCard({ task, onEdit, onDelete }: Props) {
    const [menuOpened, setMenuOpened] = useState(false)

    return (
        <SettingsCardSurface className={classes.cardRoot} menuOpen={menuOpened}>
            <div className={classes.content}>
                <div className={classes.titleCell}>
                    <Text fw={500} className={classes.title}>{task.title}</Text>
                </div>

                <div className={classes.scoreCell}>
                    <SettingsBadge>{task.cost} баллов</SettingsBadge>
                </div>

                <div className={classes.dateCell}>
                    <Tooltip label="Дата последнего выполнения">
                        <Text fw={500}>{formatLastCompletedAt(task.last_completed_at)}</Text>
                    </Tooltip>
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
                            <Menu.Item color="red" leftSection={<TrashIcon size={25} />} onClick={() => onDelete(task.id)}>
                                Удалить
                            </Menu.Item>
                        </Menu.Dropdown>
                    </Menu>
                </div>
            </div>
        </SettingsCardSurface>
    )
}
