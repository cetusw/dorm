import { useState } from 'react'

import { DotsThreeVerticalIcon, TrashIcon } from '@phosphor-icons/react'
import { ActionIcon, Loader, Menu, Table, Text } from '@mantine/core'

import type { PenaltyItem } from '../model/types'
import { formatPenaltyDate, formatPenaltyWeight } from '../model/utils'
import { ListTable, listTableClasses } from '../../../shared/ui/ListTable'
import classes from './ResidentPenaltiesTable.module.css'

type Props = {
    penalties: PenaltyItem[]
    pendingPenaltyId?: string | null
    onDelete?: (penaltyId: string) => void
    dateLabel?: string
    minWidth?: number
}

export function ResidentPenaltiesTable({
    penalties,
    pendingPenaltyId = null,
    onDelete,
    dateLabel = 'Дата получения',
    minWidth = 720,
}: Props) {
    const [openedMenuId, setOpenedMenuId] = useState<string | null>(null)
    const hasActions = typeof onDelete === 'function'

    return (
        <ListTable minWidth={minWidth} verticalSpacing="sm">
            <Table.Thead>
                <Table.Tr className={listTableClasses.headerRow}>
                    <Table.Th className={classes.reasonColumn}>Причина</Table.Th>
                    <Table.Th className={classes.weightColumn}>Вес</Table.Th>
                    <Table.Th className={classes.dateColumn}>{dateLabel}</Table.Th>
                    {hasActions ? <Table.Th w={68} /> : null}
                </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
                {penalties.map((penalty) => {
                    const menuOpened = openedMenuId === penalty.id
                    const deleting = pendingPenaltyId === penalty.id

                    return (
                        <Table.Tr
                            key={penalty.id}
                            className={`${listTableClasses.bodyRow} ${classes.row}`}
                            data-menu-open={menuOpened ? 'true' : undefined}
                        >
                            <Table.Td className={classes.reasonColumn}>
                                <Text className={classes.reason}>{penalty.reason}</Text>
                            </Table.Td>
                            <Table.Td className={classes.weightColumn}>{formatPenaltyWeight(penalty.weight)}</Table.Td>
                            <Table.Td className={classes.dateColumn}>{formatPenaltyDate(penalty.issued_on)}</Table.Td>
                            {hasActions ? (
                                <Table.Td>
                                    <Menu
                                        opened={menuOpened}
                                        onChange={(opened) => setOpenedMenuId(opened ? penalty.id : null)}
                                        withinPortal
                                        position="bottom-end"
                                    >
                                        <Menu.Target>
                                            <ActionIcon
                                                variant="subtle"
                                                color="gray"
                                                aria-label={`Действия с предупреждением от ${formatPenaltyDate(penalty.issued_on)}`}
                                                className={classes.actionButton}
                                                disabled={deleting}
                                            >
                                                {deleting ? <Loader size={16} /> : <DotsThreeVerticalIcon size={18} />}
                                            </ActionIcon>
                                        </Menu.Target>
                                        <Menu.Dropdown>
                                            <Menu.Item
                                                color="red"
                                                leftSection={<TrashIcon size={20} />}
                                                onClick={() => onDelete?.(penalty.id)}
                                            >
                                                Удалить
                                            </Menu.Item>
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
