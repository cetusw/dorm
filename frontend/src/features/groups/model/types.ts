export type GroupLeader = {
    id: string
    name: string
}

export type GroupListItem = {
    id: string
    name: string
    leader: GroupLeader | null
}

export type GroupListResponse = {
    groups: GroupListItem[]
}

export type GroupDetails = {
    id: string
    name: string
    leader: GroupLeader | null
}

export type GroupFormValues = {
    name: string
    leaderId: string | null
}

export type CreateGroupRequest = {
    name: string
    leader_id: string | null
    dormitory_id: number
}

export type UpdateGroupRequest = CreateGroupRequest

export type GroupUserOption = {
    id: string
    name: string
}

export type GroupUserOptionsResponse = {
    users: GroupUserOption[]
}
