import { useState } from 'react'

import { CSS } from '@dnd-kit/utilities'
import { useSortable } from '@dnd-kit/sortable'
import {
    DotsSixVerticalIcon,
    DotsThreeVerticalIcon,
    PencilIcon,
    TrashIcon,
    UsersThreeIcon,
} from '@phosphor-icons/react'
import { ActionIcon, Menu, Text } from '@mantine/core'

import type { DutySettingsTeam } from '../../../features/duty-settings/model/types'
import { SettingsBadge } from '../../../shared/ui/SettingsBadge'
import { SettingsCardSurface } from '../../../shared/ui/SettingsCardSurface'
import classes from './DutySettingsTeamCard.module.css'

type Props = {
    team: DutySettingsTeam
    isDutyTeam: boolean
    disabled?: boolean
    onEdit: (teamId: string) => void
    onDelete: (team: DutySettingsTeam) => void
}

function formatMembersCount(count: number): string {
    const remainder10 = count % 10
    const remainder100 = count % 100

    if (remainder10 === 1 && remainder100 !== 11) {
        return `${count} участник`
    }
    if (remainder10 >= 2 && remainder10 <= 4 && (remainder100 < 12 || remainder100 > 14)) {
        return `${count} участника`
    }
    return `${count} участников`
}

export function DutySettingsTeamCard({ team, isDutyTeam, disabled = false, onEdit, onDelete }: Props) {
    const [menuOpened, setMenuOpened] = useState(false)
    const {
        attributes,
        listeners,
        setNodeRef,
        setActivatorNodeRef,
        transform,
        transition,
        isDragging,
    } = useSortable({ id: team.id, disabled })

    return (
        <div
            ref={setNodeRef}
            style={{
                transform: CSS.Transform.toString(transform),
                transition,
            }}
            data-menu-open={menuOpened ? 'true' : undefined}
            data-dragging={isDragging ? 'true' : undefined}
        >
            <SettingsCardSurface dragging={isDragging} menuOpen={menuOpened}>
                <div className={classes.content}>
                    <div className={classes.handleCell}>
                        <button
                            ref={setActivatorNodeRef}
                            type="button"
                            className={classes.handleButton}
                            aria-label="Изменить порядок команды"
                            data-dragging={isDragging ? 'true' : undefined}
                            disabled={disabled}
                            {...attributes}
                            {...listeners}
                        >
                            <DotsSixVerticalIcon size={25} />
                        </button>
                    </div>

                    <div className={classes.positionCell}>
                        <Text fw={500}>{team.rotation_position}</Text>
                    </div>

                    <div className={classes.leaderCell}>
                        <Text fw={500} className={classes.leaderText} c={team.leader ? undefined : 'dimmed'}>
                            {team.leader?.name ?? 'Не назначен'}
                        </Text>
                    </div>

                    <div className={classes.membersCell}>
                        <SettingsBadge>{formatMembersCount(team.members_count)}</SettingsBadge>
                    </div>

                    <div className={classes.statusCell}>
                        {isDutyTeam ? <SettingsBadge color="success">Дежурная</SettingsBadge> : null}
                    </div>

                    <div />

                    <div className={classes.actionsCell}>
                        <Menu
                            opened={menuOpened}
                            onChange={setMenuOpened}
                            withinPortal
                            position="bottom-end"
                        >
                            <Menu.Target>
                                <ActionIcon
                                    variant="subtle"
                                    color="gray"
                                    aria-label="Действия с командой"
                                    className={classes.actionButton}
                                    onPointerDown={(event) => event.stopPropagation()}
                                >
                                    <DotsThreeVerticalIcon size={25} />
                                </ActionIcon>
                            </Menu.Target>
                            <Menu.Dropdown>
                                <Menu.Item leftSection={<UsersThreeIcon size={25} />} onClick={() => setMenuOpened(false)}>
                                    Участники
                                </Menu.Item>
                                <Menu.Item leftSection={<PencilIcon size={25} />} onClick={() => onEdit(team.id)}>
                                    Редактировать
                                </Menu.Item>
                                <Menu.Item color="red" leftSection={<TrashIcon size={25} />} onClick={() => onDelete(team)}>
                                    Удалить
                                </Menu.Item>
                            </Menu.Dropdown>
                        </Menu>
                    </div>
                </div>
            </SettingsCardSurface>
        </div>
    )
}
