export type DutyTaskStatus = 'free' | 'assigned' | 'completed' | 'verified'

export type ResidentDutyTask = {
    id: string
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
}

export type ResidentDutyGroupOption = {
    id: string
    name: string
}

export type ResidentCurrentDuty = {
    selected_group_id: string
    has_active_duty: boolean
    can_manage_tasks: boolean
    groups: ResidentDutyGroupOption[]
    duty_id: string
    group: string
    team: string
    start_date: string
    end_date: string
    cost_per_resident_goal: number
    my_taken_cost_sum: number
    tasks: ResidentDutyTask[]
}
