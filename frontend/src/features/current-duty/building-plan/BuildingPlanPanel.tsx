import { useEffect, useMemo, useState } from 'react'

import { Box, Stack } from '@mantine/core'

import type { ResidentDutyTask } from '../model/types'
import type { TaskRowActionMode } from '../ui/TaskRowActions'
import { BuildingPlanDrawer } from './BuildingPlanDrawer'
import { FloorPlanView } from './FloorPlanView'
import { buildAreaTaskSummaries } from './planState'
import type { FloorPlan, PlanAreaShape } from './types'

type Props = {
    actionMode: TaskRowActionMode
    floorPlan: FloorPlan
    isReadOnly?: boolean
    pendingTaskId: string | null
    tasks: ResidentDutyTask[]
    onTake: (taskId: string) => void | Promise<unknown>
    onReturn: (taskId: string) => void | Promise<unknown>
    onComplete: (taskId: string) => void | Promise<unknown>
    onOpen: (taskId: string) => void | Promise<unknown>
    onReopen?: (taskId: string) => void | Promise<unknown>
    onVerify?: (taskId: string) => void | Promise<unknown>
}

export function BuildingPlanPanel({
    actionMode,
    floorPlan,
    isReadOnly = false,
    pendingTaskId,
    tasks,
    onTake,
    onReturn,
    onComplete,
    onOpen,
    onReopen,
    onVerify,
}: Props) {
    const [selectedArea, setSelectedArea] = useState<PlanAreaShape | null>(null)
    const summaries = useMemo(() => buildAreaTaskSummaries(tasks), [tasks])
    const selectedAreaTasks = selectedArea ? summaries.get(selectedArea.areaId)?.tasks ?? [] : []

    useEffect(() => {
        if (selectedArea && !summaries.has(selectedArea.areaId)) {
            setSelectedArea(null)
        }
    }, [selectedArea, summaries])

    function handleClose() {
        setSelectedArea(null)
    }

    return (
        <Box visibleFrom="md">
            <Stack gap="md">
                <FloorPlanView
                    plan={floorPlan}
                    selectedAreaId={selectedArea?.id ?? null}
                    summaries={summaries}
                    onAreaClick={setSelectedArea}
                />
            </Stack>

            <BuildingPlanDrawer
                actionMode={actionMode}
                area={selectedArea}
                floorPlan={floorPlan}
                isReadOnly={isReadOnly}
                opened={selectedArea !== null}
                pendingTaskId={pendingTaskId}
                tasks={selectedAreaTasks}
                onClose={handleClose}
                onTake={onTake}
                onReturn={onReturn}
                onComplete={onComplete}
                onOpen={onOpen}
                onReopen={onReopen}
                onVerify={onVerify}
            />
        </Box>
    )
}
