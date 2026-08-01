export type ResidentListItem = {
    id: string
    name: string
    room_number: string
}

export type ResidentListResponse = {
    users: ResidentListItem[]
}

export type ResidentDetails = {
    id: string
    first_name: string
    last_name: string
    middle_name: string | null
    login: string
    floor: number | null
    room_number: string | null
}

export type ResidentFormValues = {
    lastName: string
    firstName: string
    middleName: string
    login: string
    password: string
    floor: string
    roomNumber: string
}

export type CreateResidentRequest = {
    first_name: string
    last_name: string
    middle_name: string | null
    login: string
    password: string
    dormitory_id: number
    floor: number | null
    room_number: string | null
}

export type UpdateResidentRequest = CreateResidentRequest
