import { useEffect, useMemo, useRef, useState } from 'react'

import {
    CrownSimpleIcon,
    UserMinusIcon,
} from '@phosphor-icons/react'
import {
    ActionIcon,
    Alert,
    Button,
    Center,
    Drawer,
    FocusTrap,
    Group,
    Loader,
    Modal,
    ScrollArea,
    Stack,
    Text,
    Tooltip,
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
import { useOverlayAutofocus } from '../../../shared/ui/useOverlayAutofocus'
import {
    ResidentSearchCombobox,
    type ResidentSearchOption,
} from '../../../shared/ui/ResidentSearchCombobox'
import { SettingsBadge } from '../../../shared/ui/SettingsBadge'
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
    { member: DutySettingsTeamMember }

function TeamMemberCard({
    member,
    onlyLeader,
    assigningLeader,
    onAssignLeader,
    onRemove,
}: {
    member: DutySettingsTeamMember
    onlyLeader: boolean
    assigningLeader: boolean
    onAssignLeader: () => void
    onRemove: () => void
}) {
    const removeDisabled = onlyLeader && member.is_leader
    const removeTooltip = removeDisabled ? 'Нельзя исключить единственного главу' : 'Исключить из команды'

    return (
        <div className={classes.memberCard}>
            <div className={classes.memberContent}>
                <div className={classes.memberDetails}>
                    <Text fw={500} className={classes.memberName}>{member.name}</Text>
                    {member.is_leader ? <SettingsBadge color="leader">Глава команды</SettingsBadge> : null}
                </div>

                <div className={classes.memberActionsCell}>
                    {!member.is_leader ? (
                        <Tooltip label="Назначить главой команды" withArrow>
                            <ActionIcon
                                variant="subtle"
                                color="gray"
                                aria-label={`Назначить ${member.name} главой команды`}
                                className={classes.actionButton}
                                loading={assigningLeader}
                                disabled={assigningLeader}
                                onClick={onAssignLeader}
                            >
                                <CrownSimpleIcon size={20} />
                            </ActionIcon>
                        </Tooltip>
                    ) : null}
                    <Tooltip label={removeTooltip} withArrow>
                        <span>
                            <ActionIcon
                                variant="subtle"
                                color="gray"
                                aria-label={`Исключить ${member.name} из команды`}
                                className={classes.actionButton}
                                disabled={removeDisabled}
                                onClick={onRemove}
                            >
                                <UserMinusIcon size={20} />
                            </ActionIcon>
                        </span>
                    </Tooltip>
                </div>
            </div>
        </div>
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
    const [searchComboboxKey, setSearchComboboxKey] = useState(0)
    const [pendingMove, setPendingMove] = useState<PendingMove | null>(null)
    const [pendingAction, setPendingAction] = useState<PendingMemberAction | null>(null)
    const [replacementLeaderId, setReplacementLeaderId] = useState<string | null>(null)
    const [replacementSearch, setReplacementSearch] = useState('')
    const [removingMember, setRemovingMember] = useState(false)
    const [removeError, setRemoveError] = useState<string | null>(null)
    const [assigningLeaderId, setAssigningLeaderId] = useState<string | null>(null)
    const teamID = team?.id ?? null
    const contentRef = useRef<HTMLDivElement | null>(null)

    const title = 'Исполнители команды'

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
            setReplacementLeaderId(null)
            setReplacementSearch('')
            setRemoveError(null)
            return
        }

        if (!teamID) {
            setSearchItems([])
            return
        }

        if (!searchFocused) {
            setSearchItems([])
            setSearchError(null)
            setSearchLoading(false)
            return
        }

        let cancelled = false
        setSearchLoading(true)
        setSearchError(null)

        void searchDutySettingsTeamMembers(groupId, teamID, debouncedSearch.trim())
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

    function resetMemberSearch() {
        setSearchValue('')
        setSearchItems([])
        setSearchError(null)
        setSearchLoading(false)
        setSearchFocused(false)
        setSearchComboboxKey((current) => current + 1)
    }

    function closeRemovalModal() {
        setPendingAction(null)
        setReplacementLeaderId(null)
        setReplacementSearch('')
        setRemoveError(null)
    }

    async function addMember(user: DutySettingsTeamSearchItem) {
        if (!teamID) {
            return
        }

        setError(null)

        try {
            await addDutySettingsTeamMember(groupId, teamID, user.id)
            resetMemberSearch()
            await refreshAfterMutation()
        } catch (currentError) {
            if (currentError instanceof ApiError) {
                setError(currentError.message)
            } else {
                setError('Не удалось добавить участника в команду')
            }
        }
    }

    const members = useMemo(() => data?.members ?? [], [data])
    const pendingMemberIsLeader = pendingAction?.member.is_leader ?? false
    const replacementOptions = useMemo(() => {
        const query = replacementSearch.trim().toLocaleLowerCase()
        return members
            .filter((member) => member.id !== pendingAction?.member.id)
            .filter((member) => query === '' || member.name.toLocaleLowerCase().includes(query))
            .map((member) => ({ id: member.id, name: member.name }))
    }, [members, pendingAction?.member.id, replacementSearch])
    const searchOptions: ResidentSearchOption[] = searchItems.map((item) => ({
        id: item.id,
        name: item.name,
        description: item.current_team_id && item.current_team_id !== teamID
            ? `В команде ${formatDutySettingsLeaderName(item.current_team_leader)}`
            : null,
    }))

    useOverlayAutofocus(contentRef, { enabled: opened })

    return (
        <>
            <Drawer opened={opened} onClose={onClose} position="right" size={860} title={title}>
                <Stack ref={contentRef} gap="md">
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
                                        onlyLeader={members.length === 1}
                                        assigningLeader={assigningLeaderId === member.id}
                                        onAssignLeader={() => {
                                            if (!teamID) {
                                                return
                                            }

                                            setAssigningLeaderId(member.id)
                                            void assignDutySettingsTeamLeader(groupId, teamID, member.id)
                                                .then(refreshAfterMutation)
                                                .catch((currentError) => {
                                                    if (currentError instanceof ApiError) {
                                                        setError(currentError.message)
                                                    } else {
                                                        setError('Не удалось назначить главу команды')
                                                    }
                                                })
                                                .finally(() => setAssigningLeaderId(null))
                                        }}
                                        onRemove={() => {
                                            setPendingAction({ member })
                                            setReplacementLeaderId(null)
                                            setReplacementSearch('')
                                            setRemoveError(null)
                                        }}
                                    />
                                ))}
                            </Stack>
                        </ScrollArea>
                    )}

                    <ResidentSearchCombobox
                        key={searchComboboxKey}
                        overlayOpened={opened}
                        searchValue={searchValue}
                        options={searchOptions}
                        loading={searchLoading}
                        error={searchError}
                        placeholder="Добавить участника"
                        selectedId={null}
                        onSearchChange={setSearchValue}
                        onFocus={() => {
                            setSearchFocused(true)
                            if (searchItems.length === 0) {
                                setSearchLoading(true)
                            }
                        }}
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

            <Modal
                opened={pendingAction !== null}
                onClose={closeRemovalModal}
                title="Исключение исполнителя"
                withCloseButton={false}
                centered
                radius="xl"
                size={620}
            >
                <FocusTrap.InitialFocus />
                <Stack gap="lg">
                    {removeError ? <Alert color="red">{removeError}</Alert> : null}
                    {pendingAction ? (
                        <Text>
                            Вы уверены, что хотите исключить исполнителя <Text component="span" fw={700}>{pendingAction.member.name}</Text> из команды?
                        </Text>
                    ) : null}
                    {pendingAction && pendingMemberIsLeader ? (
                        <>
                            <Text>
                                <Text component="span" fw={700}>{pendingAction.member.name}</Text> является главой команды, назначьте нового главу команды:
                            </Text>
                            <ResidentSearchCombobox
                                overlayOpened={pendingAction !== null}
                                searchValue={replacementSearch}
                                options={replacementOptions}
                                loading={false}
                                placeholder="Житель*"
                                selectedId={replacementLeaderId}
                                selectedLabel={replacementOptions.find((option) => option.id === replacementLeaderId)?.name ?? null}
                                hideDropdownWhenSelected
                                onSearchChange={(value) => {
                                    setReplacementSearch(value)
                                    setReplacementLeaderId(null)
                                }}
                                onOptionSelect={(option) => {
                                    setReplacementLeaderId(option.id)
                                    setReplacementSearch(option.name)
                                }}
                            />
                        </>
                    ) : null}
                    <Group justify="flex-end" gap="15">
                        <Button variant="default" onClick={closeRemovalModal}>Отменить</Button>
                        <Button
                            color="dark"
                            loading={removingMember}
                            disabled={pendingMemberIsLeader && replacementLeaderId === null}
                            onClick={async () => {
                                if (!teamID || !pendingAction) {
                                    return
                                }
                                setRemovingMember(true)
                                setRemoveError(null)
                                try {
                                    await removeDutySettingsTeamMember(groupId, teamID, pendingAction.member.id, replacementLeaderId)
                                    closeRemovalModal()
                                    resetMemberSearch()
                                    await refreshAfterMutation()
                                } catch (currentError) {
                                    setRemoveError(currentError instanceof ApiError ? currentError.message : 'Не удалось исключить участника из команды')
                                } finally {
                                    setRemovingMember(false)
                                }
                            }}
                        >
                            Исключить
                        </Button>
                    </Group>
                </Stack>
            </Modal>
        </>
    )
}
