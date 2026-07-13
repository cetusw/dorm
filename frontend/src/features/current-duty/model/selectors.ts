import type { TaskRowActionMode } from '../ui/TaskRowActions'
import type { ResidentCurrentDuty, ResidentDutyTask, DutyTaskTab } from './types'
import {
    countTasksCompletedByMe,
    countTasksTakenByMe,
    sumCostTakenByMe,
} from './utils'

export type DutyAnalytics = {
    takenCostSum: number
    takenTasksCount: number
    completedTasksCount: number
}

export type DutyViewOptions = {
    actionMode: TaskRowActionMode
    emptyMessage: string
    isReadOnly: boolean
    showAnalytics: boolean
    showAssigneeColumn: boolean
    showControls: boolean
}

type SelectTasksForTabParams = {
    activeTab: DutyTaskTab
    duty: ResidentCurrentDuty
    reviewVisibleTaskIds: string[]
    visibleFreeTaskIds: string[]
    visibleMineTaskIds: string[]
}

export function calculateDutyAnalytics(tasks: ResidentDutyTask[]): DutyAnalytics {
    return {
        takenCostSum: sumCostTakenByMe(tasks),
        takenTasksCount: countTasksTakenByMe(tasks),
        completedTasksCount: countTasksCompletedByMe(tasks),
    }
}

export function selectDutyViewOptions(
    duty: ResidentCurrentDuty,
    activeTab: DutyTaskTab,
): DutyViewOptions {
    const isReadOnly = duty.read_only

    return {
        actionMode: activeTab === 'review' ? 'review' : 'default',
        emptyMessage: activeTab === 'review' ? 'Нет задач на проверке' : 'В этом разделе нет задач.',
        isReadOnly,
        showAnalytics: !isReadOnly && duty.visible_tabs.length > 0,
        showAssigneeColumn: isReadOnly || activeTab === 'all' || activeTab === 'review',
        showControls: duty.show_group_select || duty.visible_tabs.length > 0,
    }
}

export function selectReviewVisibleTaskIds(tasks: ResidentDutyTask[]): string[] {
    return tasks
        .filter((task) => task.status === 'completed')
        .map((task) => task.id)
}

export function selectTasksForTab({
    activeTab,
    duty,
    reviewVisibleTaskIds,
    visibleFreeTaskIds,
    visibleMineTaskIds,
}: SelectTasksForTabParams): ResidentDutyTask[] {
    if (duty.read_only || duty.visible_tabs.length === 0) {
        return duty.tasks
    }

    switch (activeTab) {
        case 'mine':
            return duty.tasks.filter((task) => visibleMineTaskIds.includes(task.id))
        case 'free':
            return duty.tasks.filter((task) => visibleFreeTaskIds.includes(task.id))
        case 'review':
            return duty.tasks.filter((task) => reviewVisibleTaskIds.includes(task.id))
        case 'all':
        default:
            return duty.tasks
    }
}
