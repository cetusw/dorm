import { Box, ScrollArea, Table, Text } from '@mantine/core'

import type { TaskActionHandler, TaskRowActionMode } from '../model/taskActions'
import type { TaskAreaGroup } from '../model/utils'
import { TaskTableRow } from './TaskTableRow'

type Props = {
    actionMode?: TaskRowActionMode
    group: TaskAreaGroup
    isReadOnly?: boolean
    pendingTaskId: string | null
    onTake: TaskActionHandler
    onReturn: TaskActionHandler
    onComplete: TaskActionHandler
    onOpen: TaskActionHandler
    onReopen?: TaskActionHandler
    onVerify?: TaskActionHandler
}

export function TaskGroupSection({
    actionMode = 'default',
    group,
    isReadOnly = false,
    pendingTaskId,
    onTake,
    onReturn,
    onComplete,
    onOpen,
    onReopen,
    onVerify,
}: Props) {
    return (
        <Box>
            <Text fw={700} size="lg" pb="md">
                {group.label}
            </Text>

            <ScrollArea>
                <Table miw={780}>
                    <Table.Tbody>
                        {group.tasks.map((task) => (
                            <TaskTableRow
                                key={task.id}
                                isReadOnly={isReadOnly}
                                mode={actionMode}
                                pending={pendingTaskId === task.id}
                                task={task}
                                onTake={onTake}
                                onReturn={onReturn}
                                onComplete={onComplete}
                                onOpen={onOpen}
                                onReopen={onReopen}
                                onVerify={onVerify}
                            />
                        ))}
                    </Table.Tbody>
                </Table>
            </ScrollArea>
        </Box>
    )
}
