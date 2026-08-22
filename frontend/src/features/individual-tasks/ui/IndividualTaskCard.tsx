import { useState } from 'react'

import { CheckIcon, DotsThreeVerticalIcon, PencilIcon, TrashIcon, XIcon } from '@phosphor-icons/react'
import { ActionIcon, Menu, Text, Tooltip } from '@mantine/core'

import { SettingsBadge } from '../../../shared/ui/SettingsBadge'
import type { IndividualTask } from '../model/types'
import { formatArea, formatDeadline, formatRedemptionWeight } from '../model/utils'
import classes from './IndividualTaskCard.module.css'

type Props = { task: IndividualTask; pending: boolean; onEdit: () => void; onDelete: () => void; onVerify: () => void; onReject: () => void }
type ReviewAction = { label: string; ariaLabel: string; icon: typeof CheckIcon; visible: boolean; onClick: () => void }

export function IndividualTaskCard({ task, pending, onEdit, onDelete, onVerify, onReject }: Props) {
    const [menuOpened, setMenuOpened] = useState(false)
    const reviewActions: ReviewAction[] = [
        { label: 'Подтвердить', ariaLabel: 'Подтвердить индивидуальную задачу', icon: CheckIcon, visible: task.can_verify, onClick: onVerify },
        { label: 'Переоткрыть', ariaLabel: 'Переоткрыть индивидуальную задачу', icon: XIcon, visible: task.can_reject, onClick: onReject },
    ]
    const showMenu = task.status === 'issued' && (task.can_edit || task.can_delete)

    return <div className={classes.card} data-menu-open={menuOpened || undefined}>
        <div className={classes.titleRow}>
            <Text className={classes.title}>{task.title}</Text>
            {task.status === 'completed' ? <div className={classes.actions}>{reviewActions.filter((action) => action.visible).map((action) => { const Icon = action.icon; return <Tooltip key={action.label} label={action.label}><ActionIcon aria-label={action.ariaLabel} variant="subtle" color="gray" loading={pending} onClick={action.onClick}><Icon size={20} /></ActionIcon></Tooltip> })}</div> : null}
            {showMenu ? <Menu opened={menuOpened} onChange={setMenuOpened} withinPortal position="bottom-end"><Menu.Target><ActionIcon className={classes.menuButton} aria-label="Действия с индивидуальной задачей" variant="subtle" color="gray"><DotsThreeVerticalIcon size={24} /></ActionIcon></Menu.Target><Menu.Dropdown>{task.can_edit ? <Menu.Item leftSection={<PencilIcon size={24} />} onClick={onEdit}>Редактировать</Menu.Item> : null}{task.can_delete ? <Menu.Item color="red" leftSection={<TrashIcon size={24} />} onClick={onDelete}>Удалить</Menu.Item> : null}</Menu.Dropdown></Menu> : null}
        </div>
        <div className={classes.meta}>
            {task.status === 'completed' ? <SettingsBadge color="success">Выполнено</SettingsBadge> : null}
            {task.redemption_weight > 0 ? <SettingsBadge>- {formatRedemptionWeight(task.redemption_weight)}</SettingsBadge> : null}
            {task.area ? <SettingsBadge color="area">{formatArea(task.area)}</SettingsBadge> : null}
            {task.deadline ? <Text className={`${classes.deadline} ${task.is_overdue ? classes.overdue : ''}`}>{formatDeadline(task.deadline)}</Text> : null}
        </div>
    </div>
}
