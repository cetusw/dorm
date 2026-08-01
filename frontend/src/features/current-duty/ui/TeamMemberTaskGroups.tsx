import { Accordion, Alert, Box, Group, Stack, Text } from '@mantine/core'

import type { TeamMemberTaskGroup } from '../model/selectors'
import { MemberDutyProgressCard } from './MemberDutyProgressCard'
import { TeamMemberTaskList } from './TeamMemberTaskList'

type Props = {
    groups: TeamMemberTaskGroup[]
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

export function TeamMemberTaskGroups({ groups }: Props) {
    return (
        <Accordion
            radius="lg"
            variant="separated"
            styles={{
                item: {
                    backgroundColor: 'var(--app-color-surface)',
                    border: '1px solid var(--app-color-border)',
                    borderRadius: '16px',
                    overflow: 'hidden',
                },
                control: {
                    backgroundColor: 'var(--app-color-surface)',
                    paddingTop: 14,
                    paddingBottom: 14,
                    paddingLeft: 20,
                    paddingRight: 16,
                    borderRadius: '16px',
                },
                chevron: {
                    marginInlineStart: 12,
                },
                panel: {
                    backgroundColor: 'var(--app-color-surface)',
                    paddingTop: 0,
                    borderBottomLeftRadius: '16px',
                    borderBottomRightRadius: '16px',
                },
                content: {
                    paddingRight: 0,
                },
            }}
        >
            {groups.map((group) => (
                <Accordion.Item key={group.member.id} value={group.member.id}>
                    <Accordion.Control>
                        <Group justify="space-between" wrap="nowrap" gap="md">
                            <Text fw={700} size="md" style={{ flex: 1, minWidth: 0 }}>
                                {group.member.name}
                            </Text>
                            <Box style={{ width: 216, maxWidth: '100%', marginLeft: 'auto' }}>
                                <MemberDutyProgressCard
                                    progress={group.progress}
                                    showTitle={false}
                                    withContainer={false}
                                    tooltip={<TooltipContent lines={group.tooltipLines} />}
                                    barWidth="100%"
                                />
                            </Box>
                        </Group>
                    </Accordion.Control>
                    <Accordion.Panel px="lg" pb="lg">
                        {group.tasks.length === 0 ? (
                            <Alert color="gray">Участник не взял ни одной задачи</Alert>
                        ) : (
                            <TeamMemberTaskList tasks={group.tasks} />
                        )}
                    </Accordion.Panel>
                </Accordion.Item>
            ))}
        </Accordion>
    )
}
