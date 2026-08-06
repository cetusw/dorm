import { useMemo, useState } from 'react'

import { DotsThreeVerticalIcon, MagnifyingGlassIcon, PencilIcon, TrashIcon } from '@phosphor-icons/react'
import { ActionIcon, Menu, Table, Text, TextInput, UnstyledButton } from '@mantine/core'

import type { TaskListItem } from '../model/types'
import { formatRecurrenceInterval } from '../model/recurrence'
import { ListTable, listTableClasses } from '../../../shared/ui/ListTable'
import classes from './TaskCatalogTable.module.css'

type SortField = 'title' | 'cost' | 'recurrenceInterval' | 'areaName'
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

    const visibleTasks = useMemo(() => {
        const query = normalize(search)
        const filteredTasks = query === ''
            ? tasks
            : tasks.filter((task) =>
                normalize(`${task.title} ${task.cost} ${formatRecurrenceInterval(task.recurrenceInterval)} ${task.area.name}`).includes(query),
            )

        return [...filteredTasks].sort((left, right) => {
            if (sortField === 'cost' || sortField === 'recurrenceInterval') {
                const result = left[sortField] - right[sortField]
                return sortDirection === 'asc' ? result : result * -1
            }

            const leftValue = normalize(sortField === 'areaName' ? left.area.name : left.title)
            const rightValue = normalize(sortField === 'areaName' ? right.area.name : right.title)
            const result = leftValue.localeCompare(rightValue, 'ru')

            return sortDirection === 'asc' ? result : result * -1
        })
    }, [tasks, search, sortDirection, sortField])

    function toggleSort(nextField: SortField) {
        if (sortField === nextField) {
            setSortDirection((current) => (current === 'asc' ? 'desc' : 'asc'))
            return
        }

        setSortField(nextField)
        setSortDirection('asc')
    }

    return (
        <>
            <TextInput
                value={search}
                onChange={(event) => setSearch(event.currentTarget.value)}
                placeholder="Поиск"
                mb="md"
                leftSection={<MagnifyingGlassIcon size={18} />}
            />

            <ListTable minWidth={860} verticalSpacing="sm">
                <Table.Thead>
                    <Table.Tr className={listTableClasses.headerRow}>
                        <Table.Th>
                            <SortButton
                                label="Название"
                                active={sortField === 'title'}
                                direction={sortDirection}
                                onClick={() => toggleSort('title')}
                            />
                        </Table.Th>
                        <Table.Th w={130}>
                            <SortButton
                                label="Стоимость"
                                active={sortField === 'cost'}
                                direction={sortDirection}
                                onClick={() => toggleSort('cost')}
                            />
                        </Table.Th>
                        <Table.Th w={150}>
                            <SortButton
                                label="Частота"
                                active={sortField === 'recurrenceInterval'}
                                direction={sortDirection}
                                onClick={() => toggleSort('recurrenceInterval')}
                            />
                        </Table.Th>
                        <Table.Th w={220}>
                            <SortButton
                                label="Территория"
                                active={sortField === 'areaName'}
                                direction={sortDirection}
                                onClick={() => toggleSort('areaName')}
                            />
                        </Table.Th>
                        <Table.Th w={68} />
                    </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                    {visibleTasks.map((task) => {
                        const menuOpened = openedMenuId === task.id

                        return (
                            <Table.Tr
                                key={task.id}
                                className={`${listTableClasses.bodyRow} ${classes.row}`}
                                data-menu-open={menuOpened ? 'true' : undefined}
                            >
                                <Table.Td>
                                    <Text fw={600}>{task.title}</Text>
                                </Table.Td>
                                <Table.Td>
                                    <Text>{task.cost}</Text>
                                </Table.Td>
                                <Table.Td>
                                    <Text>{formatRecurrenceInterval(task.recurrenceInterval)}</Text>
                                </Table.Td>
                                <Table.Td>
                                    <Text>{task.area.name}</Text>
                                </Table.Td>
                                <Table.Td className={classes.actionCell}>
                                    <div className={classes.actionCellInner}>
                                        <Menu
                                            opened={menuOpened}
                                            onChange={(opened) => setOpenedMenuId(opened ? task.id : null)}
                                            withinPortal
                                            position="bottom-end"
                                        >
                                            <Menu.Target>
                                                <ActionIcon
                                                    variant="subtle"
                                                    color="gray"
                                                    aria-label={`Действия с задачей ${task.title}`}
                                                    className={classes.actionButton}
                                                >
                                                    <DotsThreeVerticalIcon size={24} />
                                                </ActionIcon>
                                            </Menu.Target>
                                            <Menu.Dropdown>
                                                <Menu.Item
                                                    leftSection={<PencilIcon size={20} />}
                                                    onClick={() => onEdit(task.id)}
                                                >
                                                    Редактировать
                                                </Menu.Item>
                                                <Menu.Item
                                                    color="red"
                                                    leftSection={<TrashIcon size={20} />}
                                                    onClick={() => onDelete(task)}
                                                >
                                                    Удалить
                                                </Menu.Item>
                                            </Menu.Dropdown>
                                        </Menu>
                                    </div>
                                </Table.Td>
                            </Table.Tr>
                        )
                    })}
                </Table.Tbody>
            </ListTable>
        </>
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
            <Text fw={700} inherit className={classes.sortButtonLabel}>
                {label}
                <span className={classes.sortButtonIndicator}>
                    {active ? (direction === 'asc' ? '▲' : '▼') : '↕'}
                </span>
            </Text>
        </UnstyledButton>
    )
}
