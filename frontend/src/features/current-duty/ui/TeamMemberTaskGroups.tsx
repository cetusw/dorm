import { useMemo, useState } from 'react'

import { Stack, Text } from '@mantine/core'

import {
    groupAndSortMemberTasks,
    selectSortedTeamMembers,
    selectTasksForTeamMember,
} from '../model/selectors'
import type { ResidentCurrentDuty } from '../model/types'
import { EmptyState } from '../../../shared/ui/EmptyState'
import { TeamMemberCard } from './TeamMemberCard'
import { TeamMemberTasksDrawer } from './TeamMemberTasksDrawer'

type Props = {
    duty: ResidentCurrentDuty
}

function TooltipContent({ lines }: { lines: string[] }) {
    return (
        <Stack gap={2}>
            {lines.map((line) => (
                <Text key={line} size="sm">
                    {line}
                </Text>
            ))}
        </Stack>
    )
}

export function TeamMemberTaskGroups({ duty }: Props) {
    const [selectedMemberId, setSelectedMemberId] = useState<string | null>(null)
    const teamMembers = useMemo(() => selectSortedTeamMembers(duty), [duty])
    const selectedMemberSummary = selectedMemberId
        ? teamMembers.find((memberSummary) => memberSummary.member.id === selectedMemberId) ?? null
        : null
    const selectedMemberTaskGroups = useMemo(() => {
        if (!selectedMemberId) {
            return []
        }

        return groupAndSortMemberTasks(selectTasksForTeamMember(duty, selectedMemberId))
    }, [duty, selectedMemberId])

    if (teamMembers.length === 0) {
        return (
            <EmptyState
                title="Участники не найдены"
                description="В дежурной команде нет участников."
            />
        )
    }

    return (
        <>
            <Stack gap={0}>
                {teamMembers.map((memberSummary) => (
                    <TeamMemberCard
                        key={memberSummary.member.id}
                        memberSummary={memberSummary}
                        tooltip={<TooltipContent lines={memberSummary.tooltipLines} />}
                        onOpen={() => setSelectedMemberId(memberSummary.member.id)}
                    />
                ))}
            </Stack>

            <TeamMemberTasksDrawer
                opened={selectedMemberSummary !== null}
                memberSummary={selectedMemberSummary}
                taskGroups={selectedMemberTaskGroups}
                tooltip={<TooltipContent lines={selectedMemberSummary?.tooltipLines ?? []} />}
                onClose={() => setSelectedMemberId(null)}
            />
        </>
    )
}
