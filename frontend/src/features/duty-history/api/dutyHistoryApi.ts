import { apiRequest } from '../../../shared/api/apiClient'
import type { DutyHistoryResponse, ResidentDutyDetails } from '../../current-duty/model/types'

function normalizeDutyDetails(duty: ResidentDutyDetails): ResidentDutyDetails {
    return {
        ...duty,
        visible_tabs: Array.isArray(duty.visible_tabs) ? duty.visible_tabs : [],
        groups: Array.isArray(duty.groups) ? duty.groups : [],
        team_members: Array.isArray(duty.team_members) ? duty.team_members : [],
        tasks: Array.isArray(duty.tasks) ? duty.tasks : [],
    }
}

export async function getDutyHistory(groupId?: string): Promise<DutyHistoryResponse> {
    const params = new URLSearchParams()
    if (groupId) {
        params.set('group_id', groupId)
    }

    const suffix = params.toString()
    const response = await apiRequest(
        suffix ? `/api/v1/resident/duties?${suffix}` : '/api/v1/resident/duties',
    )
    const payload = await response.json() as DutyHistoryResponse

    return {
        selected_group_id: payload.selected_group_id,
        show_group_select: Boolean(payload.show_group_select),
        groups: Array.isArray(payload.groups) ? payload.groups : [],
        duties: Array.isArray(payload.duties)
            ? payload.duties.map((item) => ({
                ...item,
                progress: item?.progress ?? {
                    total_cost_sum: 0,
                    taken_cost_sum: 0,
                    total_tasks_count: 0,
                    taken_tasks_count: 0,
                    completed_tasks_count: 0,
                    verified_tasks_count: 0,
                },
            }))
            : [],
    }
}

export async function getDutyDetails(dutyId: string): Promise<ResidentDutyDetails> {
    const response = await apiRequest(`/api/v1/resident/duties/${encodeURIComponent(dutyId)}`)
    return normalizeDutyDetails(await response.json())
}
