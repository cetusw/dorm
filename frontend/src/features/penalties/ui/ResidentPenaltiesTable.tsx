import { useState } from 'react'

import { DotsThreeVerticalIcon, TrashIcon } from '@phosphor-icons/react'
import { ActionIcon, Loader, Menu, Table, Text } from '@mantine/core'

import type { PenaltyItem } from '../model/types'
import { formatPenaltyDate, formatPenaltyWeight } from '../model/utils'
import { AppTable } from '../../../shared/ui/AppTable'
import classes from './ResidentPenaltiesTable.module.css'

type Props = {
    penalties: PenaltyItem[]
    pendingPenaltyId: string | null
    onDelete: (penaltyId: string) => void
}

export function ResidentPenaltiesTable({
    penalties,
    pendingPenaltyId,
    onDelete,
}: Props) {
    const [openedMenuId, setOpenedMenuId] = useState<string | null>(null)

    return (
        <AppTable minWidth={720} verticalSpacing="sm">
            <Table.Thead>
                <Table.Tr>
                    <Table.Th>Причина</Table.Th>
                    <Table.Th w={120}>Вес</Table.Th>
                    <Table.Th w={168}>Дата получения</Table.Th>
                    <Table.Th w={68} />
                </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
                {penalties.map((penalty) => {
                    const menuOpened = openedMenuId === penalty.id
                    const deleting = pendingPenaltyId === penalty.id

                    return (
                        <Table.Tr
                            key={penalty.id}
                            className={classes.row}
                            data-menu-open={menuOpened ? 'true' : undefined}
                        >
                            <Table.Td>
                                <Text className={classes.reason}>{penalty.reason}</Text>
                            </Table.Td>
                            <Table.Td>{formatPenaltyWeight(penalty.weight)}</Table.Td>
                            <Table.Td>{formatPenaltyDate(penalty.issued_on)}</Table.Td>
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
                                            onClick={() => onDelete(penalty.id)}
                                        >
                                            Удалить
                                        </Menu.Item>
                                    </Menu.Dropdown>
                                </Menu>
                            </Table.Td>
                        </Table.Tr>
                    )
                })}
            </Table.Tbody>
        </AppTable>
    )
}
