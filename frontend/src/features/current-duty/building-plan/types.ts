export type ShapeType = 'rect' | 'path' | 'polygon'

export type ShapeGeometry =
    | {
          x: number
          y: number
          width: number
          height: number
          rx?: number
      }
    | {
          d: string
      }
    | {
          points: string
      }

export type PlanShape = {
    id: string
    type: ShapeType
    geometry: ShapeGeometry
    fill?: string
    stroke?: string
}

export type PlanAreaShape = {
    id: string
    areaId: string
    type: ShapeType
    geometry: ShapeGeometry
    label?: {
        x: number
        y: number
    }
}

export type FloorPlanDefinition = {
    viewBox: {
        minX: number
        minY: number
        width: number
        height: number
    }
    background: PlanShape[]
    areas: PlanAreaShape[]
}

export type FloorPlan = {
    id: string
    floor: number
    name: string
    definition: FloorPlanDefinition
}
