import { useEffect, useMemo, useState } from 'react'

import {
    CrownSimpleIcon,
    DotsThreeVerticalIcon,
    TrashIcon,
} from '@phosphor-icons/react'
import {
    ActionIcon,
    Alert,
    Center,
    Drawer,
    FocusTrap,
    Loader,
    Menu,
    ScrollArea,
    Stack,
    Text,
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
import { EmptyState } from '../../../shared/ui/EmptyState'
import {
    ResidentSearchCombobox,
    type ResidentSearchOption,
} from '../../../shared/ui/ResidentSearchCombobox'
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
            })
            .catch((currentError) => {
                if (cancelled) {
                    return
                }

                setSearchError(currentError instanceof Error ? currentError.message : 'Не удалось выполнить поиск')
                setSearchItems([])
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
    const searchOptions: ResidentSearchOption[] = searchItems.map((item) => ({
        id: item.id,
        name: item.name,
        description: item.current_team_id && item.current_team_id !== teamID
            ? `В команде ${formatDutySettingsLeaderName(item.current_team_leader)}`
            : null,
    }))

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
                        <EmptyState
                            title="Участники не найдены"
                            description="В этой команде пока нет участников."
                        />
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

                    <ResidentSearchCombobox
                        searchValue={searchValue}
                        options={searchOptions}
                        loading={searchLoading}
                        error={searchError}
                        placeholder="Добавить участника"
                        selectedId={null}
                        onSearchChange={setSearchValue}
                        onFocus={() => setSearchFocused(true)}
                        onOptionSelect={(option) => {
                            const selectedUser = searchItems.find((item) => item.id === option.id)
                            if (!selectedUser) {
                                return
                            }

                            if (selectedUser.current_team_id && selectedUser.current_team_id !== teamID) {
                                setPendingMove({ user: selectedUser })
                                return
                            }

                            void addMember(selectedUser)
                        }}
                    />
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
