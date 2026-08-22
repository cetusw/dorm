export type IndividualTaskStatus = 'issued' | 'completed' | 'verified'

export type IndividualTaskArea = { id: number; name: string; floor: number | null }
export type IndividualTaskResident = { id: string; name: string }
export type IndividualTask = {
    id: string; dormitory_id: number; resident: IndividualTaskResident; title: string
    area: IndividualTaskArea | null; redemption_weight: number; deadline: string | null
    status: IndividualTaskStatus; is_overdue: boolean; completed_at: string | null
    verified_at: string | null; created_at: string; updated_at: string; version: number
    can_edit: boolean; can_delete: boolean; can_complete: boolean; can_open: boolean; can_verify: boolean; can_reject: boolean
}
export type IndividualTaskResidentOption = IndividualTaskResident & {
    dormitory_id: number; penalty_balance: number; reserved_redemption_weight: number; available_redemption_weight: number
}
export type IndividualTaskRequest = { resident_id: string; title: string; area_id: number | null; redemption_weight: number; deadline: string | null; version: number }
