import { useMemo, useState } from 'react'

import { DotsThreeVerticalIcon, MagnifyingGlassIcon, PencilIcon, TrashIcon } from '@phosphor-icons/react'
import { ActionIcon, Menu, Table, Text, TextInput, UnstyledButton } from '@mantine/core'

import type { TeamListItem } from '../model/types'
import { ListTable, listTableClasses } from '../../../shared/ui/ListTable'
import classes from './TeamsTable.module.css'

type SortField = 'name' | 'leaderName'
type SortDirection = 'asc' | 'desc'

type Props = {
    teams: TeamListItem[]
    onEdit: (teamId: string) => void
    onDelete: (team: TeamListItem) => void
}

function normalize(value: string): string {
    return value.trim().toLowerCase()
}

export function TeamsTable({ teams, onEdit, onDelete }: Props) {
    const [search, setSearch] = useState('')
    const [sortField, setSortField] = useState<SortField>('name')
    const [sortDirection, setSortDirection] = useState<SortDirection>('asc')
    const [openedMenuId, setOpenedMenuId] = useState<string | null>(null)

    const visibleTeams = useMemo(() => {
        const query = normalize(search)
        const filteredTeams = query === ''
            ? teams
            : teams.filter((team) =>
                normalize(team.leader.name).includes(query),
            )

        return [...filteredTeams].sort((left, right) => {
            const leftValue = normalize(left.leader.name)
            const rightValue = normalize(right.leader.name)
            const result = leftValue.localeCompare(rightValue, 'ru')

            return sortDirection === 'asc' ? result : result * -1
        })
    }, [search, sortDirection, sortField, teams])

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
                    {visibleTeams.map((team) => {
                        const menuOpened = openedMenuId === team.id

                        return (
                            <Table.Tr
                                key={team.id}
                                className={`${listTableClasses.bodyRow} ${classes.row}`}
                                data-menu-open={menuOpened ? 'true' : undefined}
                            >
                                <Table.Td>
                                    <Text fw={600}>{team.leader.name}</Text>
                                </Table.Td>
                                <Table.Td>
                                    <Text>{team.leader.name}</Text>
                                </Table.Td>
                                <Table.Td className={classes.actionCell}>
                                    <div className={classes.actionCellInner}>
                                        <Menu
                                            opened={menuOpened}
                                            onChange={(opened) => setOpenedMenuId(opened ? team.id : null)}
                                            withinPortal
                                            position="bottom-end"
                                        >
                                            <Menu.Target>
                                                <ActionIcon
                                                    variant="subtle"
                                                    color="gray"
                                                    aria-label={`Действия с командой ${team.leader.name}`}
                                                    className={classes.actionButton}
                                                >
                                                    <DotsThreeVerticalIcon size={24} />
                                                </ActionIcon>
                                            </Menu.Target>
                                            <Menu.Dropdown>
                                                <Menu.Item
                                                    leftSection={<PencilIcon size={20} />}
                                                    onClick={() => onEdit(team.id)}
                                                >
                                                    Редактировать
                                                </Menu.Item>
                                                <Menu.Item
                                                    color="red"
                                                    leftSection={<TrashIcon size={20} />}
                                                    onClick={() => onDelete(team)}
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
