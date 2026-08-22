import { apiRequest } from '../../../shared/api/apiClient'
import type { IndividualTask, IndividualTaskArea, IndividualTaskRequest, IndividualTaskResidentOption } from '../model/types'

type Payload = Record<string, unknown>

function asPayload(value: unknown): Payload { return value as Payload }
function nullableString(value: unknown): string | null { return typeof value === 'string' ? value : null }
function normalizeArea(value: unknown): IndividualTaskArea | null {
    if (!value || typeof value !== 'object') return null
    const area = asPayload(value)
    return { id: Number(area.id ?? 0), name: String(area.name ?? ''), floor: typeof area.floor === 'number' ? area.floor : null }
}
function normalizeTask(value: Payload): IndividualTask {
    const resident = asPayload(value.resident ?? {})
    return {
        id: String(value.id ?? ''), dormitory_id: Number(value.dormitory_id ?? 0), title: String(value.title ?? ''), area: normalizeArea(value.area), redemption_weight: Number(value.redemption_weight ?? 0), deadline: nullableString(value.deadline), status: value.status === 'completed' || value.status === 'verified' ? value.status : 'issued', is_overdue: Boolean(value.is_overdue), completed_at: nullableString(value.completed_at), verified_at: nullableString(value.verified_at), created_at: String(value.created_at ?? ''), updated_at: String(value.updated_at ?? ''), version: Number(value.version ?? 0), can_edit: Boolean(value.can_edit), can_delete: Boolean(value.can_delete), can_complete: Boolean(value.can_complete), can_open: Boolean(value.can_open), can_verify: Boolean(value.can_verify), can_reject: Boolean(value.can_reject), resident: { id: String(resident.id ?? ''), name: String(resident.name ?? '') },
    }
}
async function getPayload(path: string): Promise<Payload> { return (await apiRequest(path)).json() as Promise<Payload> }
async function mutate(path: string, method: string, request?: IndividualTaskRequest): Promise<IndividualTask> {
    const response = await apiRequest(path, { method, headers: request ? { 'Content-Type': 'application/json' } : undefined, body: request ? JSON.stringify(request) : undefined })
    return normalizeTask(await response.json() as Payload)
}

export async function searchIndividualTaskResidents(query: string): Promise<IndividualTaskResidentOption[]> {
    const params = new URLSearchParams(); if (query.trim()) params.set('q', query.trim())
    const payload = await getPayload(`/api/v1/individual-tasks/residents${params.size ? `?${params}` : ''}`)
    return Array.isArray(payload.residents) ? payload.residents.map((value) => { const item = asPayload(value); return { id: String(item.id ?? ''), name: String(item.name ?? ''), dormitory_id: Number(item.dormitory_id ?? 0), penalty_balance: Number(item.penalty_balance ?? 0), reserved_redemption_weight: Number(item.reserved_redemption_weight ?? 0), available_redemption_weight: Number(item.available_redemption_weight ?? 0) } }) : []
}
export async function getIndividualTaskAreas(dormitoryID: number): Promise<IndividualTaskArea[]> { const payload = await getPayload(`/api/v1/individual-tasks/areas?dormitory_id=${dormitoryID}`); return Array.isArray(payload.areas) ? payload.areas.map(normalizeArea).filter((area): area is IndividualTaskArea => area !== null) : [] }
export async function getResidentIndividualTasks(residentID: string): Promise<IndividualTask[]> { const payload = await getPayload(`/api/v1/individual-tasks/residents/${encodeURIComponent(residentID)}`); return Array.isArray(payload.tasks) ? payload.tasks.map((item) => normalizeTask(asPayload(item))) : [] }
export async function getMyIndividualTasks(): Promise<IndividualTask[]> { const payload = await getPayload('/api/v1/individual-tasks/me'); return Array.isArray(payload.tasks) ? payload.tasks.map((item) => normalizeTask(asPayload(item))) : [] }
export const createIndividualTask = (request: IndividualTaskRequest) => mutate('/api/v1/individual-tasks', 'POST', request)
export const updateIndividualTask = (id: string, request: IndividualTaskRequest) => mutate(`/api/v1/individual-tasks/${encodeURIComponent(id)}`, 'PUT', request)
export const verifyIndividualTask = (id: string) => mutate(`/api/v1/individual-tasks/${encodeURIComponent(id)}/verify`, 'POST')
export const rejectIndividualTask = (id: string) => mutate(`/api/v1/individual-tasks/${encodeURIComponent(id)}/reject`, 'POST')
export const completeIndividualTask = (id: string) => mutate(`/api/v1/individual-tasks/${encodeURIComponent(id)}/complete`, 'POST')
export const openIndividualTask = (id: string) => mutate(`/api/v1/individual-tasks/${encodeURIComponent(id)}/open`, 'POST')
export async function deleteIndividualTask(id: string) { await apiRequest(`/api/v1/individual-tasks/${encodeURIComponent(id)}`, { method: 'DELETE' }) }
