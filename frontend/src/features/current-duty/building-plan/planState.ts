import type { ResidentDutyTask } from '../model/types'

export type AreaTaskCounters = {
    total: number
    free: number
    assigned: number
    completed: number
    verified: number
    revision: number
}

export type AreaUserCounters = {
    taken: number
    assigned: number
    completedOrVerified: number
    verified: number
}

export type AreaPresentationState =
    | 'other-group'
    | 'no-free'
    | 'has-free'
    | 'mine-assigned'
    | 'mine-completed'
    | 'mine-verified'

export type AreaTaskSummary = {
    counters: AreaTaskCounters
    userCounters: AreaUserCounters
    state: AreaPresentationState
    tasks: ResidentDutyTask[]
}

export type AreaStyle = {
    fill: string
    stroke: string
}

export const areaStyles: Record<AreaPresentationState, AreaStyle> = {
    'other-group': {
        fill: '#F0FDFA',
        stroke: '#99F6E4',
    },
    'no-free': {
        fill: '#F0FDFA',
        stroke: '#99F6E4',
    },
    'has-free': {
        fill: '#CCFBF1',
        stroke: '#14B8A6',
    },
    'mine-assigned': {
        fill: '#FEF3C7',
        stroke: '#92400E',
    },
    'mine-completed': {
        fill: '#E0F2FE',
        stroke: '#0369A1',
    },
    'mine-verified': {
        fill: '#DCFCE7',
        stroke: '#166534',
    },
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
        const userCounters = countUserTasks(areaTasks)

        summaries.set(areaId, {
            counters,
            userCounters,
            state: resolveAreaPresentationState(counters, userCounters),
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

export function countUserTasks(tasks: ResidentDutyTask[]): AreaUserCounters {
    const myTasks = tasks.filter((task) => task.is_mine)

    return {
        taken: myTasks.length,
        assigned: myTasks.filter((task) => task.status === 'assigned').length,
        completedOrVerified: myTasks.filter(
            (task) => task.status === 'completed' || task.status === 'verified',
        ).length,
        verified: myTasks.filter((task) => task.status === 'verified').length,
    }
}

export function resolveAreaPresentationState(
    counters: AreaTaskCounters,
    userCounters: AreaUserCounters,
): AreaPresentationState {
    if (userCounters.taken > 0 && userCounters.verified === userCounters.taken) {
        return 'mine-verified'
    }

    if (userCounters.taken > 0 && userCounters.completedOrVerified === userCounters.taken) {
        return 'mine-completed'
    }

    if (userCounters.taken > 0) {
        return 'mine-assigned'
    }

    if (counters.free > 0) {
        return 'has-free'
    }

    return 'no-free'
}
