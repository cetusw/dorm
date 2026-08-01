import { Text, UnstyledButton } from '@mantine/core'

import type { TeamMemberSummary } from '../model/selectors'
import { MemberDutyProgressCard } from './MemberDutyProgressCard'
import classes from './TeamMemberTaskGroups.module.css'

type Props = {
    memberSummary: TeamMemberSummary
    tooltip: React.ReactNode
    onOpen: () => void
}

export function TeamMemberCard({
    memberSummary,
    tooltip,
    onOpen,
}: Props) {
    return (
        <UnstyledButton
            className={classes.memberCard}
            aria-label={`Открыть задачи участника ${memberSummary.member.name}`}
            onClick={onOpen}
        >
            <div className={classes.memberCardInner}>
                <Text className={classes.memberName}>{memberSummary.member.name}</Text>

                <div className={classes.memberProgress}>
                    <MemberDutyProgressCard
                        progress={memberSummary.progress}
                        showTitle={false}
                        withContainer={false}
                        tooltip={tooltip}
                        barWidth="100%"
                    />
                </div>
            </div>
        </UnstyledButton>
    )
}
