import { Alert, Button, Stack } from '@mantine/core'

import type { DutySettingsArea } from '../model/types'
import { DutySettingsAreaSection } from './DutySettingsAreaSection'

type Props = {
    areas: DutySettingsArea[]
    onCreateArea: () => void
    onEditArea: (areaId: number) => void
    onDeleteArea: (areaId: number) => void
    onCreateTask: (areaId: number) => void
    onEditTask: (taskId: string) => void
    onDeleteTask: (taskId: string) => void
}

export function DutySettingsAreaList(props: Props) {
    if (props.areas.length === 0) {
        return (
            <Alert color="gray">
                <Stack gap="md">
                    <span>У этой группы пока нет территорий.</span>
                    <Button w="fit-content" onClick={props.onCreateArea}>
                        Добавить территорию
                    </Button>
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
                    onDeleteTask={props.onDeleteTask}
                />
            ))}
        </Stack>
    )
}
