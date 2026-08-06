import { useMemo, useState } from 'react'

import { DotsThreeVerticalIcon, MagnifyingGlassIcon, PencilIcon, TrashIcon } from '@phosphor-icons/react'
import { ActionIcon, Menu, Table, Text, TextInput, UnstyledButton } from '@mantine/core'

import type { DormitoryListItem } from '../model/types'
import { ListTable, listTableClasses } from '../../../shared/ui/ListTable'
import classes from './DormitoriesTable.module.css'

type Props = {
    dormitories: DormitoryListItem[]
    onDelete: (dormitory: DormitoryListItem) => void
    onEdit: (dormitoryId: number) => void
}

type SortField = 'name' | 'address' | 'leaderName'
type SortDirection = 'asc' | 'desc'

function normalize(value: string): string {
    return value.trim().toLowerCase()
}

export function DormitoriesTable({ dormitories, onDelete, onEdit }: Props) {
    const [search, setSearch] = useState('')
    const [sortField, setSortField] = useState<SortField>('name')
    const [sortDirection, setSortDirection] = useState<SortDirection>('asc')
    const [openedMenuId, setOpenedMenuId] = useState<number | null>(null)

    const visibleDormitories = useMemo(() => {
        const query = normalize(search)
        const filteredDormitories = query === ''
            ? dormitories
            : dormitories.filter((dormitory) =>
                normalize(`${dormitory.name} ${dormitory.address} ${dormitory.leader?.name ?? 'Не назначен'}`).includes(query),
            )

        return [...filteredDormitories].sort((left, right) => {
            const leftValue = normalize(
                sortField === 'leaderName'
                    ? left.leader?.name ?? 'Не назначен'
                    : left[sortField],
            )
            const rightValue = normalize(
                sortField === 'leaderName'
                    ? right.leader?.name ?? 'Не назначен'
                    : right[sortField],
            )
            const result = leftValue.localeCompare(rightValue, 'ru')

            return sortDirection === 'asc' ? result : result * -1
        })
    }, [dormitories, search, sortDirection, sortField])

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

            <ListTable minWidth={920} verticalSpacing="sm">
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
                        <Table.Th>
                            <SortButton
                                label="Адрес"
                                active={sortField === 'address'}
                                direction={sortDirection}
                                onClick={() => toggleSort('address')}
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
                    {visibleDormitories.map((dormitory) => {
                        const menuOpened = openedMenuId === dormitory.id

                        return (
                            <Table.Tr
                                key={dormitory.id}
                                className={`${listTableClasses.bodyRow} ${classes.row}`}
                                data-menu-open={menuOpened ? 'true' : undefined}
                            >
                                <Table.Td>
                                    <Text fw={600}>{dormitory.name}</Text>
                                </Table.Td>
                                <Table.Td>
                                    <Text>{dormitory.address}</Text>
                                </Table.Td>
                                <Table.Td>
                                    <Text c={dormitory.leader ? undefined : 'dimmed'}>
                                        {dormitory.leader?.name ?? 'Не назначен'}
                                    </Text>
                                </Table.Td>
                                <Table.Td className={classes.actionCell}>
                                    <div className={classes.actionCellInner}>
                                        <Menu
                                            opened={menuOpened}
                                            onChange={(opened) => setOpenedMenuId(opened ? dormitory.id : null)}
                                            withinPortal
                                            position="bottom-end"
                                        >
                                            <Menu.Target>
                                                <ActionIcon
                                                    variant="subtle"
                                                    color="gray"
                                                    aria-label={`Действия с общежитием ${dormitory.name}`}
                                                    className={classes.actionButton}
                                                >
                                                    <DotsThreeVerticalIcon size={24} />
                                                </ActionIcon>
                                            </Menu.Target>
                                            <Menu.Dropdown>
                                                <Menu.Item
                                                    leftSection={<PencilIcon size={20} />}
                                                    onClick={() => onEdit(dormitory.id)}
                                                >
                                                    Редактировать
                                                </Menu.Item>
                                                <Menu.Item
                                                    color="red"
                                                    leftSection={<TrashIcon size={20} />}
                                                    onClick={() => onDelete(dormitory)}
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
