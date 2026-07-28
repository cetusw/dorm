export type TaskArea = {
    id: number
    name: string
}

export type TaskListItem = {
    id: string
    title: string
    cost: number
    recurrenceInterval: number
    startSequence: number
    area: TaskArea
}

export type TaskListResponse = {
    tasks: TaskListItem[]
}

export type TaskDetails = {
    id: string
    title: string
    cost: number
    recurrenceInterval: number
    startSequence: number
    area: TaskArea
}

export type TaskFormValues = {
    title: string
    cost: string
    recurrenceInterval: string | null
    areaId: string | null
}

export type CreateTaskRequest = {
    title: string
    cost: number
    recurrenceInterval: number
    area_id: number
    include_in_current_duty?: boolean
}

export type UpdateTaskRequest = CreateTaskRequest
