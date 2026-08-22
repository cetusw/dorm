import { Checkbox, Paper, Text } from '@mantine/core'

import { SettingsBadge } from '../../../shared/ui/SettingsBadge'
import dutyCardClasses from '../../current-duty/ui/TaskMobileCard.module.css'
import type { IndividualTask } from '../model/types'
import { formatArea, formatDeadline, formatRedemptionWeight } from '../model/utils'

type Props = {
    task: IndividualTask
    pending: boolean
    onToggle: (task: IndividualTask) => void
}

export function IndividualTaskResidentCard({ task, pending, onToggle }: Props) {
    const completed = task.status === 'completed'
    const interactive = !pending && (completed ? task.can_open : task.can_complete)

    return (
        <div className={dutyCardClasses.cardRoot}>
            <Paper withBorder={false} radius={0} p={0} bg="transparent" className={dutyCardClasses.surface}>
                <div className={dutyCardClasses.headerRow}>
                    <div className={dutyCardClasses.checkboxWrap}>
                        <Checkbox
                            checked={completed}
                            disabled={pending}
                            readOnly={!interactive}
                            size="25px"
                            radius="xl"
                            iconColor="#FFFFFF"
                            styles={{
                                input: completed
                                    ? { backgroundColor: '#8C8C8C', borderColor: '#8C8C8C' }
                                    : undefined,
                            }}
                            aria-label={`${completed ? 'Отменить выполнение' : 'Выполнить'} задачу ${task.title}`}
                            onChange={() => {
                                if (interactive) {
                                    onToggle(task)
                                }
                            }}
                        />
                    </div>

                    <div className={dutyCardClasses.content}>
                        <div className={dutyCardClasses.titleWrap}>
                            <Text fw={400} className={dutyCardClasses.title}>{task.title}</Text>
                        </div>
                        <div className={dutyCardClasses.metaRow}>
                            {task.redemption_weight > 0 ? (
                                <SettingsBadge>
                                    <span className="individual-task-desktop-weight">- {formatRedemptionWeight(task.redemption_weight)}</span>
                                    <span className="individual-task-mobile-weight">- {String(task.redemption_weight).replace('.', ',')}</span>
                                </SettingsBadge>
                            ) : null}
                            {task.area ? <SettingsBadge color="area">{formatArea(task.area)}</SettingsBadge> : null}
                            {task.deadline ? <Text size="sm" c={task.is_overdue ? '#991B1B' : 'dimmed'}>{formatDeadline(task.deadline)}</Text> : null}
                        </div>
                    </div>
                </div>
            </Paper>
        </div>
    )
}
