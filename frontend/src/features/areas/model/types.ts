export type AreaGroup = {
    id: string
    name: string
}

export type AreaListItem = {
    id: number
    name: string
    floor: number | null
    group: AreaGroup | null
}

export type AreaListResponse = {
    areas: AreaListItem[]
}

export type AreaDetails = {
    id: number
    name: string
    floor: number | null
    group: AreaGroup | null
}

export type AreaFormValues = {
    name: string
    groupId: string | null
    floor: string
}

export type CreateAreaRequest = {
    name: string
    group_id: string | null
    floor: number | null
}

export type UpdateAreaRequest = CreateAreaRequest
