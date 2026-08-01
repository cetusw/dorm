import { mkdir, readFile, writeFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import type { FloorPlan } from '../../src/features/current-duty/building-plan/types.ts'
import { buildingPlanSources, type BuildingPlanSource } from './manifest.ts'
import { parseSvgPlan } from './parseSvgPlan.ts'

const currentFilePath = fileURLToPath(import.meta.url)
const currentDirectory = path.dirname(currentFilePath)
const frontendRoot = path.resolve(currentDirectory, '../..')
const generatedDir = path.resolve(frontendRoot, 'src/features/current-duty/building-plan/generated')
const generatedFilePath = path.resolve(generatedDir, 'plans.ts')

function serializePlans(floorPlans: FloorPlan[]): string {
    return `
import type { FloorPlan } from '../types'

export const floorPlans: FloorPlan[] = ${JSON.stringify(floorPlans, null, 4)}
`
}

function resolveSvgPath(sourcePath: string): string {
    return path.resolve(frontendRoot, sourcePath)
}

function validateBuildingPlanSources(sources: readonly BuildingPlanSource[]): void {
    const seenFloors = new Set<number>()

    for (const source of sources) {
        if (!Number.isInteger(source.floor) || source.floor <= 0) {
            throw new Error(`Invalid floor number in building plan manifest: ${String(source.floor)}`)
        }

        if (seenFloors.has(source.floor)) {
            throw new Error(`Duplicate floor number in building plan manifest: ${String(source.floor)}`)
        }

        if (source.source.trim() === '') {
            throw new Error(`Empty SVG source path in building plan manifest for floor ${String(source.floor)}`)
        }

        seenFloors.add(source.floor)
    }
}

async function loadFloorPlan(planSource: BuildingPlanSource): Promise<FloorPlan> {
    const svgPath = resolveSvgPath(planSource.source)
    const svgSource = await readFile(svgPath, 'utf8')
    const geometry = parseSvgPlan(svgSource, svgPath)

    return {
        floor: planSource.floor,
        ...geometry,
    }
}

async function main(): Promise<void> {
    validateBuildingPlanSources(buildingPlanSources)

    await mkdir(generatedDir, { recursive: true })

    const floorPlans = await Promise.all(buildingPlanSources.map(loadFloorPlan))
    floorPlans.sort((left, right) => left.floor - right.floor)

    await writeFile(generatedFilePath, serializePlans(floorPlans), 'utf8')
}

void main().catch((error) => {
    console.error(error instanceof Error ? error.message : String(error))
    process.exitCode = 1
})
