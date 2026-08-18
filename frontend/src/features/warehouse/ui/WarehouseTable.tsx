import { useMemo, useState } from 'react'

import {
    BookOpenIcon,
    DotsThreeVerticalIcon,
    MagnifyingGlassIcon,
    MinusIcon,
    PencilIcon,
    PlusIcon,
    TrashIcon,
} from '@phosphor-icons/react'
import { ActionIcon, Menu, Table, Text, TextInput, UnstyledButton } from '@mantine/core'

import { ListTable, listTableClasses } from '../../../shared/ui/ListTable'
import type { WarehouseItem } from '../model/types'
import classes from './WarehouseTable.module.css'

type Props = {
    items: WarehouseItem[]
    onOpenHistory: (item: WarehouseItem) => void
    onAdd: (item: WarehouseItem) => void
    onWriteOff: (item: WarehouseItem) => void
    onEdit: (item: WarehouseItem) => void
    onDelete: (item: WarehouseItem) => void
}

type SortField = 'name' | 'quantity'
type SortDirection = 'asc' | 'desc'

function normalize(value: string): string {
    return value.trim().toLowerCase()
}

export function WarehouseTable({ items, onOpenHistory, onAdd, onWriteOff, onEdit, onDelete }: Props) {
    const [search, setSearch] = useState('')
    const [sortField, setSortField] = useState<SortField>('name')
    const [sortDirection, setSortDirection] = useState<SortDirection>('asc')
    const [openedMenuId, setOpenedMenuId] = useState<string | null>(null)

    const visibleItems = useMemo(() => {
        const query = normalize(search)
        const filteredItems = query === ''
            ? items
            : items.filter((item) => normalize(item.name).includes(query))

        return [...filteredItems].sort((left, right) => {
            const leftIsZero = left.quantity === 0
            const rightIsZero = right.quantity === 0

            if (leftIsZero !== rightIsZero) {
                return leftIsZero ? 1 : -1
            }

            if (sortField === 'quantity') {
                const result = left.quantity - right.quantity

                if (result !== 0) {
                    return sortDirection === 'asc' ? result : result * -1
                }
            } else {
                const result = normalize(left.name).localeCompare(normalize(right.name), 'ru')

                if (result !== 0) {
                    return sortDirection === 'asc' ? result : result * -1
                }
            }

            return normalize(left.name).localeCompare(normalize(right.name), 'ru')
        })
    }, [items, search, sortDirection, sortField])

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
                                label="Количество"
                                active={sortField === 'quantity'}
                                direction={sortDirection}
                                onClick={() => toggleSort('quantity')}
                            />
                        </Table.Th>
                        <Table.Th w={68} />
                    </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                    {visibleItems.map((item) => {
                        const menuOpened = openedMenuId === item.id

                        return (
                            <Table.Tr
                                key={item.id}
                                className={`${listTableClasses.bodyRow} ${listTableClasses.interactiveRow} ${classes.row}`}
                                data-menu-open={menuOpened ? 'true' : undefined}
                                data-zero-quantity={item.quantity === 0 ? 'true' : undefined}
                                onClick={() => onOpenHistory(item)}
                            >
                                <Table.Td>
                                    <Text fw={500}>{item.name}</Text>
                                </Table.Td>
                                <Table.Td>
                                    <Text>{item.quantity}</Text>
                                </Table.Td>
                                <Table.Td className={classes.actionCell}>
                                    <div className={classes.actionCellInner}>
                                        <Menu
                                            opened={menuOpened}
                                            onChange={(opened) => setOpenedMenuId(opened ? item.id : null)}
                                            withinPortal
                                            position="bottom-end"
                                        >
                                            <Menu.Target>
                                                <ActionIcon
                                                    variant="subtle"
                                                    color="gray"
                                                    aria-label={`Действия с позицией ${item.name}`}
                                                    className={classes.actionButton}
                                                    onClick={(event) => event.stopPropagation()}
                                                >
                                                    <DotsThreeVerticalIcon size={24} />
                                                </ActionIcon>
                                            </Menu.Target>

                                            <Menu.Dropdown>
                                                <Menu.Item
                                                    leftSection={<PlusIcon size={24} />}
                                                    onClick={(event) => {
                                                        event.stopPropagation()
                                                        onAdd(item)
                                                    }}
                                                >
                                                    Добавить
                                                </Menu.Item>
                                                {item.quantity > 0 ? (
                                                    <Menu.Item
                                                        leftSection={<MinusIcon size={24} />}
                                                        onClick={(event) => {
                                                            event.stopPropagation()
                                                            onWriteOff(item)
                                                        }}
                                                    >
                                                        Списать
                                                    </Menu.Item>
                                                ) : null}
                                                <Menu.Item
                                                    leftSection={<PencilIcon size={24} />}
                                                    onClick={(event) => {
                                                        event.stopPropagation()
                                                        onEdit(item)
                                                    }}
                                                >
                                                    Редактировать
                                                </Menu.Item>
                                                <Menu.Item
                                                    leftSection={<BookOpenIcon size={24} />}
                                                    onClick={(event) => {
                                                        event.stopPropagation()
                                                        onOpenHistory(item)
                                                    }}
                                                >
                                                    История
                                                </Menu.Item>
                                                <Menu.Item
                                                    color="red"
                                                    leftSection={<TrashIcon size={24} />}
                                                    onClick={(event) => {
                                                        event.stopPropagation()
                                                        onDelete(item)
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
