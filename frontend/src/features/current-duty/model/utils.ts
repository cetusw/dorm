import type { ResidentDutyTask } from './types'

export type TaskAreaGroup = {
    key: string
    label: string
    tasks: ResidentDutyTask[]
}

export type InitialTaskOrderMode = 'all' | 'mine'

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

function compareTaskArea(left: ResidentDutyTask, right: ResidentDutyTask): number {
    if (left.area_floor !== right.area_floor) {
        return right.area_floor - left.area_floor
    }

    if (left.area_name !== right.area_name) {
        return left.area_name.localeCompare(right.area_name, 'ru')
    }

    return 0
}

function compareTaskIdentity(left: ResidentDutyTask, right: ResidentDutyTask): number {
    if (left.cost !== right.cost) {
        return right.cost - left.cost
    }

    const titleDiff = left.title.localeCompare(right.title, 'ru')
    if (titleDiff !== 0) {
        return titleDiff
    }

    return left.id.localeCompare(right.id)
}

function isDoneMineTask(task: ResidentDutyTask): boolean {
    return task.status === 'completed' || task.status === 'verified'
}

function compareTasksForAllDisplay(left: ResidentDutyTask, right: ResidentDutyTask): number {
    const areaDiff = compareTaskArea(left, right)
    if (areaDiff !== 0) {
        return areaDiff
    }

    const freeDiff = Number(left.status !== 'free') - Number(right.status !== 'free')
    if (freeDiff !== 0) {
        return freeDiff
    }

    const priorityDiff = taskStatusPriority(left) - taskStatusPriority(right)
    if (priorityDiff !== 0) {
        return priorityDiff
    }

    return compareTaskIdentity(left, right)
}

function compareTasksForMineDisplay(left: ResidentDutyTask, right: ResidentDutyTask): number {
    const areaDiff = compareTaskArea(left, right)
    if (areaDiff !== 0) {
        return areaDiff
    }

    const doneDiff = Number(isDoneMineTask(left)) - Number(isDoneMineTask(right))
    if (doneDiff !== 0) {
        return doneDiff
    }

    const priorityDiff = taskStatusPriority(left) - taskStatusPriority(right)
    if (priorityDiff !== 0) {
        return priorityDiff
    }

    return compareTaskIdentity(left, right)
}

function compareAreaPriorityForAll(leftTasks: ResidentDutyTask[], rightTasks: ResidentDutyTask[]): number {
    const leftHasFree = leftTasks.some((task) => task.status === 'free')
    const rightHasFree = rightTasks.some((task) => task.status === 'free')
    if (leftHasFree !== rightHasFree) {
        return leftHasFree ? -1 : 1
    }

    return compareTaskArea(leftTasks[0]!, rightTasks[0]!)
}

function compareAreaPriorityForMine(leftTasks: ResidentDutyTask[], rightTasks: ResidentDutyTask[]): number {
    const leftHasOpen = leftTasks.some((task) => !isDoneMineTask(task))
    const rightHasOpen = rightTasks.some((task) => !isDoneMineTask(task))
    if (leftHasOpen !== rightHasOpen) {
        return leftHasOpen ? -1 : 1
    }

    return compareTaskArea(leftTasks[0]!, rightTasks[0]!)
}

function groupTasksByAreaKey(tasks: ResidentDutyTask[]): ResidentDutyTask[][] {
    const groups = new Map<string, ResidentDutyTask[]>()

    tasks.forEach((task) => {
        const key = `${task.area_floor}:${task.area_name}`
        const existing = groups.get(key)

        if (existing) {
            existing.push(task)
            return
        }

        groups.set(key, [task])
    })

    return Array.from(groups.values())
}

function sortTasksByGroupedOrder(
    tasks: ResidentDutyTask[],
    mode: InitialTaskOrderMode,
): ResidentDutyTask[] {
    const groups = groupTasksByAreaKey(tasks)
    const compareArea = mode === 'mine' ? compareAreaPriorityForMine : compareAreaPriorityForAll
    const compareTask = mode === 'mine' ? compareTasksForMineDisplay : compareTasksForAllDisplay

    return groups
        .sort((left, right) => compareArea(left, right))
        .flatMap((groupTasks) => [...groupTasks].sort(compareTask))
}

export function sortTasksForInitialDisplay(
    tasks: ResidentDutyTask[],
    mode: InitialTaskOrderMode = 'all',
): ResidentDutyTask[] {
    return sortTasksByGroupedOrder(tasks, mode)
}

export function buildInitialVisibleMineTaskIds(tasks: ResidentDutyTask[]): string[] {
    return sortTasksByGroupedOrder(
        tasks.filter((task) => task.is_mine),
        'mine',
    ).map((task) => task.id)
}

export function buildInitialVisibleFreeTaskIds(tasks: ResidentDutyTask[]): string[] {
    return sortTasksByGroupedOrder(
        tasks.filter((task) => task.status === 'free'),
        'all',
    ).map((task) => task.id)
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
