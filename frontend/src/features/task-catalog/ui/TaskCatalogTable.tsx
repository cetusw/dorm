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

import type { TaskListItem } from '../model/types'

type SortField = 'title' | 'cost' | 'frequency' | 'areaName'
type SortDirection = 'asc' | 'desc'

type Props = {
    tasks: TaskListItem[]
    onEdit: (taskId: string) => void
    onDelete: (task: TaskListItem) => void
}

function normalize(value: string): string {
    return value.trim().toLowerCase()
}

export function TaskCatalogTable({ tasks, onEdit, onDelete }: Props) {
    const [search, setSearch] = useState('')
    const [sortField, setSortField] = useState<SortField>('title')
    const [sortDirection, setSortDirection] = useState<SortDirection>('asc')
    const [openedMenuId, setOpenedMenuId] = useState<string | null>(null)

    const filteredTasks = useMemo(() => {
        const query = normalize(search)
        const visibleTasks = query === ''
            ? tasks
            : tasks.filter((task) =>
                normalize(`${task.title} ${task.cost} ${task.frequency} ${task.area.name}`).includes(query),
            )

        return [...visibleTasks].sort((left, right) => {
            if (sortField === 'cost' || sortField === 'frequency') {
                const result = left[sortField] - right[sortField]
                return sortDirection === 'asc' ? result : result * -1
            }

            const leftValue = normalize(sortField === 'areaName' ? left.area.name : left.title)
            const rightValue = normalize(sortField === 'areaName' ? right.area.name : right.title)
            const result = leftValue.localeCompare(rightValue, 'ru')
            return sortDirection === 'asc' ? result : result * -1
        })
    }, [tasks, search, sortField, sortDirection])

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
                <Table horizontalSpacing="lg" verticalSpacing="md" highlightOnHover miw={860}>
                    <Table.Thead>
                        <Table.Tr>
                            <Table.Th w="34%">
                                <SortButton
                                    label="Название"
                                    active={sortField === 'title'}
                                    direction={sortDirection}
                                    onClick={() => toggleSort('title')}
                                />
                            </Table.Th>
                            <Table.Th w="14%">
                                <SortButton
                                    label="Стоимость"
                                    active={sortField === 'cost'}
                                    direction={sortDirection}
                                    onClick={() => toggleSort('cost')}
                                />
                            </Table.Th>
                            <Table.Th w="14%">
                                <SortButton
                                    label="Частота"
                                    active={sortField === 'frequency'}
                                    direction={sortDirection}
                                    onClick={() => toggleSort('frequency')}
                                />
                            </Table.Th>
                            <Table.Th w="24%">
                                <SortButton
                                    label="Территория"
                                    active={sortField === 'areaName'}
                                    direction={sortDirection}
                                    onClick={() => toggleSort('areaName')}
                                />
                            </Table.Th>
                            <Table.Th w="14%">
                                <Text size="sm" fw={700}>
                                    Действия
                                </Text>
                            </Table.Th>
                        </Table.Tr>
                    </Table.Thead>
                    <Table.Tbody>
                        {filteredTasks.map((task) => {
                            const menuOpened = openedMenuId === task.id

                            return (
                                <Table.Tr key={task.id}>
                                    <Table.Td>
                                        <Text fw={600}>{task.title}</Text>
                                    </Table.Td>
                                    <Table.Td>{task.cost}</Table.Td>
                                    <Table.Td>{task.frequency}</Table.Td>
                                    <Table.Td>{task.area.name}</Table.Td>
                                    <Table.Td>
                                        <Group justify="flex-start">
                                            <Menu
                                                opened={menuOpened}
                                                onChange={(opened) => setOpenedMenuId(opened ? task.id : null)}
                                                withinPortal
                                                position="bottom-end"
                                            >
                                                <Menu.Target>
                                                    <Button variant="subtle" px="sm" aria-label="Открыть действия">
                                                        {menuOpened ? '▴' : '▾'}
                                                    </Button>
                                                </Menu.Target>
                                                <Menu.Dropdown>
                                                    <Menu.Item onClick={() => onEdit(task.id)}>
                                                        Редактировать
                                                    </Menu.Item>
                                                    <Menu.Item color="red" onClick={() => onDelete(task)}>
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
