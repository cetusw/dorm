import { Alert, Stack } from '@mantine/core'

import type { DutySettingsArea } from '../model/types'
import { DutySettingsAreaSection } from './DutySettingsAreaSection'

type Props = {
    areas: DutySettingsArea[]
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
                    <span>В доступных территориях пока нет задач.</span>
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
