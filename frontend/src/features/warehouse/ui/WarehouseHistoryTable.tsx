import { useEffect, useRef, useState } from 'react'

import { DotsThreeVerticalIcon, PencilIcon, TrashIcon } from '@phosphor-icons/react'
import { ActionIcon, Menu, Table, Text, Tooltip } from '@mantine/core'

import type { WarehouseMovement } from '../model/types'
import { formatWarehouseMovementDate, formatWarehouseMovementQuantity } from '../model/utils'
import { ListTable, listTableClasses } from '../../../shared/ui/ListTable'
import classes from './WarehouseHistoryTable.module.css'

type Props = {
    movements: WarehouseMovement[]
    onEdit: (movement: WarehouseMovement) => void
    onDelete: (movement: WarehouseMovement) => void
}

export function WarehouseHistoryTable({ movements, onEdit, onDelete }: Props) {
    const [openedMenuId, setOpenedMenuId] = useState<string | null>(null)

    return (
        <ListTable minWidth={720} verticalSpacing="sm">
            <Table.Thead>
                <Table.Tr className={listTableClasses.headerRow}>
                    <Table.Th w={120}>
                        <Text fw={700}>Количество</Text>
                    </Table.Th>
                    <Table.Th className={classes.commentCell}>
                        <Text fw={700}>Комментарий</Text>
                    </Table.Th>
                    <Table.Th w={140} className={classes.dateCell}>
                        <Text fw={700}>Дата</Text>
                    </Table.Th>
                    <Table.Th w={68} />
                </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
                {movements.map((movement) => {
                    const menuOpened = openedMenuId === movement.id
                    const comment = movement.comment?.trim() ?? ''

                    return (
                        <Table.Tr
                            key={movement.id}
                            className={`${listTableClasses.bodyRow} ${classes.row}`}
                            data-menu-open={menuOpened ? 'true' : undefined}
                            data-movement-type={movement.type}
                        >
                            <Table.Td>
                                <Text className={classes.movementText}>{formatWarehouseMovementQuantity(movement)}</Text>
                            </Table.Td>
                            <Table.Td className={classes.commentCell}>
                                {comment ? (
                                    <OverflowTooltipText text={comment} />
                                ) : (
                                    <Text c="dimmed">-</Text>
                                )}
                            </Table.Td>
                            <Table.Td className={classes.dateCell}>
                                <Text className={classes.movementText}>{formatWarehouseMovementDate(movement.created_at)}</Text>
                            </Table.Td>
                            <Table.Td className={classes.actionCell}>
                                <div className={classes.actionCellInner}>
                                    <Menu
                                        opened={menuOpened}
                                        onChange={(opened) => setOpenedMenuId(opened ? movement.id : null)}
                                        withinPortal
                                        position="bottom-end"
                                    >
                                        <Menu.Target>
                                            <ActionIcon
                                                variant="subtle"
                                                color="gray"
                                                aria-label="Действия с движением"
                                                className={classes.actionButton}
                                            >
                                                <DotsThreeVerticalIcon size={24} />
                                            </ActionIcon>
                                        </Menu.Target>

                                        <Menu.Dropdown>
                                            <Menu.Item
                                                leftSection={<PencilIcon size={20} />}
                                                onClick={() => onEdit(movement)}
                                            >
                                                Редактировать
                                            </Menu.Item>
                                            <Menu.Item
                                                color="red"
                                                leftSection={<TrashIcon size={20} />}
                                                onClick={() => onDelete(movement)}
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
    )
}

function OverflowTooltipText({ text }: { text: string }) {
    const textRef = useRef<HTMLParagraphElement | null>(null)
    const [overflowed, setOverflowed] = useState(false)

    useEffect(() => {
        function updateOverflow() {
            const element = textRef.current
            if (!element) {
                setOverflowed(false)
                return
            }

            setOverflowed(element.scrollWidth > element.clientWidth)
        }

        updateOverflow()
        window.addEventListener('resize', updateOverflow)

        return () => {
            window.removeEventListener('resize', updateOverflow)
        }
    }, [text])

    const content = (
        <Text ref={textRef} className={[classes.commentText, classes.movementText].join(' ')}>
            {text}
        </Text>
    )

    if (!overflowed) {
        return content
    }

    return (
        <Tooltip label={text}>
            {content}
        </Tooltip>
    )
}
