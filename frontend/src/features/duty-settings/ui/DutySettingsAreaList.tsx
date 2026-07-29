import { PlusIcon } from '@phosphor-icons/react'
import { ActionIcon, Alert, Group, Stack, Text } from '@mantine/core'
import type { DutySettingsArea } from '../model/types'
import { DutySettingsAreaSection } from './DutySettingsAreaSection'

type Props = {
    areas: DutySettingsArea[]
    onCreateArea: () => void
    onEditArea: (areaId: number) => void
    onDeleteArea: (areaId: number) => void
    onCreateTask: (areaId: number) => void
    onEditTask: (taskId: string) => void
    onIncludeTask: (taskId: string) => void
    onExcludeTask: (taskId: string) => void
    onDeleteTask: (taskId: string) => void
}

export function DutySettingsAreaList(props: Props) {
    if (props.areas.length === 0) {
        return (
            <Alert color="gray">
                <Stack gap="md">
                    <Group gap={15} wrap="nowrap" align="center">
                        <Text>В доступных территориях пока нет задач.</Text>
                        <ActionIcon
                            variant="subtle"
                            color="gray"
                            size={25}
                            aria-label="Добавить территорию"
                            onClick={props.onCreateArea}
                        >
                            <PlusIcon size={25} />
                        </ActionIcon>
                    </Group>
                </Stack>
            </Alert>
        )
    }

    return (
        <Stack gap="xl">
            {props.areas.map((area) => (
                <DutySettingsAreaSection
                    key={area.id}
                    area={area}
                    onCreateArea={props.onCreateArea}
                    onEditArea={props.onEditArea}
                    onDeleteArea={props.onDeleteArea}
                    onCreateTask={props.onCreateTask}
                    onEditTask={props.onEditTask}
                    onIncludeTask={props.onIncludeTask}
                    onExcludeTask={props.onExcludeTask}
                    onDeleteTask={props.onDeleteTask}
                />
            ))}
        </Stack>
    )
}
