import { useMemo, useState } from 'react'

import { DotsThreeVerticalIcon, MagnifyingGlassIcon, PencilIcon, TrashIcon } from '@phosphor-icons/react'
import { ActionIcon, Menu, Table, Text, TextInput, UnstyledButton } from '@mantine/core'

import type { ResidentListItem } from '../model/types'
import { ListTable, listTableClasses } from '../../../shared/ui/ListTable'
import classes from './ResidentsTable.module.css'

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

    const visibleResidents = useMemo(() => {
        const query = normalize(search)
        const filteredResidents = query === ''
            ? residents
            : residents.filter((resident) =>
                normalize(`${resident.name} ${resident.room_number}`).includes(query),
            )

        return [...filteredResidents].sort((left, right) => {
            const leftValue = normalize(left[sortField])
            const rightValue = normalize(right[sortField])
            const result = leftValue.localeCompare(rightValue, 'ru')

            return sortDirection === 'asc' ? result : result * -1
        })
    }, [residents, search, sortDirection, sortField])

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
                                label="Имя"
                                active={sortField === 'name'}
                                direction={sortDirection}
                                onClick={() => toggleSort('name')}
                            />
                        </Table.Th>
                        <Table.Th w={220}>
                            <SortButton
                                label="Комната"
                                active={sortField === 'room_number'}
                                direction={sortDirection}
                                onClick={() => toggleSort('room_number')}
                            />
                        </Table.Th>
                        <Table.Th w={68} />
                    </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                    {visibleResidents.map((resident) => {
                        const menuOpened = openedMenuId === resident.id

                        return (
                            <Table.Tr
                                key={resident.id}
                                className={`${listTableClasses.bodyRow} ${classes.row}`}
                                data-menu-open={menuOpened ? 'true' : undefined}
                            >
                                <Table.Td>
                                    <Text fw={600}>{resident.name}</Text>
                                </Table.Td>
                                <Table.Td>
                                    <Text>{resident.room_number}</Text>
                                </Table.Td>
                                <Table.Td className={classes.actionCell}>
                                    <div className={classes.actionCellInner}>
                                        <Menu
                                            opened={menuOpened}
                                            onChange={(opened) => setOpenedMenuId(opened ? resident.id : null)}
                                            withinPortal
                                            position="bottom-end"
                                        >
                                            <Menu.Target>
                                                <ActionIcon
                                                    variant="subtle"
                                                    color="gray"
                                                    aria-label={`Действия с жителем ${resident.name}`}
                                                    className={classes.actionButton}
                                                >
                                                    <DotsThreeVerticalIcon size={24} />
                                                </ActionIcon>
                                            </Menu.Target>

                                            <Menu.Dropdown>
                                                <Menu.Item
                                                    leftSection={<PencilIcon size={20} />}
                                                    onClick={() => onEdit(resident.id)}
                                                >
                                                    Редактировать
                                                </Menu.Item>
                                                <Menu.Item
                                                    color="red"
                                                    leftSection={<TrashIcon size={20} />}
                                                    onClick={() => onDelete(resident)}
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
