import { useMemo, useState } from 'react'

import { DotsThreeVerticalIcon, MagnifyingGlassIcon, PencilIcon, TrashIcon } from '@phosphor-icons/react'
import { ActionIcon, Menu, Table, Text, TextInput, UnstyledButton } from '@mantine/core'

import type { AreaListItem } from '../model/types'
import { ListTable, listTableClasses } from '../../../shared/ui/ListTable'
import classes from './AreasTable.module.css'

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

    const visibleAreas = useMemo(() => {
        const query = normalize(search)
        const filteredAreas = query === ''
            ? areas
            : areas.filter((area) =>
                normalize(`${area.name} ${area.floor ?? ''} ${area.group?.name ?? 'Общая территория'}`).includes(query),
            )

        return [...filteredAreas].sort((left, right) => {
            if (sortField === 'floor') {
                const leftFloor = left.floor ?? Number.NEGATIVE_INFINITY
                const rightFloor = right.floor ?? Number.NEGATIVE_INFINITY
                const result = leftFloor - rightFloor
                return sortDirection === 'asc' ? result : result * -1
            }

            const leftValue = normalize(sortField === 'groupName' ? left.group?.name ?? 'Общая территория' : left.name)
            const rightValue = normalize(sortField === 'groupName' ? right.group?.name ?? 'Общая территория' : right.name)
            const result = leftValue.localeCompare(rightValue, 'ru')

            return sortDirection === 'asc' ? result : result * -1
        })
    }, [areas, search, sortDirection, sortField])

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

            <ListTable minWidth={840} verticalSpacing="sm">
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
                        <Table.Th w={140}>
                            <SortButton
                                label="Этаж"
                                active={sortField === 'floor'}
                                direction={sortDirection}
                                onClick={() => toggleSort('floor')}
                            />
                        </Table.Th>
                        <Table.Th w={220}>
                            <SortButton
                                label="Группа"
                                active={sortField === 'groupName'}
                                direction={sortDirection}
                                onClick={() => toggleSort('groupName')}
                            />
                        </Table.Th>
                        <Table.Th w={68} />
                    </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                    {visibleAreas.map((area) => {
                        const menuOpened = openedMenuId === area.id

                        return (
                            <Table.Tr
                                key={area.id}
                                className={`${listTableClasses.bodyRow} ${classes.row}`}
                                data-menu-open={menuOpened ? 'true' : undefined}
                            >
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
                                <Table.Td className={classes.actionCell}>
                                    <div className={classes.actionCellInner}>
                                        <Menu
                                            opened={menuOpened}
                                            onChange={(opened) => setOpenedMenuId(opened ? area.id : null)}
                                            withinPortal
                                            position="bottom-end"
                                        >
                                            <Menu.Target>
                                                <ActionIcon
                                                    variant="subtle"
                                                    color="gray"
                                                    aria-label={`Действия с территорией ${area.name}`}
                                                    className={classes.actionButton}
                                                >
                                                    <DotsThreeVerticalIcon size={24} />
                                                </ActionIcon>
                                            </Menu.Target>
                                            <Menu.Dropdown>
                                                <Menu.Item
                                                    leftSection={<PencilIcon size={20} />}
                                                    onClick={() => onEdit(area.id)}
                                                >
                                                    Редактировать
                                                </Menu.Item>
                                                <Menu.Item
                                                    color="red"
                                                    leftSection={<TrashIcon size={20} />}
                                                    onClick={() => onDelete(area)}
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
