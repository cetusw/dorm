import { Accordion, Stack, Text } from '@mantine/core'

import type { TaskAreaGroup } from '../model/utils'
import { TaskMobileCard } from './TaskMobileCard'
import type { TaskRowActionMode } from './TaskRowActions'

type Props = {
    actionMode?: TaskRowActionMode
    group: TaskAreaGroup
    isReadOnly?: boolean
    pendingTaskId: string | null
    onTake: (taskId: string) => void | Promise<unknown>
    onReturn: (taskId: string) => void | Promise<unknown>
    onComplete: (taskId: string) => void | Promise<unknown>
    onOpen: (taskId: string) => void | Promise<unknown>
    onVerify?: (taskId: string) => void | Promise<unknown>
}

export function TaskMobileGroupSection({
    actionMode = 'default',
    group,
    pendingTaskId,
    onTake,
    onReturn,
    onComplete,
    onOpen,
    onVerify,
}: Props) {
    return (
        <Accordion radius="lg" variant="contained">
            <Accordion.Item value={group.key}>
                <Accordion.Control px="md" py="sm">
                    <Text fw={700} size="sm">
                        {group.label}
                    </Text>
                </Accordion.Control>
                <Accordion.Panel px="sm" pb="sm">
                    <Stack gap="sm">
                        {group.tasks.map((task) => (
                            <TaskMobileCard
                                key={task.id}
                                mode={actionMode}
                                pending={pendingTaskId === task.id}
                                task={task}
                                onTake={onTake}
                                onReturn={onReturn}
                                onComplete={onComplete}
                                onOpen={onOpen}
                                onVerify={onVerify}
                            />
                        ))}
                    </Stack>
                </Accordion.Panel>
            </Accordion.Item>
        </Accordion>
    )
}
