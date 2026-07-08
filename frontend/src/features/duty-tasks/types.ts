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
}

export type ResidentCurrentDuty = {
    duty_id: string
    group: string
    team: string
    start_date: string
    end_date: string
    tasks: ResidentDutyTask[]
}