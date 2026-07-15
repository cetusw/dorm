import { useMemo, useState } from 'react'

import {
    Button,
    Group,
    Menu,
    Paper,
    ScrollArea,
    Table,
    Text,
    TextInput,
    UnstyledButton,
} from '@mantine/core'

import type { ResidentListItem } from '../model/types'

type SortField = 'name' | 'room_number'
type SortDirection = 'asc' | 'desc'

type Props = {
    residents: ResidentListItem[]
    onEdit: (residentId: string) => void
    onDelete: (resident: ResidentListItem) => void
}

function normalize(value: string): string {
    return value.trim().toLowerCase()
}

export function ResidentsTable({ residents, onEdit, onDelete }: Props) {
    const [search, setSearch] = useState('')
    const [sortField, setSortField] = useState<SortField>('name')
    const [sortDirection, setSortDirection] = useState<SortDirection>('asc')
    const [openedMenuId, setOpenedMenuId] = useState<string | null>(null)

    const filteredResidents = useMemo(() => {
        const query = normalize(search)
        const visibleResidents = query === ''
            ? residents
            : residents.filter((resident) =>
                normalize(`${resident.name} ${resident.room_number}`).includes(query),
            )

        return [...visibleResidents].sort((left, right) => {
            const leftValue = normalize(left[sortField])
            const rightValue = normalize(right[sortField])
            const result = leftValue.localeCompare(rightValue, 'ru')

            return sortDirection === 'asc' ? result : result * -1
        })
    }, [residents, search, sortField, sortDirection])

    function toggleSort(nextField: SortField) {
        if (sortField === nextField) {
            setSortDirection((current) => (current === 'asc' ? 'desc' : 'asc'))
            return
        }

        setSortField(nextField)
        setSortDirection('asc')
    }

    return (
        <Paper withBorder radius="lg" bg="white" p="md">
            <TextInput
                value={search}
                onChange={(event) => setSearch(event.currentTarget.value)}
                placeholder="Поиск"
                mb="md"
            />

            <ScrollArea>
                <Table horizontalSpacing="lg" verticalSpacing="md" highlightOnHover miw={720}>
                    <Table.Thead>
                        <Table.Tr>
                            <Table.Th w="42%">
                                <SortButton
                                    label="Имя"
                                    active={sortField === 'name'}
                                    direction={sortDirection}
                                    onClick={() => toggleSort('name')}
                                />
                            </Table.Th>
                            <Table.Th w="34%">
                                <SortButton
                                    label="Комната"
                                    active={sortField === 'room_number'}
                                    direction={sortDirection}
                                    onClick={() => toggleSort('room_number')}
                                />
                            </Table.Th>
                            <Table.Th w="24%">
                                <Text size="sm" fw={700}>
                                    Действия
                                </Text>
                            </Table.Th>
                        </Table.Tr>
                    </Table.Thead>
                    <Table.Tbody>
                        {filteredResidents.map((resident) => {
                            const menuOpened = openedMenuId === resident.id

                            return (
                                <Table.Tr key={resident.id}>
                                    <Table.Td>
                                        <Text fw={600}>{resident.name}</Text>
                                    </Table.Td>
                                    <Table.Td>
                                        <Text>{resident.room_number}</Text>
                                    </Table.Td>
                                    <Table.Td>
                                        <Group justify="flex-start">
                                            <Menu
                                                opened={menuOpened}
                                                onChange={(opened) =>
                                                    setOpenedMenuId(opened ? resident.id : null)
                                                }
                                                withinPortal
                                                position="bottom-end"
                                            >
                                                <Menu.Target>
                                                    <Button variant="subtle" px="sm" aria-label="Открыть действия">
                                                        {menuOpened ? '▴' : '▾'}
                                                    </Button>
                                                </Menu.Target>

                                                <Menu.Dropdown>
                                                    <Menu.Item onClick={() => onEdit(resident.id)}>
                                                        Редактировать
                                                    </Menu.Item>
                                                    <Menu.Item color="red" onClick={() => onDelete(resident)}>
                                                        Удалить
                                                    </Menu.Item>
                                                </Menu.Dropdown>
                                            </Menu>
                                        </Group>
                                    </Table.Td>
                                </Table.Tr>
                            )
                        })}
                    </Table.Tbody>
                </Table>
            </ScrollArea>
        </Paper>
    )
}

type SortButtonProps = {
    label: string
    active: boolean
    direction: SortDirection
    onClick: () => void
}

function SortButton({ label, active, direction, onClick }: SortButtonProps) {
    return (
        <UnstyledButton onClick={onClick}>
            <Group gap={6} wrap="nowrap">
                <Text size="sm" fw={700}>
                    {label}
                </Text>
                <Text size="xs" c={active ? 'dark' : 'dimmed'}>
                    {active ? (direction === 'asc' ? '▲' : '▼') : '↕'}
                </Text>
            </Group>
        </UnstyledButton>
    )
}
