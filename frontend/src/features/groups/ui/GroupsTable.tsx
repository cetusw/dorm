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

import type { GroupListItem } from '../model/types'

type SortField = 'name' | 'leaderName'
type SortDirection = 'asc' | 'desc'

type Props = {
    groups: GroupListItem[]
    onEdit: (groupId: string) => void
    onDelete: (group: GroupListItem) => void
}

function normalize(value: string): string {
    return value.trim().toLowerCase()
}

export function GroupsTable({ groups, onEdit, onDelete }: Props) {
    const [search, setSearch] = useState('')
    const [sortField, setSortField] = useState<SortField>('name')
    const [sortDirection, setSortDirection] = useState<SortDirection>('asc')
    const [openedMenuId, setOpenedMenuId] = useState<string | null>(null)

    const filteredGroups = useMemo(() => {
        const query = normalize(search)
        const visibleGroups = query === ''
            ? groups
            : groups.filter((group) =>
                normalize(`${group.name} ${group.leader?.name ?? 'Не назначен'}`).includes(query),
            )

        return [...visibleGroups].sort((left, right) => {
            const leftValue = normalize(
                sortField === 'leaderName' ? left.leader?.name ?? 'Не назначен' : left.name,
            )
            const rightValue = normalize(
                sortField === 'leaderName' ? right.leader?.name ?? 'Не назначен' : right.name,
            )
            const result = leftValue.localeCompare(rightValue, 'ru')

            return sortDirection === 'asc' ? result : result * -1
        })
    }, [groups, search, sortField, sortDirection])

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
                                    label="Название"
                                    active={sortField === 'name'}
                                    direction={sortDirection}
                                    onClick={() => toggleSort('name')}
                                />
                            </Table.Th>
                            <Table.Th w="34%">
                                <SortButton
                                    label="Глава"
                                    active={sortField === 'leaderName'}
                                    direction={sortDirection}
                                    onClick={() => toggleSort('leaderName')}
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
                        {filteredGroups.map((group) => {
                            const menuOpened = openedMenuId === group.id

                            return (
                                <Table.Tr key={group.id}>
                                    <Table.Td>
                                        <Text fw={600}>{group.name}</Text>
                                    </Table.Td>
                                    <Table.Td>
                                        <Text c={group.leader ? undefined : 'dimmed'}>
                                            {group.leader?.name ?? 'Не назначен'}
                                        </Text>
                                    </Table.Td>
                                    <Table.Td>
                                        <Group justify="flex-start">
                                            <Menu
                                                opened={menuOpened}
                                                onChange={(opened) => setOpenedMenuId(opened ? group.id : null)}
                                                withinPortal
                                                position="bottom-end"
                                            >
                                                <Menu.Target>
                                                    <Button variant="subtle" px="sm" aria-label="Открыть действия">
                                                        {menuOpened ? '▴' : '▾'}
                                                    </Button>
                                                </Menu.Target>

                                                <Menu.Dropdown>
                                                    <Menu.Item onClick={() => onEdit(group.id)}>
                                                        Редактировать
                                                    </Menu.Item>
                                                    <Menu.Item color="red" onClick={() => onDelete(group)}>
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
