import { apiRequest } from '../../../shared/api/apiClient'
import type {
    CreateGroupRequest,
    GroupDetails,
    GroupListResponse,
    GroupUserOptionsResponse,
    UpdateGroupRequest,
} from '../model/types'

function normalizeGroupListResponse(response: GroupListResponse): GroupListResponse {
    return {
        groups: Array.isArray(response.groups) ? response.groups : [],
    }
}

function normalizeGroupDetails(group: Record<string, unknown>): GroupDetails {
    return {
        id: String(group.id ?? ''),
        name: String(group.name ?? ''),
        leader: group.leader && typeof group.leader === 'object'
            ? {
                id: String((group.leader as Record<string, unknown>).id ?? ''),
                name: String((group.leader as Record<string, unknown>).name ?? ''),
            }
            : null,
    }
}

function normalizeGroupUserOptionsResponse(response: GroupUserOptionsResponse): GroupUserOptionsResponse {
    return {
        users: Array.isArray(response.users) ? response.users : [],
    }
}

export async function getGroups(dormitoryId: string): Promise<GroupListResponse> {
    const response = await apiRequest(`/api/v1/groups?dormitory_id=${encodeURIComponent(dormitoryId)}`)
    return normalizeGroupListResponse(await response.json())
}

export async function getGroup(groupId: string): Promise<GroupDetails> {
    const response = await apiRequest(`/api/v1/groups/${groupId}`)
    return normalizeGroupDetails(await response.json())
}

export async function getGroupUserOptions(dormitoryId: string): Promise<GroupUserOptionsResponse> {
    const response = await apiRequest(`/api/v1/groups/options?dormitory_id=${encodeURIComponent(dormitoryId)}`)
    return normalizeGroupUserOptionsResponse(await response.json())
}

export async function createGroup(request: CreateGroupRequest): Promise<GroupDetails> {
    const response = await apiRequest('/api/v1/groups', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeGroupDetails(await response.json())
}

export async function updateGroup(groupId: string, request: UpdateGroupRequest): Promise<GroupDetails> {
    const response = await apiRequest(`/api/v1/groups/${groupId}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })

    return normalizeGroupDetails(await response.json())
}

export async function deleteGroup(groupId: string): Promise<void> {
    await apiRequest(`/api/v1/groups/${groupId}`, {
        method: 'DELETE',
    })
}
