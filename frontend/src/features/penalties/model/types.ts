export type PenaltyResidentSummary = {
    user_id: string
    full_name: string
    total_weight: number
    threshold_reached: boolean
}

export type PenaltyResidentsResponse = {
    residents: PenaltyResidentSummary[]
}

export type PenaltyResidentOption = {
    id: string
    name: string
}

export type PenaltyResidentOptionsResponse = {
    residents: PenaltyResidentOption[]
}

export type PenaltyEntryType = 'issue' | 'resolve'

export type PenaltyEntryItem = {
    id: string
    type: PenaltyEntryType
    reason: string
    weight: number
    created_at: string
}

export type PenaltyResidentDetailsResponse = {
    user_id: string
    full_name: string
    total_weight: number
    entries: PenaltyEntryItem[]
}

export type CurrentUserPenaltiesResponse = {
    total_weight: number
    entries: PenaltyEntryItem[]
}

export type CreatePenaltyRequest = {
    user_id: string
    reason: string
    weight: number
}

export type ResolvePenaltyRequest = {
    user_id: string
    reason: string
    weight: number
}

export type UpdatePenaltyEntryRequest = {
    reason: string
    weight: number
}
