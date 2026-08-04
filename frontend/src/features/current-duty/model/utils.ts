import type { ResidentDutyTask } from './types'

export type TaskAreaGroup = {
    key: string
    label: string
    tasks: ResidentDutyTask[]
}

export function groupTasksByArea(tasks: ResidentDutyTask[]): TaskAreaGroup[] {
    const groups = new Map<string, TaskAreaGroup>()

    tasks.forEach((task) => {
        const label = `${task.area_floor} этаж · ${task.area_name}`
        const key = `${task.area_floor}:${task.area_name}`
        const existing = groups.get(key)

        if (existing) {
            existing.tasks.push(task)
            return
        }

        groups.set(key, {
            key,
            label,
            tasks: [task],
        })
    })

    return Array.from(groups.values())
}

export function toErrorMessage(error: unknown): string {
    return error instanceof Error ? error.message : 'Неизвестная ошибка'
}

function taskStatusPriority(task: ResidentDutyTask): number {
    switch (task.status) {
        case 'free':
            return 0
        case 'assigned':
            return 1
        case 'completed':
            return 2
        case 'verified':
            return 3
        default:
            return 4
    }
}

export function sortTasksForInitialDisplay(tasks: ResidentDutyTask[]): ResidentDutyTask[] {
    return [...tasks].sort((left, right) => {
        if (left.area_floor !== right.area_floor) {
            return right.area_floor - left.area_floor
        }

        if (left.area_name !== right.area_name) {
            return left.area_name.localeCompare(right.area_name, 'ru')
        }

        const priorityDiff = taskStatusPriority(left) - taskStatusPriority(right)
        if (priorityDiff !== 0) {
            return priorityDiff
        }

        if (left.cost !== right.cost) {
            return right.cost - left.cost
        }

        return left.title.localeCompare(right.title, 'ru')
    })
}

export function preserveTaskOrder(
    tasks: ResidentDutyTask[],
    orderedTaskIds: string[],
): ResidentDutyTask[] {
    const orderMap = new Map(orderedTaskIds.map((taskId, index) => [taskId, index]))

    return [...tasks].sort((left, right) => {
        const leftOrder = orderMap.get(left.id)
        const rightOrder = orderMap.get(right.id)

        if (leftOrder === undefined && rightOrder === undefined) {
            return 0
        }
        if (leftOrder === undefined) {
            return 1
        }
        if (rightOrder === undefined) {
            return -1
        }

        return leftOrder - rightOrder
    })
}

function formatDutyDate(value: string, includeYear: boolean): string {
    const [year, month, day] = value.split('-')
    if (!year || !month || !day) {
        return value
    }

    return includeYear ? `${day}.${month}.${year}` : `${day}.${month}`
}

export function formatDutyPeriod(startDate: string, endDate: string): string {
    return `${formatDutyDate(startDate, false)} – ${formatDutyDate(endDate, false)}`
}

export function formatDutyPeriodFull(startDate: string, endDate: string): string {
    return `${formatDutyDate(startDate, true)} – ${formatDutyDate(endDate, true)}`
}

export function countTasksCompletedByMe(tasks: ResidentDutyTask[]): number {
    return tasks.filter(
        (task) => task.is_mine && (task.status === 'completed' || task.status === 'verified'),
    ).length
}

export function countTasksTakenByMe(tasks: ResidentDutyTask[]): number {
    return tasks.filter((task) => task.is_mine).length
}

export function sumCostTakenByMe(tasks: ResidentDutyTask[]): number {
    return tasks
        .filter((task) => task.is_mine)
        .reduce((sum, task) => sum + task.cost, 0)
}

export function sumCostCompletedByMe(tasks: ResidentDutyTask[]): number {
    return tasks
        .filter((task) => task.is_mine && (task.status === 'completed' || task.status === 'verified'))
        .reduce((sum, task) => sum + task.cost, 0)
}

export function sumCostVerifiedByMe(tasks: ResidentDutyTask[]): number {
    return tasks
        .filter((task) => task.is_mine && task.status === 'verified')
        .reduce((sum, task) => sum + task.cost, 0)
}
