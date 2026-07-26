import { apiRequest } from '../../../shared/api/apiClient'
import type { AreaDetails, CreateAreaRequest, UpdateAreaRequest } from '../../areas/model/types'
import type { CreateTaskRequest, TaskDetails, UpdateTaskRequest } from '../../task-catalog/model/types'
import type { DutySettingsResponse } from '../model/types'

function normalizeDutySettingsResponse(response: DutySettingsResponse): DutySettingsResponse {
    return {
        group: response.group,
        areas: Array.isArray(response.areas)
            ? response.areas.map((area) => ({
                ...area,
                floor: area.floor == null ? null : Number(area.floor),
                tasks: Array.isArray(area.tasks)
                    ? area.tasks.map((task) => ({
                        ...task,
                        last_completed_at: task.last_completed_at ?? null,
                    }))
                    : [],
            }))
            : [],
    }
}

function normalizeAreaDetails(area: Record<string, unknown>): AreaDetails {
    return {
        id: Number(area.id),
        name: String(area.name ?? ''),
        floor: area.floor == null ? null : Number(area.floor),
        group: area.group && typeof area.group === 'object'
            ? {
                id: String((area.group as Record<string, unknown>).id ?? ''),
                name: String((area.group as Record<string, unknown>).name ?? ''),
            }
            : null,
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
        frequency: Number(task.frequency ?? 0),
        area: {
            id: Number(area.id ?? 0),
            name: String(area.name ?? ''),
        },
    }
}

export async function getDutySettings(groupId: string): Promise<DutySettingsResponse> {
    const response = await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings`)
    return normalizeDutySettingsResponse(await response.json())
}

export async function createDutySettingsArea(groupId: string, request: CreateAreaRequest): Promise<AreaDetails> {
    const response = await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/areas`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeAreaDetails(await response.json())
}

export async function updateDutySettingsArea(
    groupId: string,
    areaId: number,
    request: UpdateAreaRequest,
): Promise<AreaDetails> {
    const response = await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/areas/${areaId}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeAreaDetails(await response.json())
}

export async function deleteDutySettingsArea(groupId: string, areaId: number): Promise<void> {
    await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/areas/${areaId}`, {
        method: 'DELETE',
    })
}

export async function createDutySettingsTask(
    groupId: string,
    areaId: number,
    request: CreateTaskRequest,
): Promise<TaskDetails> {
    const response = await apiRequest(
        `/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/areas/${areaId}/tasks`,
        {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(request),
        },
    )

    return normalizeTaskDetails(await response.json())
}

export async function updateDutySettingsTask(
    groupId: string,
    taskId: string,
    request: UpdateTaskRequest,
): Promise<TaskDetails> {
    const response = await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/tasks/${taskId}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeTaskDetails(await response.json())
}

export async function deleteDutySettingsTask(groupId: string, taskId: string): Promise<void> {
    await apiRequest(`/api/v1/groups/${encodeURIComponent(groupId)}/duty-settings/tasks/${taskId}`, {
        method: 'DELETE',
    })
}
