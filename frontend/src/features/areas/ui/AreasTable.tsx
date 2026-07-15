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

import type { AreaListItem } from '../model/types'

type SortField = 'name' | 'floor' | 'groupName'
type SortDirection = 'asc' | 'desc'

type Props = {
    areas: AreaListItem[]
    onEdit: (areaId: number) => void
    onDelete: (area: AreaListItem) => void
}

function normalize(value: string): string {
    return value.trim().toLowerCase()
}

export function AreasTable({ areas, onEdit, onDelete }: Props) {
    const [search, setSearch] = useState('')
    const [sortField, setSortField] = useState<SortField>('name')
    const [sortDirection, setSortDirection] = useState<SortDirection>('asc')
    const [openedMenuId, setOpenedMenuId] = useState<number | null>(null)

    const filteredAreas = useMemo(() => {
        const query = normalize(search)
        const visibleAreas = query === ''
            ? areas
            : areas.filter((area) =>
                normalize(
                    `${area.name} ${area.floor ?? ''} ${area.group?.name ?? 'Общая территория'}`,
                ).includes(query),
            )

        return [...visibleAreas].sort((left, right) => {
            if (sortField === 'floor') {
                const leftFloor = left.floor ?? Number.NEGATIVE_INFINITY
                const rightFloor = right.floor ?? Number.NEGATIVE_INFINITY
                const result = leftFloor - rightFloor
                return sortDirection === 'asc' ? result : result * -1
            }

            const leftValue = normalize(
                sortField === 'groupName' ? left.group?.name ?? 'Общая территория' : left.name,
            )
            const rightValue = normalize(
                sortField === 'groupName' ? right.group?.name ?? 'Общая территория' : right.name,
            )
            const result = leftValue.localeCompare(rightValue, 'ru')

            return sortDirection === 'asc' ? result : result * -1
        })
    }, [areas, search, sortField, sortDirection])

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
                <Table horizontalSpacing="lg" verticalSpacing="md" highlightOnHover miw={840}>
                    <Table.Thead>
                        <Table.Tr>
                            <Table.Th w="34%">
                                <SortButton
                                    label="Название"
                                    active={sortField === 'name'}
                                    direction={sortDirection}
                                    onClick={() => toggleSort('name')}
                                />
                            </Table.Th>
                            <Table.Th w="18%">
                                <SortButton
                                    label="Этаж"
                                    active={sortField === 'floor'}
                                    direction={sortDirection}
                                    onClick={() => toggleSort('floor')}
                                />
                            </Table.Th>
                            <Table.Th w="28%">
                                <SortButton
                                    label="Группа"
                                    active={sortField === 'groupName'}
                                    direction={sortDirection}
                                    onClick={() => toggleSort('groupName')}
                                />
                            </Table.Th>
                            <Table.Th w="20%">
                                <Text size="sm" fw={700}>
                                    Действия
                                </Text>
                            </Table.Th>
                        </Table.Tr>
                    </Table.Thead>
                    <Table.Tbody>
                        {filteredAreas.map((area) => {
                            const menuOpened = openedMenuId === area.id

                            return (
                                <Table.Tr key={area.id}>
                                    <Table.Td>
                                        <Text fw={600}>{area.name}</Text>
                                    </Table.Td>
                                    <Table.Td>
                                        <Text c={area.floor == null ? 'dimmed' : undefined}>
                                            {area.floor ?? 'Не указан'}
                                        </Text>
                                    </Table.Td>
                                    <Table.Td>
                                        <Text c={area.group ? undefined : 'dimmed'}>
                                            {area.group?.name ?? 'Общая территория'}
                                        </Text>
                                    </Table.Td>
                                    <Table.Td>
                                        <Group justify="flex-start">
                                            <Menu
                                                opened={menuOpened}
                                                onChange={(opened) => setOpenedMenuId(opened ? area.id : null)}
                                                withinPortal
                                                position="bottom-end"
                                            >
                                                <Menu.Target>
                                                    <Button variant="subtle" px="sm" aria-label="Открыть действия">
                                                        {menuOpened ? '▴' : '▾'}
                                                    </Button>
                                                </Menu.Target>
                                                <Menu.Dropdown>
                                                    <Menu.Item onClick={() => onEdit(area.id)}>
                                                        Редактировать
                                                    </Menu.Item>
                                                    <Menu.Item color="red" onClick={() => onDelete(area)}>
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
