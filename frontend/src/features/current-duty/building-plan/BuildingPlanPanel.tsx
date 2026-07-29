import { useEffect, useMemo, useState } from 'react'

import { Box, Stack } from '@mantine/core'

import type { TaskActionHandler } from '../model/taskActions'
import type { DutyTaskSelect, ResidentDutyTask } from '../model/types'
import type { TaskRowActionMode } from '../ui/TaskRowActions'
import { BuildingPlanDrawer } from './BuildingPlanDrawer'
import { FloorPlanView } from './FloorPlanView'
import { buildAreaTaskSummaries } from './planState'
import type { FloorPlan } from './types'

type Props = {
    activeSelect: DutyTaskSelect
    actionMode: TaskRowActionMode
    allTasks: ResidentDutyTask[]
    floorPlan: FloorPlan
    isReadOnly?: boolean
    pendingTaskId: string | null
    tasks: ResidentDutyTask[]
    onTake: TaskActionHandler
    onReturn: TaskActionHandler
    onComplete: TaskActionHandler
    onOpen: TaskActionHandler
    onReopen?: TaskActionHandler
    onVerify?: TaskActionHandler
}

export function BuildingPlanPanel({
    activeSelect,
    actionMode,
    allTasks,
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
    const [selectedAreaId, setSelectedAreaId] = useState<string | null>(null)
    const summaries = useMemo(
        () => buildAreaTaskSummaries(allTasks, tasks, activeSelect),
        [activeSelect, allTasks, tasks],
    )
    const selectedAreaTasks = selectedAreaId ? summaries.get(selectedAreaId)?.visibleTasks ?? [] : []

    useEffect(() => {
        if (selectedAreaId && (summaries.get(selectedAreaId)?.visibleTasks.length ?? 0) === 0) {
            setSelectedAreaId(null)
        }
    }, [selectedAreaId, summaries])

    function handleClose() {
        setSelectedAreaId(null)
    }

    return (
        <Box visibleFrom="md">
            <Stack gap="md">
                <FloorPlanView
                    plan={floorPlan}
                    selectedAreaId={selectedAreaId}
                    summaries={summaries}
                    onAreaClick={setSelectedAreaId}
                />
            </Stack>

            <BuildingPlanDrawer
                actionMode={actionMode}
                floorPlan={floorPlan}
                isReadOnly={isReadOnly}
                opened={selectedAreaId !== null}
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
