export type TaskArea = {
    id: number
    name: string
}

export type TaskListItem = {
    id: string
    title: string
    cost: number
    frequency: number
    area: TaskArea
}

export type TaskListResponse = {
    tasks: TaskListItem[]
}

export type TaskDetails = {
    id: string
    title: string
    cost: number
    frequency: number
    area: TaskArea
}

export type TaskFormValues = {
    title: string
    cost: string
    frequency: string
    areaId: string | null
}

export type CreateTaskRequest = {
    title: string
    cost: number
    frequency: number
    area_id: number
    one_time?: boolean
    include_in_current_duty?: boolean
}

export type UpdateTaskRequest = CreateTaskRequest
