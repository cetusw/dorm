import { useMemo } from 'react'

import { Stack, Text } from '@mantine/core'

import { areaStyles, type AreaTaskSummary } from './planState'
import { PlanGeometryView, type PlanAreaPresentation } from './PlanGeometryView'
import type { FloorPlan } from './types'

type Props = {
    plan: FloorPlan
    selectedAreaId: string | null
    summaries: Map<string, AreaTaskSummary>
    onAreaClick: (areaId: string) => void
}

function AreaStatsPopover({ summary }: { summary: AreaTaskSummary }) {
    const { counters, emptyMessage, hasVisibleTasks } = summary
    const taken = counters.assigned + counters.completed + counters.verified + counters.revision
    const done = counters.completed + counters.verified
    const firstTask = summary.tasks[0]
    const title = firstTask ? `${firstTask.area_floor} этаж . ${firstTask.area_name}` : 'Территория'

    if (!hasVisibleTasks) {
        return (
            <Stack gap={4}>
                <Text size="sm" fw={700}>
                    {title}
                </Text>
                <Text size="sm">{emptyMessage}</Text>
            </Stack>
        )
    }

    return (
        <Stack gap={4}>
            <Text size="sm" fw={700}>
                {title}
            </Text>
            <Text size="sm">Взято {taken} из {counters.total} задач</Text>
            <Text size="sm">Выполнено {done} из {counters.total} задач</Text>
            <Text size="sm">Проверено {counters.verified} из {counters.total} задач</Text>
        </Stack>
    )
}

export function FloorPlanView({ plan, selectedAreaId, summaries, onAreaClick }: Props) {
    const areas = useMemo<PlanAreaPresentation[]>(
        () => plan.areas.map((shape) => {
            const summary = summaries.get(shape.areaId)
            const state = summary?.state ?? 'muted'
            const style = areaStyles[state]

            return {
                areaId: shape.areaId,
                fill: style.fill,
                stroke: style.stroke,
                interactive: summary?.hasVisibleTasks ?? false,
                selected: shape.areaId === selectedAreaId,
                popoverContent: summary ? <AreaStatsPopover summary={summary} /> : (
                    <Text size="sm">За эту территорию отвечает другая группа</Text>
                ),
            }
        }),
        [plan.areas, selectedAreaId, summaries],
    )

    return (
        <PlanGeometryView
            plan={plan}
            areas={areas}
            selectedAreaId={selectedAreaId}
            onAreaClick={onAreaClick}
        />
    )
}
