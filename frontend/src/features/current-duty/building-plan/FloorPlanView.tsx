import { useState } from 'react'

import { Popover, Stack, Text } from '@mantine/core'

import type { AreaTaskSummary } from './planState'
import type { FloorPlan, PlanAreaShape, PlanShape, ShapeGeometry } from './types'
import classes from './FloorPlanView.module.css'

type Props = {
    plan: FloorPlan
    selectedAreaId: string | null
    summaries: Map<string, AreaTaskSummary>
    onAreaClick: (area: PlanAreaShape) => void
}

function isRectGeometry(geometry: ShapeGeometry): geometry is Extract<ShapeGeometry, { x: number }> {
    return 'x' in geometry
}

function isPathGeometry(geometry: ShapeGeometry): geometry is Extract<ShapeGeometry, { d: string }> {
    return 'd' in geometry
}

function Shape({ shape }: { shape: PlanShape }) {
    if (isRectGeometry(shape.geometry)) {
        return <rect {...shape.geometry} fill={shape.fill} stroke={shape.stroke} />
    }
    if (isPathGeometry(shape.geometry)) {
        return <path d={shape.geometry.d} fill={shape.fill} stroke={shape.stroke} />
    }

    return <polygon points={shape.geometry.points} fill={shape.fill} stroke={shape.stroke} />
}

type AreaShapeProps = {
    fill: string
    stroke: string
    strokeWidth: number
    className: string
    onClick: () => void
    onMouseEnter?: () => void
    onMouseLeave?: () => void
}

function AreaShape({ area, shapeProps }: { area: PlanAreaShape; shapeProps: AreaShapeProps }) {
    if (isRectGeometry(area.geometry)) {
        return <rect {...area.geometry} {...shapeProps} />
    }
    if (isPathGeometry(area.geometry)) {
        return <path d={area.geometry.d} {...shapeProps} />
    }

    return <polygon points={area.geometry.points} {...shapeProps} />
}

function AreaStatsPopover({ summary }: { summary: AreaTaskSummary }) {
    const { counters } = summary
    const taken = counters.assigned + counters.completed + counters.verified + counters.revision
    const done = counters.completed + counters.verified
    const firstTask = summary.tasks[0]
    const title = firstTask ? `${firstTask.area_floor} этаж · ${firstTask.area_name}` : 'Территория'

    return (
        <Stack gap={4}>
            <Text size="sm" fw={700}>{title}</Text>
            <Text size="sm">Взято {taken} из {counters.total} задач</Text>
            <Text size="sm">Выполнено {done} из {counters.total} задач</Text>
            <Text size="sm">Проверено {counters.verified} из {counters.total} задач</Text>
        </Stack>
    )
}

export function FloorPlanView({ plan, selectedAreaId, summaries, onAreaClick }: Props) {
    const { viewBox } = plan.definition
    const [hoveredAreaId, setHoveredAreaId] = useState<string | null>(null)

    return (
        <div className={classes.frame}>
            <svg
                className={classes.svg}
                viewBox={`${viewBox.minX} ${viewBox.minY} ${viewBox.width} ${viewBox.height}`}
                role="img"
                aria-label={plan.name}
            >
                {plan.definition.background.map((shape) => (
                    <Shape key={shape.id} shape={shape} />
                ))}

                {plan.definition.areas.map((area) => {
                    const summary = summaries.get(area.areaId)
                    const isInteractive = summary !== undefined
                    const selected = area.id === selectedAreaId
                    const shapeCommonProps = {
                        fill: '#CCFBF1',
                        stroke: '#14B8A6',
                        strokeWidth: selected ? 4 : 1,
                        className: [
                            classes.area,
                            isInteractive ? classes.interactive : classes.inactive,
                            selected ? classes.selected : '',
                        ].join(' '),
                        onMouseEnter: isInteractive ? () => setHoveredAreaId(area.id) : undefined,
                        onMouseLeave: isInteractive ? () => setHoveredAreaId(null) : undefined,
                        onClick: () => {
                            if (isInteractive) {
                                onAreaClick(area)
                            }
                        },
                    }

                    if (!summary) {
                        return (
                            <g key={area.id}>
                                <AreaShape area={area} shapeProps={shapeCommonProps} />
                            </g>
                        )
                    }

                    return (
                        <Popover
                            key={area.id}
                            opened={hoveredAreaId === area.id}
                            position="top"
                            withArrow
                            shadow="md"
                        >
                            <Popover.Target>
                                <g>
                                    <AreaShape area={area} shapeProps={shapeCommonProps} />
                                </g>
                            </Popover.Target>
                            <Popover.Dropdown>
                                <AreaStatsPopover summary={summary} />
                            </Popover.Dropdown>
                        </Popover>
                    )
                })}
            </svg>
        </div>
    )
}
