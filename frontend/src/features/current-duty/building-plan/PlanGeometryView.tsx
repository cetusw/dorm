import { useMemo, useState, type ReactNode } from 'react'

import { Popover, Text } from '@mantine/core'

import { getFloorLabel } from './utils'
import type { AreaShape, BackgroundShape, FloorPlan, Shape } from './types'
import classes from './FloorPlanView.module.css'

type ShapeProps = {
    fill?: string
    stroke?: string
    strokeWidth?: number
    className?: string
}

export type PlanAreaPresentation = {
    areaId: string
    fill: string
    stroke: string
    interactive: boolean
    selected?: boolean
    popoverContent?: ReactNode
}

type Props = {
    plan: FloorPlan
    areas: PlanAreaPresentation[]
    selectedAreaId: string | null
    onAreaClick: (areaId: string) => void
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

export function PlanGeometryView({ plan, areas, selectedAreaId, onAreaClick }: Props) {
    const [hoveredAreaId, setHoveredAreaId] = useState<string | null>(null)
    const areaGroups = useMemo(() => {
        const groupedAreas = new Map<string, AreaShape[]>()

        for (const area of plan.areas) {
            groupedAreas.set(area.areaId, [...(groupedAreas.get(area.areaId) ?? []), area])
        }

        return Array.from(groupedAreas.entries()).map(([areaId, shapes]) => ({ areaId, shapes }))
    }, [plan.areas])

    const presentationByArea = useMemo(
        () => new Map(areas.map((area) => [area.areaId, area])),
        [areas],
    )

    return (
        <div className={classes.frame}>
            <svg
                className={classes.svg}
                viewBox={`${plan.viewBox.minX} ${plan.viewBox.minY} ${plan.viewBox.width} ${plan.viewBox.height}`}
                role="img"
                aria-label={getFloorLabel(plan.floor)}
            >
                {plan.background.map((shape, index) => renderBackgroundShape(shape, index))}

                {areaGroups.map(({ areaId, shapes }) => {
                    const presentation = presentationByArea.get(areaId)
                    const interactive = presentation?.interactive ?? false
                    const selected = presentation?.selected ?? areaId === selectedAreaId
                    const shapeProps: ShapeProps = {
                        fill: presentation?.fill ?? '#F0FDFA',
                        stroke: presentation?.stroke ?? '#99F6E4',
                        strokeWidth: selected ? 4 : 1,
                        className: [
                            classes.area,
                            interactive ? classes.interactive : classes.inactive,
                            selected ? classes.selected : '',
                        ].join(' '),
                    }

                    const target = (
                        <g
                            onMouseEnter={() => setHoveredAreaId(areaId)}
                            onMouseLeave={() => setHoveredAreaId(null)}
                            onClick={() => {
                                if (interactive) {
                                    onAreaClick(areaId)
                                }
                            }}
                        >
                            {shapes.map((shape, index) => renderAreaShape(shape, index, shapeProps))}
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
                                {presentation?.popoverContent ?? <Text size="sm">За эту территорию отвечает другая группа</Text>}
                            </Popover.Dropdown>
                        </Popover>
                    )
                })}
            </svg>
        </div>
    )
}
