import { useState } from 'react'

import { ArrowUUpLeftIcon, BookOpenIcon, CheckCircleIcon, DotsThreeVerticalIcon, ListIcon, WarningIcon } from '@phosphor-icons/react'
import { ActionIcon, Menu, Table, Text } from '@mantine/core'

import type { PenaltyResidentSummary } from '../model/types'
import { formatPenaltyWeight } from '../model/utils'
import { ListTable, listTableClasses } from '../../../shared/ui/ListTable'
import classes from './PenaltiesTable.module.css'

type Props = {
    residents: PenaltyResidentSummary[]
    onOpenResident: (residentId: string) => void
    onIssuePenalty: (resident: PenaltyResidentSummary) => void
    onResolvePenalty: (resident: PenaltyResidentSummary) => void
    onIssueIndividualTask: (resident: PenaltyResidentSummary) => void
    onOpenIndividualTasks: (resident: PenaltyResidentSummary) => void
}

export function PenaltiesTable({
    residents,
    onOpenResident,
    onIssuePenalty,
    onResolvePenalty,
    onIssueIndividualTask,
    onOpenIndividualTasks,
}: Props) {
    const [openedMenuId, setOpenedMenuId] = useState<string | null>(null)

    return (
        <ListTable minWidth={520}>
            <Table.Thead>
                <Table.Tr className={listTableClasses.headerRow}>
                    <Table.Th>Имя</Table.Th>
                    <Table.Th w={180}>Предупреждения</Table.Th>
                    <Table.Th w={180}>Индивидуальные задачи</Table.Th>
                    <Table.Th w={68} />
                </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
                {residents.map((resident) => {
                    const menuOpened = openedMenuId === resident.user_id

                    return (
                        <Table.Tr
                            key={resident.user_id}
                            className={`${listTableClasses.interactiveRow} ${listTableClasses.bodyRow} ${classes.row} ${resident.threshold_reached ? classes.thresholdRow : ''}`}
                            data-menu-open={menuOpened ? 'true' : undefined}
                            tabIndex={0}
                            aria-label={`Открыть предупреждения жителя ${resident.full_name}`}
                            onClick={() => onOpenResident(resident.user_id)}
                            onKeyDown={(event) => {
                                if (event.key === 'Enter' || event.key === ' ') {
                                    event.preventDefault()
                                    onOpenResident(resident.user_id)
                                }
                            }}
                        >
                            <Table.Td>
                                <Text fw={500} inherit>{resident.full_name}</Text>
                            </Table.Td>
                            <Table.Td>
                                <Text inherit>{formatPenaltyWeight(resident.total_weight)}</Text>
                            </Table.Td>
                            <Table.Td>
                                <Text inherit>{resident.individual_task_count}</Text>
                            </Table.Td>
                            <Table.Td className={classes.actionCell}>
                                <div className={classes.actionCellInner}>
                                    <Menu
                                        opened={menuOpened}
                                        onChange={(opened) => setOpenedMenuId(opened ? resident.user_id : null)}
                                        withinPortal
                                        position="bottom-end"
                                    >
                                        <Menu.Target>
                                            <ActionIcon
                                                variant="subtle"
                                                color="gray"
                                                aria-label={`Действия с предупреждениями жителя ${resident.full_name}`}
                                                className={classes.actionButton}
                                                onClick={(event) => event.stopPropagation()}
                                                onKeyDown={(event) => event.stopPropagation()}
                                            >
                                                <DotsThreeVerticalIcon size={24} />
                                            </ActionIcon>
                                        </Menu.Target>

                                        <Menu.Dropdown>
                                            <Menu.Item
                                                leftSection={<WarningIcon size={24} />}
                                                onClick={(event) => {
                                                    event.stopPropagation()
                                                    onIssuePenalty(resident)
                                                }}
                                            >
                                                Выдать предупреждение
                                            </Menu.Item>
                                            <Menu.Item
                                                leftSection={<ArrowUUpLeftIcon size={24} />}
                                                onClick={(event) => {
                                                    event.stopPropagation()
                                                    onResolvePenalty(resident)
                                                }}
                                            >
                                                Погасить предупреждение
                                            </Menu.Item>
                                            <Menu.Item
                                                leftSection={<BookOpenIcon size={24} />}
                                                onClick={(event) => {
                                                    event.stopPropagation()
                                                    onOpenResident(resident.user_id)
                                                }}
                                            >
                                                Предупреждения
                                            </Menu.Item>
                                            <Menu.Divider />
                                            <Menu.Item leftSection={<CheckCircleIcon size={24} />} onClick={(event) => { event.stopPropagation(); onIssueIndividualTask(resident) }}>
                                                Выдать индивидуальную задачу
                                            </Menu.Item>
                                            <Menu.Item leftSection={<ListIcon size={24} />} onClick={(event) => { event.stopPropagation(); onOpenIndividualTasks(resident) }}>
                                                Индивидуальные задачи
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
    )
}
