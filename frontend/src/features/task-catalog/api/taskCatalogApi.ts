import { apiRequest } from '../../../shared/api/apiClient'
import type {
    CreateTaskRequest,
    TaskDetails,
    TaskListResponse,
    UpdateTaskRequest,
} from '../model/types'

function normalizeTaskListResponse(response: TaskListResponse): TaskListResponse {
    return {
        tasks: Array.isArray(response.tasks) ? response.tasks : [],
    }
}

function normalizeTaskDetails(task: Record<string, unknown>): TaskDetails {
    const area = task.area && typeof task.area === 'object'
        ? (task.area as Record<string, unknown>)
        : {}

    return {
        id: String(task.id ?? ''),
        title: String(task.title ?? ''),
        cost: Number(task.cost ?? 0),
        recurrenceInterval: Number(task.recurrenceInterval ?? 0),
        startSequence: Number(task.startSequence ?? 1),
        area: {
            id: Number(area.id ?? 0),
            name: String(area.name ?? ''),
        },
    }
}

export async function getTaskDefinitions(dormitoryId: string): Promise<TaskListResponse> {
    const response = await apiRequest(`/api/v1/task-definitions?dormitory_id=${encodeURIComponent(dormitoryId)}`)
    return normalizeTaskListResponse(await response.json())
}

export async function getTaskDefinition(dormitoryId: string, taskId: string): Promise<TaskDetails> {
    const response = await apiRequest(`/api/v1/task-definitions/${taskId}?dormitory_id=${encodeURIComponent(dormitoryId)}`)
    return normalizeTaskDetails(await response.json())
}

export async function createTaskDefinition(
    dormitoryId: string,
    request: CreateTaskRequest,
): Promise<TaskDetails> {
    const response = await apiRequest(`/api/v1/task-definitions?dormitory_id=${encodeURIComponent(dormitoryId)}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeTaskDetails(await response.json())
}

export async function updateTaskDefinition(
    dormitoryId: string,
    taskId: string,
    request: UpdateTaskRequest,
): Promise<TaskDetails> {
    const response = await apiRequest(
        `/api/v1/task-definitions/${taskId}?dormitory_id=${encodeURIComponent(dormitoryId)}`,
        {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(request),
        },
    )

    return normalizeTaskDetails(await response.json())
}

export async function deleteTaskDefinition(dormitoryId: string, taskId: string): Promise<void> {
    await apiRequest(`/api/v1/task-definitions/${taskId}?dormitory_id=${encodeURIComponent(dormitoryId)}`, {
        method: 'DELETE',
    })
}
