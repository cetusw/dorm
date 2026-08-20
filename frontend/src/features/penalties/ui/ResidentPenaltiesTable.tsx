import { useState } from 'react'

import { DotsThreeVerticalIcon, PencilIcon, TrashIcon } from '@phosphor-icons/react'
import { ActionIcon, Loader, Menu, Table, Text } from '@mantine/core'

import type { PenaltyEntryItem } from '../model/types'
import { formatPenaltyDate, formatPenaltyEntryWeight } from '../model/utils'
import { ListTable, listTableClasses } from '../../../shared/ui/ListTable'
import classes from './ResidentPenaltiesTable.module.css'

type Props = {
    entries: PenaltyEntryItem[]
    pendingEntryId?: string | null
    onDelete?: (entry: PenaltyEntryItem) => void
    onEdit?: (entry: PenaltyEntryItem) => void
    dateLabel?: string
    minWidth?: number
}

export function ResidentPenaltiesTable({
    entries,
    pendingEntryId = null,
    onDelete,
    onEdit,
    dateLabel = 'Дата',
    minWidth = 720,
}: Props) {
    const [openedMenuId, setOpenedMenuId] = useState<string | null>(null)
    const hasActions = typeof onDelete === 'function' || typeof onEdit === 'function'
    const maxWeightWidthCh = Math.max(
        4,
        'Вес'.length,
        ...entries.map((entry) => formatPenaltyEntryWeight(entry).length),
    )
    const weightColumnStyle = {
        width: `${maxWeightWidthCh}ch`,
        minWidth: `${maxWeightWidthCh}ch`,
        maxWidth: `${maxWeightWidthCh}ch`,
    }

    return (
        <ListTable minWidth={minWidth} verticalSpacing="sm">
            <Table.Thead>
                <Table.Tr className={listTableClasses.headerRow}>
                    <Table.Th className={classes.reasonColumn}>Причина</Table.Th>
                    <Table.Th className={classes.weightColumn} style={weightColumnStyle}>Вес</Table.Th>
                    <Table.Th className={classes.dateColumn}>{dateLabel}</Table.Th>
                    {hasActions ? <Table.Th w={68} /> : null}
                </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
                {entries.map((entry) => {
                    const menuOpened = openedMenuId === entry.id
                    const pending = pendingEntryId === entry.id

                    return (
                        <Table.Tr
                            key={entry.id}
                            className={`${listTableClasses.bodyRow} ${classes.row}`}
                            data-entry-type={entry.type}
                            data-menu-open={menuOpened ? 'true' : undefined}
                        >
                            <Table.Td className={classes.reasonColumn}>
                                <Text className={classes.reason}>{entry.reason}</Text>
                            </Table.Td>
                            <Table.Td className={classes.weightColumn} style={weightColumnStyle}>
                                <Text className={classes.weightValue}>
                                    {formatPenaltyEntryWeight(entry)}
                                </Text>
                            </Table.Td>
                            <Table.Td className={classes.dateColumn}>{formatPenaltyDate(entry.created_at)}</Table.Td>
                            {hasActions ? (
                                <Table.Td>
                                    <Menu
                                        opened={menuOpened}
                                        onChange={(opened) => setOpenedMenuId(opened ? entry.id : null)}
                                        withinPortal
                                        position="bottom-end"
                                    >
                                        <Menu.Target>
                                            <ActionIcon
                                                variant="subtle"
                                                color="gray"
                                                aria-label={`Действия с предупреждением от ${formatPenaltyDate(entry.created_at)}`}
                                                className={classes.actionButton}
                                                disabled={pending}
                                            >
                                                {pending ? <Loader size={16} /> : <DotsThreeVerticalIcon size={18} />}
                                            </ActionIcon>
                                        </Menu.Target>
                                        <Menu.Dropdown>
                                            {typeof onEdit === 'function' ? (
                                                <Menu.Item
                                                    leftSection={<PencilIcon size={18} />}
                                                    onClick={() => onEdit(entry)}
                                                >
                                                    Редактировать
                                                </Menu.Item>
                                            ) : null}
                                            {typeof onDelete === 'function' ? (
                                                <Menu.Item
                                                    color="red"
                                                    leftSection={<TrashIcon size={18} />}
                                                    onClick={() => onDelete(entry)}
                                                >
                                                    Удалить
                                                </Menu.Item>
                                            ) : null}
                                        </Menu.Dropdown>
                                    </Menu>
                                </Table.Td>
                            ) : null}
                        </Table.Tr>
                    )
                })}
            </Table.Tbody>
        </ListTable>
    )
}
