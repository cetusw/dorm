import { useEffect, useMemo, useState } from 'react'

import {
    CrownSimpleIcon,
    DotsThreeVerticalIcon,
    TrashIcon,
    UserIcon,
} from '@phosphor-icons/react'
import {
    ActionIcon,
    Alert,
    Center,
    Combobox,
    Drawer,
    FocusTrap,
    InputBase,
    Loader,
    Menu,
    ScrollArea,
    Stack,
    Text,
    useCombobox,
} from '@mantine/core'
import { useDebouncedValue } from '@mantine/hooks'

import {
    addDutySettingsTeamMember,
    assignDutySettingsTeamLeader,
    getDutySettingsTeamMembers,
    removeDutySettingsTeamMember,
    searchDutySettingsTeamMembers,
} from '../../../features/duty-settings/api/dutySettingsApi'
import type {
    DutySettingsTeam,
    DutySettingsTeamMember,
    DutySettingsTeamMembersResponse,
    DutySettingsTeamSearchItem,
} from '../../../features/duty-settings/model/types'
import { formatDutySettingsLeaderName } from '../../../features/duty-settings/model/utils'
import { ApiError } from '../../../shared/api/ApiError'
import { ConfirmActionModal } from '../../../shared/ui/ConfirmActionModal'
import { SettingsBadge } from '../../../shared/ui/SettingsBadge'
import { SettingsCardSurface } from '../../../shared/ui/SettingsCardSurface'
import classes from './DutySettingsTeamMembersDrawer.module.css'

type Props = {
    groupId: string
    team: DutySettingsTeam | null
    opened: boolean
    onClose: () => void
    onUpdated: () => Promise<void>
}

type PendingMove = {
    user: DutySettingsTeamSearchItem
}

type PendingMemberAction =
    { type: 'remove-member'; member: DutySettingsTeamMember }

function TeamMemberCard({
    member,
    onAssignLeader,
    onRemove,
}: {
    member: DutySettingsTeamMember
    onAssignLeader: () => void
    onRemove: () => void
}) {
    const [menuOpened, setMenuOpened] = useState(false)

    return (
        <SettingsCardSurface className={classes.memberCard} menuOpen={menuOpened}>
            <div className={classes.memberContent}>
                <div className={classes.memberNameCell}>
                    <Text fw={500} className={classes.memberName}>{member.name}</Text>
                </div>

                <div className={classes.memberBadgeCell}>
                    {member.is_leader ? <SettingsBadge color="leader">Глава команды</SettingsBadge> : null}
                </div>

                <div className={classes.memberActionsCell}>
                    <Menu opened={menuOpened} onChange={setMenuOpened} withinPortal position="bottom-end">
                        <Menu.Target>
                            <ActionIcon
                                variant="subtle"
                                color="gray"
                                aria-label="Действия с участником"
                                className={classes.actionButton}
                            >
                                <DotsThreeVerticalIcon size={25} />
                            </ActionIcon>
                        </Menu.Target>
                        <Menu.Dropdown>
                            <Menu.Item leftSection={<CrownSimpleIcon size={25} />} onClick={onAssignLeader}>
                                Назначить главой
                            </Menu.Item>
                            <Menu.Item color="red" leftSection={<TrashIcon size={25} />} onClick={onRemove}>
                                Исключить из команды
                            </Menu.Item>
                        </Menu.Dropdown>
                    </Menu>
                </div>
            </div>
        </SettingsCardSurface>
    )
}

export function DutySettingsTeamMembersDrawer({ groupId, team, opened, onClose, onUpdated }: Props) {
    const [data, setData] = useState<DutySettingsTeamMembersResponse | null>(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [searchValue, setSearchValue] = useState('')
    const [debouncedSearch] = useDebouncedValue(searchValue, 300)
    const [searchLoading, setSearchLoading] = useState(false)
    const [searchError, setSearchError] = useState<string | null>(null)
    const [searchItems, setSearchItems] = useState<DutySettingsTeamSearchItem[]>([])
    const [searchFocused, setSearchFocused] = useState(false)
    const [pendingMove, setPendingMove] = useState<PendingMove | null>(null)
    const [pendingAction, setPendingAction] = useState<PendingMemberAction | null>(null)
    const combobox = useCombobox()
    const teamID = team?.id ?? null

    const title = useMemo(() => {
        if (!team) {
            return 'Участники команды'
        }
        return `Участники команды ${formatDutySettingsLeaderName(team.leader)}`
    }, [team])

    async function loadMembers() {
        if (!team) {
            setData(null)
            return
        }

        setLoading(true)
        setError(null)

        try {
            const response = await getDutySettingsTeamMembers(groupId, team.id)
            setData(response)
        } catch (currentError) {
            setError(currentError instanceof Error ? currentError.message : 'Не удалось загрузить участников команды')
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        if (!opened || !teamID) {
            return
        }

        void loadMembers()
    }, [groupId, opened, teamID])

    useEffect(() => {
        if (!opened) {
            setSearchValue('')
            setSearchItems([])
            setSearchError(null)
            setSearchFocused(false)
            setPendingMove(null)
            setPendingAction(null)
            combobox.closeDropdown()
            return
        }

        if (!teamID) {
            setSearchItems([])
            return
        }

        const trimmedQuery = debouncedSearch.trim()
        if (trimmedQuery === '') {
            setSearchItems([])
            setSearchError(null)
            setSearchLoading(false)
            combobox.closeDropdown()
            return
        }

        let cancelled = false
        setSearchLoading(true)
        setSearchError(null)

        void searchDutySettingsTeamMembers(groupId, teamID, trimmedQuery)
            .then((response) => {
                if (cancelled) {
                    return
                }

                setSearchItems(response.users)
                if (searchFocused) {
                    combobox.openDropdown()
                }
            })
            .catch((currentError) => {
                if (cancelled) {
                    return
                }

                setSearchError(currentError instanceof Error ? currentError.message : 'Не удалось выполнить поиск')
                setSearchItems([])
                if (searchFocused) {
                    combobox.openDropdown()
                }
            })
            .finally(() => {
                if (!cancelled) {
                    setSearchLoading(false)
                }
            })

        return () => {
            cancelled = true
        }
    }, [debouncedSearch, groupId, opened, searchFocused, teamID])

    async function refreshAfterMutation() {
        await loadMembers()
        await onUpdated()
    }

    async function addMember(user: DutySettingsTeamSearchItem) {
        if (!teamID) {
            return
        }

        setError(null)

        try {
            await addDutySettingsTeamMember(groupId, teamID, user.id)
            setSearchValue('')
            setSearchItems([])
            combobox.closeDropdown()
            await refreshAfterMutation()
        } catch (currentError) {
            if (currentError instanceof ApiError) {
                setError(currentError.message)
            } else {
                setError('Не удалось добавить участника в команду')
            }
        }
    }

    const members = data?.members ?? []

    return (
        <>
            <Drawer opened={opened} onClose={onClose} position="right" size={860} title={title}>
                <FocusTrap.InitialFocus />

                <Stack gap="md">
                    {error ? <Alert color="red">{error}</Alert> : null}

                    {loading ? (
                        <Center pt="md">
                            <Loader size={32} />
                        </Center>
                    ) : members.length === 0 ? (
                        <Alert color="gray">В этой команде пока нет участников.</Alert>
                    ) : (
                        <ScrollArea>
                            <Stack gap={6} pb={4}>
                                {members.map((member) => (
                                    <TeamMemberCard
                                        key={member.id}
                                        member={member}
                                        onAssignLeader={() => {
                                            if (!teamID) {
                                                return
                                            }

                                            void assignDutySettingsTeamLeader(groupId, teamID, member.id)
                                                .then(refreshAfterMutation)
                                                .catch((currentError) => {
                                                    if (currentError instanceof ApiError) {
                                                        setError(currentError.message)
                                                    } else {
                                                        setError('Не удалось назначить главу команды')
                                                    }
                                                })
                                        }}
                                        onRemove={() => setPendingAction({ type: 'remove-member', member })}
                                    />
                                ))}
                            </Stack>
                        </ScrollArea>
                    )}

                    <Combobox
                        store={combobox}
                        onOptionSubmit={(value) => {
                            const selectedUser = searchItems.find((item) => item.id === value)
                            if (!selectedUser) {
                                return
                            }

                            if (selectedUser.current_team_id && selectedUser.current_team_id !== teamID) {
                                setPendingMove({ user: selectedUser })
                                combobox.closeDropdown()
                                return
                            }

                            void addMember(selectedUser)
                        }}
                        withinPortal={false}
                    >
                        <Combobox.Target>
                            <InputBase
                                value={searchValue}
                                onChange={(event) => {
                                    const nextValue = event.currentTarget.value
                                    setSearchValue(nextValue)
                                    if (nextValue.trim() !== '') {
                                        combobox.openDropdown()
                                    }
                                }}
                                onFocus={() => {
                                    setSearchFocused(true)
                                    if (searchItems.length > 0) {
                                        combobox.openDropdown()
                                    }
                                }}
                                onBlur={() => {
                                    setSearchFocused(false)
                                    combobox.closeDropdown()
                                }}
                                placeholder="Добавить участника"
                                leftSection={searchLoading ? <Loader size={20} /> : <UserIcon size={20} />}
                                classNames={{
                                    input: classes.searchInput,
                                    section: classes.searchSection,
                                }}
                            />
                        </Combobox.Target>

                        <Combobox.Dropdown hidden={searchValue.trim() === '' && !searchLoading}>
                            <Combobox.Options mah={280} style={{ overflowY: 'auto' }}>
                                {searchItems.length === 0 ? (
                                    <Combobox.Empty>
                                        {searchLoading ? 'Поиск...' : 'Ничего не найдено'}
                                    </Combobox.Empty>
                                ) : (
                                    searchItems.map((item) => (
                                        <Combobox.Option value={item.id} key={item.id}>
                                            <Stack gap={2}>
                                                <Text fw={500}>{item.name}</Text>
                                                {item.current_team_id && item.current_team_id !== teamID ? (
                                                    <Text size="sm" className={classes.searchHint}>
                                                        В команде {formatDutySettingsLeaderName(item.current_team_leader)}
                                                    </Text>
                                                ) : null}
                                            </Stack>
                                        </Combobox.Option>
                                    ))
                                )}
                            </Combobox.Options>
                        </Combobox.Dropdown>
                    </Combobox>

                    {searchError ? <Alert color="red">{searchError}</Alert> : null}
                </Stack>
            </Drawer>

            <ConfirmActionModal
                opened={pendingMove !== null}
                onClose={() => setPendingMove(null)}
                title="Перемещение участника"
                description={pendingMove ? `Переместить жителя «${pendingMove.user.name}» в эту команду?` : ''}
                confirmLabel="Переместить"
                onConfirm={async () => {
                    if (!pendingMove) {
                        return
                    }
                    await addMember(pendingMove.user)
                    setPendingMove(null)
                }}
                errorMessage="Не удалось переместить участника"
            />

            <ConfirmActionModal
                opened={pendingAction !== null}
                onClose={() => setPendingAction(null)}
                title="Исключение из команды"
                description={pendingAction == null ? '' : `Исключить жителя «${pendingAction.member.name}» из команды?`}
                confirmLabel="Исключить"
                confirmColor="red"
                onConfirm={async () => {
                    if (!teamID || !pendingAction) {
                        return
                    }

                    await removeDutySettingsTeamMember(groupId, teamID, pendingAction.member.id)
                    setPendingAction(null)
                    await refreshAfterMutation()
                }}
                errorMessage="Не удалось исключить участника из команды"
            />
        </>
    )
}
