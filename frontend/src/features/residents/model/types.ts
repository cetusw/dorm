export type ResidentListItem = {
    id: string
    name: string
    room_number: string
}

export type ResidentListResponse = {
    users: ResidentListItem[]
}
