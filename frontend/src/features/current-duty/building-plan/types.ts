export type ViewBox = {
    minX: number
    minY: number
    width: number
    height: number
}

export type RectShape = {
    type: 'rect'
    x: number
    y: number
    width: number
    height: number
    rx?: number
}

export type PathShape = {
    type: 'path'
    d: string
}

export type Shape = RectShape | PathShape

export type BackgroundStyle = {
    fill?: string
    stroke?: string
    strokeWidth?: number
}

export type BackgroundShape =
    | (RectShape & BackgroundStyle)
    | (PathShape & BackgroundStyle)

export type AreaShape =
    | (RectShape & { areaId: string })
    | (PathShape & { areaId: string })

export type FloorPlanGeometry = {
    viewBox: ViewBox
    background: BackgroundShape[]
    areas: AreaShape[]
}

export type FloorPlan = FloorPlanGeometry & {
    floor: number
}
