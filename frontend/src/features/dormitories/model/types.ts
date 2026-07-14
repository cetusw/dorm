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
