export type DormitoryLeader = {
    id: string
    name: string
}

export type DormitoryListItem = {
    id: number
    name: string
    address: string
    leader: DormitoryLeader | null
}

export type DormitoryListResponse = {
    dormitories: DormitoryListItem[]
}

export type DormitoryDetails = {
    id: number
    name: string
    city: string
    streetType: string
    streetName: string
    houseNumber: string
    leader: DormitoryLeader | null
}

export type DormitoryFormValues = {
    name: string
    city: string
    streetType: string
    streetName: string
    houseNumber: string
    leaderId: string | null
}

export type CreateDormitoryRequest = {
    name: string
    city: string
    street_type: string
    street_name: string
    house_number: string
    leader_id: string | null
}

export type UpdateDormitoryRequest = CreateDormitoryRequest

export type UserOption = {
    id: string
    name: string
}

export type UserOptionsResponse = {
    users: UserOption[]
}
