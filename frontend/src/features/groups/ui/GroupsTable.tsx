import { useMemo, useState } from 'react'

import { DotsThreeVerticalIcon, MagnifyingGlassIcon, PencilIcon, TrashIcon } from '@phosphor-icons/react'
import { ActionIcon, Menu, Table, Text, TextInput, UnstyledButton } from '@mantine/core'

import type { GroupListItem } from '../model/types'
import { ListTable, listTableClasses } from '../../../shared/ui/ListTable'
import classes from './GroupsTable.module.css'

type SortField = 'name' | 'leaderName'
type SortDirection = 'asc' | 'desc'

type Props = {
    groups: GroupListItem[]
    onOpen: (groupId: string) => void
    onEdit: (groupId: string) => void
    onDelete: (group: GroupListItem) => void
}

function normalize(value: string): string {
    return value.trim().toLowerCase()
}

export function GroupsTable({ groups, onOpen, onEdit, onDelete }: Props) {
    const [search, setSearch] = useState('')
    const [sortField, setSortField] = useState<SortField>('name')
    const [sortDirection, setSortDirection] = useState<SortDirection>('asc')
    const [openedMenuId, setOpenedMenuId] = useState<string | null>(null)

    const visibleGroups = useMemo(() => {
        const query = normalize(search)
        const filteredGroups = query === ''
            ? groups
            : groups.filter((group) =>
                normalize(`${group.name} ${group.leader?.name ?? 'Не назначен'}`).includes(query),
            )

        return [...filteredGroups].sort((left, right) => {
            const leftValue = normalize(sortField === 'leaderName' ? left.leader?.name ?? 'Не назначен' : left.name)
            const rightValue = normalize(sortField === 'leaderName' ? right.leader?.name ?? 'Не назначен' : right.name)
            const result = leftValue.localeCompare(rightValue, 'ru')

            return sortDirection === 'asc' ? result : result * -1
        })
    }, [groups, search, sortDirection, sortField])

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

            <ListTable minWidth={720} verticalSpacing="sm">
                <Table.Thead>
                    <Table.Tr className={listTableClasses.headerRow}>
                        <Table.Th>
                            <SortButton
                                label="Название"
                                active={sortField === 'name'}
                                direction={sortDirection}
                                onClick={() => toggleSort('name')}
                            />
                        </Table.Th>
                        <Table.Th w={220}>
                            <SortButton
                                label="Глава"
                                active={sortField === 'leaderName'}
                                direction={sortDirection}
                                onClick={() => toggleSort('leaderName')}
                            />
                        </Table.Th>
                        <Table.Th w={68} />
                    </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                    {visibleGroups.map((group) => {
                        const menuOpened = openedMenuId === group.id

                        return (
                            <Table.Tr
                                key={group.id}
                                className={`${listTableClasses.bodyRow} ${listTableClasses.interactiveRow} ${classes.row}`}
                                data-menu-open={menuOpened ? 'true' : undefined}
                                onClick={() => onOpen(group.id)}
                            >
                                <Table.Td>
                                    <Text fw={600}>{group.name}</Text>
                                </Table.Td>
                                <Table.Td>
                                    <Text c={group.leader ? undefined : 'dimmed'}>
                                        {group.leader?.name ?? 'Не назначен'}
                                    </Text>
                                </Table.Td>
                                <Table.Td className={classes.actionCell}>
                                    <div className={classes.actionCellInner}>
                                        <Menu
                                            opened={menuOpened}
                                            onChange={(opened) => setOpenedMenuId(opened ? group.id : null)}
                                            withinPortal
                                            position="bottom-end"
                                        >
                                            <Menu.Target>
                                                <ActionIcon
                                                    variant="subtle"
                                                    color="gray"
                                                    aria-label={`Действия с группой ${group.name}`}
                                                    className={classes.actionButton}
                                                    onClick={(event) => event.stopPropagation()}
                                                >
                                                    <DotsThreeVerticalIcon size={24} />
                                                </ActionIcon>
                                            </Menu.Target>
                                            <Menu.Dropdown>
                                                <Menu.Item
                                                    leftSection={<PencilIcon size={20} />}
                                                    onClick={(event) => {
                                                        event.stopPropagation()
                                                        onEdit(group.id)
                                                    }}
                                                >
                                                    Редактировать
                                                </Menu.Item>
                                                <Menu.Item
                                                    color="red"
                                                    leftSection={<TrashIcon size={20} />}
                                                    onClick={(event) => {
                                                        event.stopPropagation()
                                                        onDelete(group)
                                                    }}
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
