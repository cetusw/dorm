import { useMemo } from 'react'

import { Text } from '@mantine/core'

import { areaStyles } from '../../current-duty/building-plan/planState'
import { PlanGeometryView, type PlanAreaPresentation } from '../../current-duty/building-plan/PlanGeometryView'
import type { FloorPlan } from '../../current-duty/building-plan/types'
import type { DutySettingsArea } from '../model/types'
import { formatDutySettingsAreaTitle } from '../model/utils'

type Props = {
    areas: DutySettingsArea[]
    floorPlan: FloorPlan
    selectedAreaId: string | null
    onAreaClick: (areaId: string) => void
}

export function DutySettingsPlanPanel({ areas, floorPlan, selectedAreaId, onAreaClick }: Props) {
    const areasById = useMemo(
        () => new Map(areas.map((area) => [String(area.id), area])),
        [areas],
    )

    const planAreas = useMemo<PlanAreaPresentation[]>(
        () => floorPlan.areas.map((shape) => {
            const area = areasById.get(shape.areaId)
            const isActive = Boolean(area)
            const style = isActive ? areaStyles.info : areaStyles.muted

            return {
                areaId: shape.areaId,
                fill: style.fill,
                stroke: style.stroke,
                interactive: isActive,
                selected: shape.areaId === selectedAreaId,
                popoverContent: area ? (
                    <Text size="sm" fw={700}>{formatDutySettingsAreaTitle(area)}</Text>
                ) : (
                    <Text size="sm">Территория не относится к текущей группе</Text>
                ),
            }
        }),
        [areasById, floorPlan.areas, selectedAreaId],
    )

    return (
        <PlanGeometryView
            plan={floorPlan}
            areas={planAreas}
            selectedAreaId={selectedAreaId}
            onAreaClick={onAreaClick}
        />
    )
}
