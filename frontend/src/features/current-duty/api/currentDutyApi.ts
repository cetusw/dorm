import { apiRequest } from '../../../shared/api/apiClient'
import type { ResidentCurrentDuty } from '../model/types'

function normalizeCurrentDuty(duty: ResidentCurrentDuty): ResidentCurrentDuty {
    return {
        ...duty,
        visible_tabs: Array.isArray(duty.visible_tabs) ? duty.visible_tabs : [],
        groups: Array.isArray(duty.groups) ? duty.groups : [],
        team_members: Array.isArray(duty.team_members) ? duty.team_members : [],
        tasks: Array.isArray(duty.tasks) ? duty.tasks : [],
    }
}

function buildCurrentDutyPath(groupId?: string): string {
    if (!groupId) {
        return '/api/v1/resident/current-duty'
    }

    return `/api/v1/resident/current-duty?group_id=${encodeURIComponent(groupId)}`
}

export async function getCurrentDuty(groupId?: string): Promise<ResidentCurrentDuty> {
    const response = await apiRequest(buildCurrentDutyPath(groupId))
    return normalizeCurrentDuty(await response.json())
}

async function requestDutyAction(
    taskId: string,
    action: string,
    groupId?: string,
): Promise<ResidentCurrentDuty> {
    const path = groupId
        ? `/api/v1/resident/tasks/${taskId}/${action}?group_id=${encodeURIComponent(groupId)}`
        : `/api/v1/resident/tasks/${taskId}/${action}`

    const response = await apiRequest(path, {
        method: 'POST',
    })

    return normalizeCurrentDuty(await response.json())
}

export async function takeTask(taskId: string, groupId?: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'take', groupId)
}

export async function returnTask(taskId: string, groupId?: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'return', groupId)
}

export async function completeTask(taskId: string, groupId?: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'complete', groupId)
}

export async function openTask(taskId: string, groupId?: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'open', groupId)
}

export async function reopenTask(taskId: string, groupId?: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'reopen', groupId)
}

export async function verifyTask(taskId: string, groupId?: string): Promise<ResidentCurrentDuty> {
    return requestDutyAction(taskId, 'verify', groupId)
}

export async function createDormitoryDutyWeek(
    dormitoryId: string,
    startDate: string,
    endDate: string,
): Promise<void> {
    await apiRequest('/api/v1/resident/duties', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            dormitory_id: Number(dormitoryId),
            start_date: startDate,
            end_date: endDate,
        }),
    })
}
