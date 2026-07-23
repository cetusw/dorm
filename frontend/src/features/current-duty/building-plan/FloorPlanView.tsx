import { useMemo, useState } from 'react'

import { Popover, Stack, Text } from '@mantine/core'

import { areaStyles, type AreaTaskSummary } from './planState'
import { getFloorLabel } from './utils'
import type { AreaShape, BackgroundShape, FloorPlan, Shape } from './types'
import classes from './FloorPlanView.module.css'

type Props = {
    plan: FloorPlan
    selectedAreaId: string | null
    summaries: Map<string, AreaTaskSummary>
    onAreaClick: (areaId: string) => void
}

type ShapeProps = {
    fill?: string
    stroke?: string
    strokeWidth?: number
    className?: string
}

function renderShape(shape: Shape, shapeProps?: ShapeProps) {
    switch (shape.type) {
        case 'rect':
            return <rect x={shape.x} y={shape.y} width={shape.width} height={shape.height} rx={shape.rx} {...shapeProps} />
        case 'path':
            return <path d={shape.d} {...shapeProps} />
    }

    return null
}

function renderBackgroundShape(shape: BackgroundShape, index: number) {
    return (
        <g key={`${shape.type}-${index}`}>
            {renderShape(shape, {
                fill: shape.fill,
                stroke: shape.stroke,
                strokeWidth: shape.strokeWidth,
            })}
        </g>
    )
}

function renderAreaShape(shape: AreaShape, index: number, shapeProps: ShapeProps) {
    return <g key={`${shape.areaId}-${index}`}>{renderShape(shape, shapeProps)}</g>
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
    const [hoveredAreaId, setHoveredAreaId] = useState<string | null>(null)
    const areaGroups = useMemo(() => {
        const groupedAreas = new Map<string, AreaShape[]>()

        for (const area of plan.areas) {
            groupedAreas.set(area.areaId, [...(groupedAreas.get(area.areaId) ?? []), area])
        }

        return Array.from(groupedAreas.entries()).map(([areaId, areas]) => ({ areaId, areas }))
    }, [plan.areas])

    return (
        <div className={classes.frame}>
            <svg
                className={classes.svg}
                viewBox={`${plan.viewBox.minX} ${plan.viewBox.minY} ${plan.viewBox.width} ${plan.viewBox.height}`}
                role="img"
                aria-label={getFloorLabel(plan.floor)}
            >
                {plan.background.map((shape, index) => renderBackgroundShape(shape, index))}

                {areaGroups.map(({ areaId, areas }) => {
                    const summary = summaries.get(areaId)
                    const state = summary?.state ?? 'muted'
                    const style = areaStyles[state]
                    const isClickable = summary?.hasVisibleTasks ?? false
                    const selected = areaId === selectedAreaId
                    const shapeProps: ShapeProps = {
                        fill: style.fill,
                        stroke: style.stroke,
                        strokeWidth: selected ? 4 : 1,
                        className: [
                            classes.area,
                            isClickable ? classes.interactive : classes.inactive,
                            selected ? classes.selected : '',
                        ].join(' '),
                    }

                    const target = (
                        <g
                            onMouseEnter={() => setHoveredAreaId(areaId)}
                            onMouseLeave={() => setHoveredAreaId(null)}
                            onClick={() => {
                                if (isClickable) {
                                    onAreaClick(areaId)
                                }
                            }}
                        >
                            {areas.map((shape, index) => renderAreaShape(shape, index, shapeProps))}
                        </g>
                    )

                    return (
                        <Popover
                            key={areaId}
                            opened={hoveredAreaId === areaId}
                            position="top"
                            withArrow
                            shadow="md"
                        >
                            <Popover.Target>{target}</Popover.Target>
                            <Popover.Dropdown>
                                {summary ? (
                                    <AreaStatsPopover summary={summary} />
                                ) : (
                                    <Text size="sm">За эту территорию отвечает другая группа</Text>
                                )}
                            </Popover.Dropdown>
                        </Popover>
                    )
                })}
            </svg>
        </div>
    )
}
