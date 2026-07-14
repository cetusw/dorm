import type { TaskRowActionMode } from '../ui/TaskRowActions'
import type { ResidentCurrentDuty, ResidentDutyTask, DutyTaskSelect } from './types'
import {
    countTasksCompletedByMe,
    countTasksTakenByMe,
    sumCostTakenByMe,
} from './utils'

export type DutyAnalytics = {
    myVerifiedTasksCount: number
    totalTasksCount: number
    totalTakenTasksCount: number
    totalCompletedTasksCount: number
    totalVerifiedTasksCount: number
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
    activeTab: DutyTaskSelect
    duty: ResidentCurrentDuty
    reviewVisibleTaskIds: string[]
    visibleFreeTaskIds: string[]
    visibleMineTaskIds: string[]
}

export function calculateDutyAnalytics(tasks: ResidentDutyTask[]): DutyAnalytics {
    return {
        myVerifiedTasksCount: tasks.filter((task) => task.is_mine && task.status === 'verified').length,
        totalTasksCount: tasks.length,
        totalTakenTasksCount: tasks.filter((task) => task.status !== 'free').length,
        totalCompletedTasksCount: tasks.filter(
            (task) => task.status === 'completed' || task.status === 'verified',
        ).length,
        totalVerifiedTasksCount: tasks.filter((task) => task.status === 'verified').length,
        takenCostSum: sumCostTakenByMe(tasks),
        takenTasksCount: countTasksTakenByMe(tasks),
        completedTasksCount: countTasksCompletedByMe(tasks),
    }
}

export function selectDutyViewOptions(
    duty: ResidentCurrentDuty,
    activeTab: DutyTaskSelect,
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
