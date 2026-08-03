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

export type PenaltyItem = {
    id: string
    reason: string
    weight: number
    issued_on: string
}

export type PenaltyResidentDetailsResponse = {
    user_id: string
    full_name: string
    penalties: PenaltyItem[]
}

export type CreatePenaltyRequest = {
    user_id: string
    reason: string
    weight: number
    issued_on: string
}
