import { Checkbox, Paper, ScrollArea, Table, Text } from '@mantine/core'

import type { TeamMemberOption } from '../model/types'

type Props = {
    members: TeamMemberOption[]
    selectedMemberIds: string[]
    onToggleMember: (memberId: string) => void
}

export function TeamMembersSelection({ members, selectedMemberIds, onToggleMember }: Props) {
    return (
        <Paper withBorder radius="md" bg="white">
            <ScrollArea h={280}>
                <Table horizontalSpacing="md" verticalSpacing="sm" highlightOnHover>
                    <Table.Thead>
                        <Table.Tr>
                            <Table.Th w={44} />
                            <Table.Th>
                                <Text size="sm" fw={700}>
                                    Участники команды
                                </Text>
                            </Table.Th>
                        </Table.Tr>
                    </Table.Thead>
                    <Table.Tbody>
                        {members.map((member) => {
                            const checked = selectedMemberIds.includes(member.id)
                            const inOtherTeam =
                                Boolean(member.current_team_name) && !member.is_in_current_team

                            return (
                                <Table.Tr
                                    key={member.id}
                                    onClick={() => onToggleMember(member.id)}
                                    style={{ cursor: 'pointer' }}
                                >
                                    <Table.Td onClick={(event) => event.stopPropagation()}>
                                        <Checkbox
                                            checked={checked}
                                            onChange={() => onToggleMember(member.id)}
                                            aria-label={`Выбрать ${member.name}`}
                                        />
                                    </Table.Td>
                                    <Table.Td>
                                        <Text fw={500}>{member.name}</Text>
                                        {inOtherTeam && (
                                            <Text size="sm" c="red">
                                                Состоит в команде {member.current_team_name}
                                            </Text>
                                        )}
                                    </Table.Td>
                                </Table.Tr>
                            )
                        })}
                    </Table.Tbody>
                </Table>
            </ScrollArea>
        </Paper>
    )
}
