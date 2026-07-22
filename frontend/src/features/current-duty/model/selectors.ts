import type { TaskRowActionMode } from '../ui/TaskRowActions'
import type {
    ResidentCurrentDuty,
    ResidentDutyTask,
    DutyTaskSelect,
    ResidentDutyTeamMember,
} from './types'
import {
    countTasksCompletedByMe,
    countTasksTakenByMe,
    sumCostCompletedByMe,
    sumCostTakenByMe,
    sumCostVerifiedByMe,
} from './utils'

export type DutyAnalytics = {
    myVerifiedTasksCount: number
    totalTasksCount: number
    totalTakenTasksCount: number
    totalCompletedTasksCount: number
    totalVerifiedTasksCount: number
    takenCostSum: number
    completedCostSum: number
    verifiedCostSum: number
    takenTasksCount: number
    completedTasksCount: number
}

export type MemberDutyProgressSection = {
    color: string
    value: number
}

export type MemberDutyProgressModel = {
    title: string
    sections: MemberDutyProgressSection[]
}

export type DutyViewOptions = {
    actionMode: TaskRowActionMode
    emptyMessage: string
    isReadOnly: boolean
    showAnalytics: boolean
    showAssigneeColumn: boolean
    showControls: boolean
}

export type TeamMemberTaskGroup = {
    member: ResidentDutyTeamMember
    progress: MemberDutyProgressModel
    tasks: ResidentDutyTask[]
    tooltipLines: string[]
}

type SelectTasksForActiveSelectParams = {
    activeSelect: DutyTaskSelect
    duty: ResidentCurrentDuty
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
        completedCostSum: sumCostCompletedByMe(tasks),
        verifiedCostSum: sumCostVerifiedByMe(tasks),
        takenTasksCount: countTasksTakenByMe(tasks),
        completedTasksCount: countTasksCompletedByMe(tasks),
    }
}

function clampSectionValue(value: number): number {
    return Math.max(0, value)
}

function buildSections(values: MemberDutyProgressSection[]): MemberDutyProgressSection[] {
    return values.filter((section) => section.value > 0)
}

export function buildMemberDutyProgressModel(
    analytics: DutyAnalytics,
    targetCost: number,
): MemberDutyProgressModel {
    const safeTargetCost = Math.max(targetCost, 0)
    const verifiedTasksCount = analytics.myVerifiedTasksCount
    const hasNoTakenTasks = analytics.takenTasksCount === 0

    if (hasNoTakenTasks) {
        return {
            title: `Взято 0 из ${safeTargetCost} баллов`,
            sections: [
                {
                    color: 'var(--app-color-analytics-empty)',
                    value: Math.max(safeTargetCost, 1),
                },
            ],
        }
    }

    if (analytics.takenCostSum < safeTargetCost) {
        const verifiedValue = Math.min(analytics.verifiedCostSum, safeTargetCost)
        const completedValue = Math.min(analytics.completedCostSum, safeTargetCost)
        const takenValue = Math.min(analytics.takenCostSum, safeTargetCost)

        return {
            title: `Взято ${analytics.takenCostSum} из ${safeTargetCost} баллов`,
            sections: buildSections([
                {
                    color: 'var(--app-color-analytics-verified)',
                    value: clampSectionValue(verifiedValue),
                },
                {
                    color: 'var(--app-color-analytics-completed)',
                    value: clampSectionValue(completedValue - verifiedValue),
                },
                {
                    color: 'var(--app-color-analytics-taken)',
                    value: clampSectionValue(takenValue - completedValue),
                },
                {
                    color: 'var(--app-color-analytics-empty)',
                    value: clampSectionValue(safeTargetCost - takenValue),
                },
            ]),
        }
    }

    if (analytics.completedTasksCount < analytics.takenTasksCount) {
        return {
            title: `Выполнено ${analytics.completedTasksCount} из ${analytics.takenTasksCount} задач`,
            sections: buildSections([
                {
                    color: 'var(--app-color-analytics-verified)',
                    value: clampSectionValue(verifiedTasksCount),
                },
                {
                    color: 'var(--app-color-analytics-completed)',
                    value: clampSectionValue(analytics.completedTasksCount - verifiedTasksCount),
                },
                {
                    color: 'var(--app-color-analytics-taken)',
                    value: clampSectionValue(analytics.takenTasksCount - analytics.completedTasksCount),
                },
            ]),
        }
    }

    if (verifiedTasksCount < analytics.takenTasksCount) {
        return {
            title: `Проверено ${verifiedTasksCount} из ${analytics.takenTasksCount} задач`,
            sections: buildSections([
                {
                    color: 'var(--app-color-analytics-verified)',
                    value: clampSectionValue(verifiedTasksCount),
                },
                {
                    color: 'var(--app-color-analytics-completed)',
                    value: clampSectionValue(analytics.takenTasksCount - verifiedTasksCount),
                },
            ]),
        }
    }

    return {
        title: 'Все задачи выполнены и проверены!',
        sections: [
            {
                color: 'var(--app-color-analytics-verified)',
                value: Math.max(analytics.takenTasksCount, 1),
            },
        ],
    }
}

export function buildReadonlyDutyProgressModel(
    analytics: DutyAnalytics,
): MemberDutyProgressModel {
    const totalTasksCount = Math.max(analytics.totalTasksCount, 0)

    if (analytics.totalTakenTasksCount < totalTasksCount) {
        return {
            title: `Взято ${analytics.totalTakenTasksCount} из ${totalTasksCount} задач`,
            sections: buildSections([
                {
                    color: 'var(--app-color-analytics-verified)',
                    value: clampSectionValue(analytics.totalVerifiedTasksCount),
                },
                {
                    color: 'var(--app-color-analytics-completed)',
                    value: clampSectionValue(
                        analytics.totalCompletedTasksCount - analytics.totalVerifiedTasksCount,
                    ),
                },
                {
                    color: 'var(--app-color-analytics-taken)',
                    value: clampSectionValue(
                        analytics.totalTakenTasksCount - analytics.totalCompletedTasksCount,
                    ),
                },
                {
                    color: 'var(--app-color-analytics-empty)',
                    value: clampSectionValue(totalTasksCount - analytics.totalTakenTasksCount),
                },
            ]),
        }
    }

    if (analytics.totalCompletedTasksCount < totalTasksCount) {
        return {
            title: `Выполнено ${analytics.totalCompletedTasksCount} из ${totalTasksCount} задач`,
            sections: buildSections([
                {
                    color: 'var(--app-color-analytics-verified)',
                    value: clampSectionValue(analytics.totalVerifiedTasksCount),
                },
                {
                    color: 'var(--app-color-analytics-completed)',
                    value: clampSectionValue(
                        analytics.totalCompletedTasksCount - analytics.totalVerifiedTasksCount,
                    ),
                },
                {
                    color: 'var(--app-color-analytics-taken)',
                    value: clampSectionValue(totalTasksCount - analytics.totalCompletedTasksCount),
                },
            ]),
        }
    }

    if (analytics.totalVerifiedTasksCount < totalTasksCount) {
        return {
            title: `Проверено ${analytics.totalVerifiedTasksCount} из ${totalTasksCount} задач`,
            sections: buildSections([
                {
                    color: 'var(--app-color-analytics-verified)',
                    value: clampSectionValue(analytics.totalVerifiedTasksCount),
                },
                {
                    color: 'var(--app-color-analytics-completed)',
                    value: clampSectionValue(totalTasksCount - analytics.totalVerifiedTasksCount),
                },
            ]),
        }
    }

    return {
        title: 'Все задачи проверены!',
        sections: [
            {
                color: 'var(--app-color-analytics-verified)',
                value: Math.max(totalTasksCount, 1),
            },
        ],
    }
}

export function selectDutyViewOptions(
    duty: ResidentCurrentDuty,
    activeSelect: DutyTaskSelect,
    visibleTabs: DutyTaskSelect[] = duty.visible_tabs,
): DutyViewOptions {
    const isReadOnly = duty.read_only

    return {
        actionMode: activeSelect === 'review' ? 'review' : 'default',
        emptyMessage: activeSelect === 'review' ? 'В этой группе нет задач.' : 'В этом разделе нет задач.',
        isReadOnly,
        showAnalytics: !isReadOnly && visibleTabs.length > 0,
        showAssigneeColumn: isReadOnly || activeSelect === 'all' || activeSelect === 'review',
        showControls: duty.show_group_select || visibleTabs.length > 0,
    }
}

export function selectVisibleDutyTabs(duty: ResidentCurrentDuty): DutyTaskSelect[] {
    if (duty.read_only || duty.visible_tabs.length === 0) {
        return duty.visible_tabs
    }

    const hasMineTasks = duty.tasks.some((task) => task.is_mine)
    const hasFreeTasks = duty.tasks.some((task) => !task.assignee_id)
    const filteredTabs = duty.visible_tabs.filter((tab) => tab !== 'mine')
    const visibleTabs: DutyTaskSelect[] = []

    if (hasMineTasks) {
        visibleTabs.push('mine')
    }

    if (hasFreeTasks) {
        visibleTabs.push('free')
    }

    for (const tab of filteredTabs) {
        if (tab === 'free') {
            continue
        }

        if (!visibleTabs.includes(tab)) {
            visibleTabs.push(tab)
        }
    }

    return visibleTabs
}

export function selectTasksForActiveSelect({
    activeSelect,
    duty,
    visibleFreeTaskIds,
    visibleMineTaskIds,
}: SelectTasksForActiveSelectParams): ResidentDutyTask[] {
    if (duty.read_only || duty.visible_tabs.length === 0) {
        return duty.tasks
    }

    switch (activeSelect) {
        case 'mine':
            return duty.tasks.filter((task) => visibleMineTaskIds.includes(task.id))
        case 'free':
            return duty.tasks.filter((task) => visibleFreeTaskIds.includes(task.id))
        case 'team':
            return duty.tasks
        case 'review':
            return duty.tasks
        case 'all':
        default:
            return duty.tasks
    }
}

function calculateAnalyticsForMember(tasks: ResidentDutyTask[], memberId: string): DutyAnalytics {
    const memberTasks = tasks.filter((task) => task.assignee_id === memberId)
    const completedTasks = memberTasks.filter(
        (task) => task.status === 'completed' || task.status === 'verified',
    )
    const verifiedTasks = memberTasks.filter((task) => task.status === 'verified')

    return {
        myVerifiedTasksCount: verifiedTasks.length,
        totalTasksCount: 0,
        totalTakenTasksCount: 0,
        totalCompletedTasksCount: 0,
        totalVerifiedTasksCount: 0,
        takenCostSum: memberTasks.reduce((sum, task) => sum + task.cost, 0),
        completedCostSum: completedTasks.reduce((sum, task) => sum + task.cost, 0),
        verifiedCostSum: verifiedTasks.reduce((sum, task) => sum + task.cost, 0),
        takenTasksCount: memberTasks.length,
        completedTasksCount: completedTasks.length,
    }
}

export function selectTeamMemberTaskGroups(
    duty: ResidentCurrentDuty,
): TeamMemberTaskGroup[] {
    return duty.team_members.map((member) => {
        const tasks = duty.tasks.filter((task) => task.assignee_id === member.id)
        const analytics = calculateAnalyticsForMember(duty.tasks, member.id)

        return {
            member,
            progress: buildMemberDutyProgressModel(analytics, duty.cost_per_resident_goal),
            tasks,
            tooltipLines: [
                `Взято ${analytics.takenCostSum} из ${duty.cost_per_resident_goal} баллов`,
                `Выполнено ${analytics.completedTasksCount} из ${analytics.takenTasksCount} задач`,
                `Проверено ${analytics.myVerifiedTasksCount} из ${analytics.takenTasksCount} задач`,
            ],
        }
    })
}
