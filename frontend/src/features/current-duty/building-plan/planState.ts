import type { ResidentDutyTask } from '../model/types'

export type AreaTaskCounters = {
    total: number
    free: number
    assigned: number
    completed: number
    verified: number
    revision: number
}

export type AreaTaskSummary = {
    counters: AreaTaskCounters
    tasks: ResidentDutyTask[]
}

export function buildAreaTaskSummaries(tasks: ResidentDutyTask[]): Map<string, AreaTaskSummary> {
    const groups = new Map<string, ResidentDutyTask[]>()

    tasks.forEach((task) => {
        const areaId = String(task.area_id)
        groups.set(areaId, [...(groups.get(areaId) ?? []), task])
    })

    const summaries = new Map<string, AreaTaskSummary>()

    groups.forEach((areaTasks, areaId) => {
        const counters = countAreaTasks(areaTasks)

        summaries.set(areaId, {
            counters,
            tasks: areaTasks,
        })
    })

    return summaries
}

export function countAreaTasks(tasks: ResidentDutyTask[]): AreaTaskCounters {
    return {
        total: tasks.length,
        free: tasks.filter((task) => task.status === 'free').length,
        assigned: tasks.filter((task) => task.status === 'assigned' && !task.needs_revision).length,
        completed: tasks.filter((task) => task.status === 'completed').length,
        verified: tasks.filter((task) => task.status === 'verified').length,
        revision: tasks.filter((task) => task.needs_revision).length,
    }
}
