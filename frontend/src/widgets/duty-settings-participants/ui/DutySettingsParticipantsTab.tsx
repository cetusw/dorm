import { useEffect, useMemo, useState } from 'react'

import { CrownSimpleIcon, HandCoinsIcon, UserMinusIcon, UsersIcon, UserSwitchIcon } from '@phosphor-icons/react'
import { ActionIcon, Alert, Button, Center, FocusTrap, Group, Loader, Menu, Modal, Stack, Text, Tooltip } from '@mantine/core'
import { useDebouncedValue } from '@mantine/hooks'

import {
    addDutyParticipant,
    changeDutyLeader,
    excludeDutyParticipant,
    getDutyParticipantCandidates,
    restoreDutyParticipant,
    assignDutySettingsActiveTeam,
} from '../../../features/duty-settings/api/dutySettingsApi'
import type { DutyParticipant, DutyParticipantCandidate, DutySettingsActiveDuty, DutySettingsTeam } from '../../../features/duty-settings/model/types'
import { ApiError } from '../../../shared/api/ApiError'
import { ConfirmActionModal } from '../../../shared/ui/ConfirmActionModal'
import { EmptyState } from '../../../shared/ui/EmptyState'
import { ResidentSearchCombobox, type ResidentSearchOption } from '../../../shared/ui/ResidentSearchCombobox'
import { SettingsBadge } from '../../../shared/ui/SettingsBadge'
import classes from './DutySettingsParticipantsTab.module.css'

type Props = {
    groupId: string
    activeDuty: DutySettingsActiveDuty
    teams: DutySettingsTeam[]
    activeDutyTeamId: string | null
    participants: DutyParticipant[]
    loading: boolean
    participantsError: string | null
    onParticipantsChanged: () => Promise<void>
}

type PendingExclusion = { participant: DutyParticipant }

function formatPeriod(startDate: string, endDate: string): string {
    const format = (value: string) => {
        const parts = value.split('-')
        return parts.length === 3 ? `${parts[2]}.${parts[1]}` : value
    }
    return `${format(startDate)} - ${format(endDate)}`
}

function ParticipantCard({
    participant,
    onlyLeader,
    changingLeader,
    onChangeLeader,
    onExclude,
}: {
    participant: DutyParticipant
    onlyLeader: boolean
    changingLeader: boolean
    onChangeLeader: (participantId: string) => void
    onExclude: (participant: DutyParticipant) => void
}) {
    const exclusionDisabled = participant.is_leader && onlyLeader
    const exclusionLabel = exclusionDisabled ? 'Нельзя исключить единственного главу' : 'Исключить из дежурства'

    return (
        <div className={`${classes.participantCard} ${participant.is_leader ? classes.leaderCard : ''}`}>
            <div className={classes.participantContent}>
                <div className={classes.participantDetails}>
                    <Text fw={500} className={classes.participantName}>{participant.full_name}</Text>
                    {participant.is_leader ? <SettingsBadge color="leader">Глава дежурства</SettingsBadge> : null}
                </div>
                <div className={classes.participantActions}>
                    {!participant.is_leader ? (
                        <Tooltip label="Назначить главой дежурства" withArrow>
                            <ActionIcon
                                variant="subtle"
                                color="gray"
                                className={classes.actionButton}
                                aria-label={`Назначить ${participant.full_name} главой дежурства`}
                                loading={changingLeader}
                                disabled={changingLeader}
                                onClick={() => onChangeLeader(participant.participant_id)}
                            >
                                <CrownSimpleIcon size={20} />
                            </ActionIcon>
                        </Tooltip>
                    ) : null}
                    <Tooltip label={exclusionLabel} withArrow>
                        <span>
                            <ActionIcon
                                variant="subtle"
                                color="gray"
                                className={classes.actionButton}
                                aria-label={`Исключить ${participant.full_name} из дежурства`}
                                disabled={exclusionDisabled}
                                onClick={() => onExclude(participant)}
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

export function DutySettingsParticipantsTab({
    groupId,
    activeDuty,
    teams,
    activeDutyTeamId,
    participants,
    loading,
    participantsError,
    onParticipantsChanged,
}: Props) {
    const [error, setError] = useState<string | null>(null)
    const [searchValue, setSearchValue] = useState('')
    const [searchFocused, setSearchFocused] = useState(false)
    const [candidateLoading, setCandidateLoading] = useState(false)
    const [candidates, setCandidates] = useState<DutyParticipantCandidate[]>([])
    const [pendingExclusion, setPendingExclusion] = useState<PendingExclusion | null>(null)
    const [replacementLeaderID, setReplacementLeaderID] = useState<string | null>(null)
    const [replacementSearch, setReplacementSearch] = useState('')
    const [mutationError, setMutationError] = useState<string | null>(null)
    const [excluding, setExcluding] = useState(false)
    const [changingLeaderID, setChangingLeaderID] = useState<string | null>(null)
    const [adding, setAdding] = useState(false)
    const [pendingDutyTeam, setPendingDutyTeam] = useState<DutySettingsTeam | null>(null)
    const [debouncedSearch] = useDebouncedValue(searchValue, 300)

    useEffect(() => {
        if (!searchFocused) {
            return
        }
        let cancelled = false
        setCandidateLoading(true)
        void getDutyParticipantCandidates(activeDuty.id)
            .then((response) => { if (!cancelled) setCandidates(response) })
            .catch((currentError) => { if (!cancelled) setError(currentError instanceof Error ? currentError.message : 'Не удалось загрузить жителей') })
            .finally(() => { if (!cancelled) setCandidateLoading(false) })
        return () => { cancelled = true }
    }, [activeDuty.id, searchFocused])

    const activeParticipants = useMemo(() => participants
        .filter((participant) => participant.excluded_at === null)
        .sort((left, right) => {
            if (left.is_leader !== right.is_leader) return left.is_leader ? -1 : 1
            return left.full_name.localeCompare(right.full_name, 'ru')
        }), [participants])
    const activeParticipantIDs = useMemo(() => new Set(activeParticipants.map((participant) => participant.participant_id)), [activeParticipants])
    const searchOptions: ResidentSearchOption[] = useMemo(() => {
        const query = debouncedSearch.trim().toLocaleLowerCase()
        return candidates
            .filter((candidate) => !activeParticipantIDs.has(candidate.participant_id))
            .filter((candidate) => query === '' || candidate.full_name.toLocaleLowerCase().includes(query))
            .map((candidate) => ({
                id: candidate.participant_id,
                name: candidate.full_name,
                description: candidate.team_name ? `В команде ${candidate.team_name}` : null,
            }))
    }, [activeParticipantIDs, candidates, debouncedSearch])
    const replacementOptions: ResidentSearchOption[] = useMemo(() => {
        const query = replacementSearch.trim().toLocaleLowerCase()
        return activeParticipants
            .filter((participant) => participant.participant_id !== pendingExclusion?.participant.participant_id)
            .filter((participant) => query === '' || participant.full_name.toLocaleLowerCase().includes(query))
            .map((participant) => ({ id: participant.participant_id, name: participant.full_name }))
    }, [activeParticipants, pendingExclusion?.participant.participant_id, replacementSearch])

    async function refreshAfterMutation() {
        await onParticipantsChanged()
    }

    async function addOrRestore(option: ResidentSearchOption) {
        if (adding) return
        const existing = participants.find((participant) => participant.participant_id === option.id)
        setAdding(true)
        setError(null)
        setSearchFocused(false)
        setSearchValue('')
        try {
            if (existing && existing.excluded_at !== null) await restoreDutyParticipant(activeDuty.id, option.id)
            else await addDutyParticipant(activeDuty.id, option.id)
            await refreshAfterMutation()
        } catch (currentError) {
            setError(currentError instanceof ApiError ? currentError.message : 'Не удалось добавить исполнителя')
            void onParticipantsChanged()
        } finally {
            setAdding(false)
        }
    }

    function closeExclusion() {
        setPendingExclusion(null)
        setReplacementLeaderID(null)
        setReplacementSearch('')
        setMutationError(null)
    }

    const pointsPerParticipant = activeParticipants.length === 0
        ? 0
        : Math.floor(activeDuty.summary.total_cost / activeParticipants.length)
    const nextDutyTeams = teams.filter((team) => team.id !== (activeDutyTeamId ?? activeDuty.team_id))

    return (
        <>
            <div className={classes.summary}>
                <div className={classes.summaryItem}><UsersIcon size={20} />{activeParticipants.length} исполнителей</div>
                <div className={classes.summaryItem}><HandCoinsIcon size={20} />{pointsPerParticipant} баллов на исполнителя</div>
            </div>
            <div className={classes.search}>
                <div className={classes.searchRow}>
                    <div className={classes.searchField}>
                        <ResidentSearchCombobox
                            overlayOpened={searchFocused}
                            searchValue={searchValue}
                            options={searchOptions}
                            loading={candidateLoading || adding}
                            placeholder="Добавить исполнителя"
                            selectedId={null}
                            inputClassName={classes.addParticipantInput}
                            onSearchChange={setSearchValue}
                            onFocus={() => {
                                setSearchFocused(true)
                                if (candidates.length === 0) setCandidateLoading(true)
                            }}
                            onOptionSelect={(option) => { void addOrRestore(option) }}
                        />
                    </div>
                    {teams.length > 0 ? (
                        <Menu position="bottom-end" shadow="sm" width="target">
                            <Menu.Target>
                                <Button
                                    variant="default"
                                    className={classes.switchTeamButton}
                                    leftSection={<UserSwitchIcon size={20} />}
                                    disabled={nextDutyTeams.length === 0}
                                >
                                    Сменить команду
                                </Button>
                            </Menu.Target>
                            <Menu.Dropdown>
                                {nextDutyTeams.map((team) => (
                                    <Menu.Item key={team.id} onClick={() => setPendingDutyTeam(team)}>
                                        {team.leader.name}
                                    </Menu.Item>
                                ))}
                            </Menu.Dropdown>
                        </Menu>
                    ) : null}
                </div>
            </div>
            {participantsError ?? error ? <Alert mt="sm" color="red">{participantsError ?? error}</Alert> : null}
            {loading ? <Center py="xl"><Loader /></Center> : activeParticipants.length === 0 ? (
                <EmptyState title="Исполнители не найдены" description="Добавьте исполнителя в это дежурство." />
            ) : (
                <div className={classes.participantList}>
                    {activeParticipants.map((participant) => (
                        <ParticipantCard
                            key={participant.participant_id}
                            participant={participant}
                            onlyLeader={activeParticipants.length === 1}
                            changingLeader={changingLeaderID === participant.participant_id}
                            onChangeLeader={(participantId) => {
                                setChangingLeaderID(participantId)
                                void changeDutyLeader(activeDuty.id, participantId)
                                    .then(refreshAfterMutation)
                                    .catch((currentError) => {
                                        setError(currentError instanceof ApiError ? currentError.message : 'Не удалось назначить главу дежурства')
                                        void onParticipantsChanged()
                                    })
                                    .finally(() => setChangingLeaderID(null))
                            }}
                            onExclude={(participant) => {
                                setPendingExclusion({ participant })
                                setReplacementLeaderID(null)
                                setReplacementSearch('')
                                setMutationError(null)
                            }}
                        />
                    ))}
                </div>
            )}
            <Modal opened={pendingExclusion !== null} onClose={closeExclusion} title="Исключение исполнителя" withCloseButton={false} centered radius="xl" size={700}>
                <FocusTrap.InitialFocus />
                <Stack gap="lg">
                    {mutationError ? <Alert color="red">{mutationError}</Alert> : null}
                    {pendingExclusion ? <Text>Вы уверены, что хотите исключить исполнителя <Text component="span" fw={700}>{pendingExclusion.participant.full_name}</Text> из дежурства <Text component="span" fw={700}>{formatPeriod(activeDuty.start_date, activeDuty.end_date)}</Text>?</Text> : null}
                    {pendingExclusion?.participant.is_leader ? (
                        <>
                            <Text><Text component="span" fw={700}>{pendingExclusion.participant.full_name}</Text> является главой дежурства, назначьте нового главу дежурства:</Text>
                            <ResidentSearchCombobox
                                overlayOpened={pendingExclusion !== null}
                                searchValue={replacementSearch}
                                options={replacementOptions}
                                loading={false}
                                placeholder="Житель*"
                                selectedId={replacementLeaderID}
                                selectedLabel={replacementOptions.find((option) => option.id === replacementLeaderID)?.name ?? null}
                                hideDropdownWhenSelected
                                onSearchChange={(value) => { setReplacementSearch(value); setReplacementLeaderID(null) }}
                                onOptionSelect={(option) => { setReplacementLeaderID(option.id); setReplacementSearch(option.name) }}
                            />
                        </>
                    ) : null}
                    <Group justify="flex-end" gap="15">
                        <Button variant="default" onClick={closeExclusion}>Отменить</Button>
                        <Button color="dark" loading={excluding} disabled={pendingExclusion?.participant.is_leader && replacementLeaderID === null} onClick={async () => {
                            if (!pendingExclusion) return
                            setExcluding(true)
                            setMutationError(null)
                            try {
                                await excludeDutyParticipant(activeDuty.id, pendingExclusion.participant.participant_id, replacementLeaderID)
                                closeExclusion()
                                await refreshAfterMutation()
                            } catch (currentError) {
                                setMutationError(currentError instanceof ApiError ? currentError.message : 'Не удалось исключить исполнителя')
                                void onParticipantsChanged()
                            } finally { setExcluding(false) }
                        }}>Исключить</Button>
                    </Group>
                </Stack>
            </Modal>
            <ConfirmActionModal
                opened={pendingDutyTeam !== null}
                title="Назначить дежурной"
                description={`Вы уверены, что хотите назначить команду ${pendingDutyTeam?.leader.name ?? ''} дежурной? Прогресс по задачам текущей дежурной команды будет утерян.`}
                confirmLabel="Назначить"
                onClose={() => setPendingDutyTeam(null)}
                onConfirm={async () => {
                    if (!pendingDutyTeam) return
                    await assignDutySettingsActiveTeam(groupId, pendingDutyTeam.id)
                    setPendingDutyTeam(null)
                    await onParticipantsChanged()
                }}
                errorMessage="Не удалось назначить дежурную команду"
            />
        </>
    )
}
