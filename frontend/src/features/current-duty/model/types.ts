import type { DutyTaskStatus } from '../../../entities/duty-task'

export type DutyTaskSelect = 'mine' | 'free' | 'all' | 'verification' | 'team'

export type ResidentDutyTask = {
    id: string
    area_id: number
    area_name: string
    area_floor: number
    title: string
    cost: number
    status: DutyTaskStatus
    assignee_id?: string
    assignee_name?: string
    is_mine: boolean
    can_take: boolean
    can_return: boolean
    can_complete: boolean
    can_open: boolean
    can_verify: boolean
    can_review_open: boolean
    needs_revision?: boolean
}

export type ResidentDutyGroupOption = {
    id: string
    name: string
}

export type ResidentDutyTeamMember = {
    id: string
    name: string
}

export type CurrentDutyNotification = {
    id: number
    color: string
    title: string
    message: string
}

export type ResidentCurrentDuty = {
    dormitory_id: number
    selected_group_id: string
    has_active_duty: boolean
    can_manage_tasks: boolean
    can_manage_duty_settings: boolean
    read_only: boolean
    show_group_select: boolean
    visible_tabs: DutyTaskSelect[]
    notice_message: string
    notice_tone?: 'info' | 'warning' | ''
    period_status: 'past' | 'active' | 'future' | ''
    groups: ResidentDutyGroupOption[]
    duty_id: string
    group: string
    team: string
    start_date: string
    end_date: string
    cost_per_resident_goal: number
    my_taken_cost_sum: number
    team_members: ResidentDutyTeamMember[]
    tasks: ResidentDutyTask[]
}

export type ResidentDutyDetails = ResidentCurrentDuty

export type DutyPeriodStatus = 'past' | 'active' | 'future'

export type DutyHistoryProgress = {
    total_cost_sum: number
    taken_cost_sum: number
    total_tasks_count: number
    taken_tasks_count: number
    completed_tasks_count: number
    verified_tasks_count: number
}

export type DutyHistoryItem = {
    id: string
    start_date: string
    end_date: string
    team_leader_name: string
    progress: DutyHistoryProgress
}

export type DutyHistoryResponse = {
    selected_group_id: string
    show_group_select: boolean
    groups: ResidentDutyGroupOption[]
    duties: DutyHistoryItem[]
}
