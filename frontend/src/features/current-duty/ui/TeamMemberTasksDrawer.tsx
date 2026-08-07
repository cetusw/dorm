import { Drawer, FocusTrap, Stack, Text } from '@mantine/core'
import { useMediaQuery } from '@mantine/hooks'

import type {
    TeamMemberSummary,
    TeamMemberTaskAreaGroup,
} from '../model/selectors'
import { EmptyState } from '../../../shared/ui/EmptyState'
import { MemberDutyProgressCard } from './MemberDutyProgressCard'
import { MemberTaskAreaSection } from './MemberTaskAreaSection'
import classes from './TeamMemberTaskGroups.module.css'

type Props = {
    memberSummary: TeamMemberSummary | null
    opened: boolean
    taskGroups: TeamMemberTaskAreaGroup[]
    tooltip: React.ReactNode
    onClose: () => void
}

export function TeamMemberTasksDrawer({
    memberSummary,
    opened,
    taskGroups,
    tooltip,
    onClose,
}: Props) {
    const isMobile = useMediaQuery('(max-width: 48em)')

    return (
        <Drawer
            opened={opened}
            onClose={onClose}
            position={isMobile ? 'bottom' : 'right'}
            size={isMobile ? 'auto' : 860}
            title={memberSummary?.member.name ?? ''}
            classNames={{
                body: classes.drawerBody,
                content: classes.drawerContent,
            }}
        >
            <FocusTrap.InitialFocus />

            {memberSummary ? (
                <Stack gap="lg">
                    <div className={classes.drawerProgress}>
                        <MemberDutyProgressCard
                            progress={memberSummary.progress}
                            withContainer={false}
                            tooltip={tooltip}
                            barWidth="100%"
                        />
                    </div>

                    {taskGroups.length === 0 ? (
                        <EmptyState
                            title="Задачи не найдены"
                            description="Этот участник ещё не взял ни одной задачи."
                        />
                    ) : (
                        <Stack gap="lg">
                            {taskGroups.map((group) => (
                                <MemberTaskAreaSection key={group.areaId} group={group} />
                            ))}
                        </Stack>
                    )}
                </Stack>
            ) : (
                <Text c="dimmed">Участник не выбран</Text>
            )}
        </Drawer>
    )
}
