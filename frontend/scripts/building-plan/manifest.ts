export type BuildingPlanSource = {
    floor: number
    source: string
}

export const buildingPlanSources = [
    {
        floor: 1,
        source: 'assets/building-plans/dev/first-floor.svg',
    },
    {
        floor: 3,
        source: 'assets/building-plans/dev/third-floor.svg',
    },
] as const satisfies readonly BuildingPlanSource[]
