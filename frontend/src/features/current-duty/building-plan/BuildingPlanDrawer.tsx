import { Alert, Drawer, FocusTrap, ScrollArea, Table } from '@mantine/core'

import type { TaskActionHandler, TaskRowActionMode } from '../model/taskActions'
import type { ResidentDutyTask } from '../model/types'
import { TaskTableRow } from '../ui/TaskTableRow'
import type { FloorPlan } from './types'
import { getFloorLabel } from './utils'

type Props = {
    actionMode: TaskRowActionMode
    floorPlan: FloorPlan
    isReadOnly?: boolean
    opened: boolean
    pendingTaskId: string | null
    tasks: ResidentDutyTask[]
    onClose: () => void
    onTake: TaskActionHandler
    onReturn: TaskActionHandler
    onComplete: TaskActionHandler
    onOpen: TaskActionHandler
    onReopen?: TaskActionHandler
    onVerify?: TaskActionHandler
}

export function BuildingPlanDrawer({
    actionMode,
    floorPlan,
    isReadOnly = false,
    opened,
    pendingTaskId,
    tasks,
    onClose,
    onTake,
    onReturn,
    onComplete,
    onOpen,
    onReopen,
    onVerify,
}: Props) {
    const title = `${getFloorLabel(floorPlan.floor)} · ${tasks[0]?.area_name ?? 'Территория'}`

    return (
        <Drawer opened={opened} onClose={onClose} position="right" size={860} title={opened ? title : 'Территория'}>
            <FocusTrap.InitialFocus />

            {tasks.length === 0 ? (
                <Alert color="gray">Для этой территории нет задач.</Alert>
            ) : (
                <ScrollArea>
                    <Table miw={780}>
                        <Table.Tbody>
                            {tasks.map((task) => (
                                <TaskTableRow
                                    key={task.id}
                                    isReadOnly={isReadOnly}
                                    layout="drawer"
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
            )}
        </Drawer>
    )
}
