import type { DutyTaskSelect, ResidentDutyTask } from '../model/types'

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

export type AreaPresentationState = 'muted' | 'info' | 'success' | 'warning' | 'danger'

export type AreaTaskSummary = {
    counters: AreaTaskCounters
    visibleCounters: AreaTaskCounters
    userCounters: AreaUserCounters
    state: AreaPresentationState
    tasks: ResidentDutyTask[]
    visibleTasks: ResidentDutyTask[]
    hasVisibleTasks: boolean
    emptyMessage: string
}

export type AreaStyle = {
    fill: string
    stroke: string
}

export const areaStyles: Record<AreaPresentationState, AreaStyle> = {
    muted: {
        fill: '#F0FDFA',
        stroke: '#99F6E4',
    },
    info: {
        fill: '#E0F2FE',
        stroke: '#0369A1',
    },
    success: {
        fill: '#DCFCE7',
        stroke: '#166534',
    },
    warning: {
        fill: '#FEF3C7',
        stroke: '#92400E',
    },
    danger: {
        fill: '#FEE2E2',
        stroke: '#991B1B',
    },
}

export function buildAreaTaskSummaries(
    allTasks: ResidentDutyTask[],
    visibleTasks: ResidentDutyTask[],
    activeSelect: DutyTaskSelect,
): Map<string, AreaTaskSummary> {
    const groupedAllTasks = new Map<string, ResidentDutyTask[]>()
    const groupedVisibleTasks = new Map<string, ResidentDutyTask[]>()

    for (const task of allTasks) {
        const areaId = String(task.area_id)
        groupedAllTasks.set(areaId, [...(groupedAllTasks.get(areaId) ?? []), task])
    }

    for (const task of visibleTasks) {
        const areaId = String(task.area_id)
        groupedVisibleTasks.set(areaId, [...(groupedVisibleTasks.get(areaId) ?? []), task])
    }

    const summaries = new Map<string, AreaTaskSummary>()

    for (const [areaId, areaTasks] of groupedAllTasks.entries()) {
        const areaVisibleTasks = groupedVisibleTasks.get(areaId) ?? []
        const counters = countAreaTasks(areaTasks)
        const visibleCounters = countAreaTasks(areaVisibleTasks)
        const userCounters = countUserTasks(areaVisibleTasks)

        summaries.set(areaId, {
            counters,
            visibleCounters,
            userCounters,
            state: resolveAreaPresentationState(activeSelect, counters, visibleCounters, userCounters),
            tasks: areaTasks,
            visibleTasks: areaVisibleTasks,
            hasVisibleTasks: areaVisibleTasks.length > 0,
            emptyMessage: getAreaEmptyMessage(activeSelect),
        })
    }

    return summaries
}

export function getAreaEmptyMessage(activeSelect: DutyTaskSelect): string {
    switch (activeSelect) {
        case 'mine':
            return 'В этой территории нет ваших задач'
        case 'free':
            return 'В этой территории нет свободных задач'
        case 'verification':
            return 'В этой территории нет задач на проверке'
        case 'all':
        default:
            return 'За эту территорию отвечает другая группа'
    }
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
    activeSelect: DutyTaskSelect,
    counters: AreaTaskCounters,
    visibleCounters: AreaTaskCounters,
    userCounters: AreaUserCounters,
): AreaPresentationState {
    if (visibleCounters.total === 0) {
        return 'muted'
    }

    switch (activeSelect) {
        case 'mine':
            if (userCounters.taken > 0 && userCounters.completedOrVerified === userCounters.taken) {
                return 'success'
            }

            return 'info'
        case 'free':
            return 'info'
        case 'all':
        case 'verification':
            if (counters.verified === counters.total) {
                return 'success'
            }

            if (counters.free > 0 || counters.assigned > 0 || counters.revision > 0) {
                return 'danger'
            }

            if (counters.completed > 0) {
                return 'warning'
            }

            return 'muted'
        default:
            return 'muted'
    }
}
