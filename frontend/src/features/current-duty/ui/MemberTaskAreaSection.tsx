import { Stack, Text } from '@mantine/core'

import type { TeamMemberTaskAreaGroup } from '../model/selectors'
import { TeamMemberTaskCard } from './TeamMemberTaskCard'
import classes from './TeamMemberTaskGroups.module.css'

type Props = {
    group: TeamMemberTaskAreaGroup
}

export function MemberTaskAreaSection({ group }: Props) {
    return (
        <div className={classes.areaSection}>
            <Text className={classes.areaTitle}>{group.label}</Text>

            <Stack gap="sm">
                {group.tasks.map((task) => (
                    <TeamMemberTaskCard key={task.id} task={task} />
                ))}
            </Stack>
        </div>
    )
}
