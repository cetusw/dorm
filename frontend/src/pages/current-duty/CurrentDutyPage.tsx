import { useEffect, useMemo, useState } from 'react'

import { BookOpenIcon, CalendarXIcon, DotsThreeOutlineIcon, GearIcon, PlusIcon } from '@phosphor-icons/react'
import { ActionIcon, Alert, Box, Center, Group, Loader, Menu, Popover, SegmentedControl, Select, Stack, Text } from '@mantine/core'
import { useMediaQuery } from '@mantine/hooks'

import type { CurrentUser } from '../../features/current-user/model/types'
import type { DutyTaskSelect } from '../../features/current-duty/model/types'
import {
    calculateDutyAnalytics,
    selectVisibleDutyTabs,
    selectDutyViewOptions,
    selectTasksForActiveSelect,
} from '../../features/current-duty/model/selectors'
import {
    isSupportedMobileDutySelect,
    selectTasksForMobileMine,
} from '../../features/current-duty/model/mobileSelect'
import { CreateGroupDutyModal } from '../../features/current-duty/ui/CreateGroupDutyModal'
import { useCurrentDuty } from '../../features/current-duty/model/useCurrentDuty'
import { useStoredDutySelect } from '../../features/current-duty/model/useStoredDutySelect'
import { formatDutyPeriod } from '../../features/current-duty/model/utils'
import { CurrentDutyAnalytics } from '../../features/current-duty/ui/CurrentDutyAnalytics'
import { MobileDutyProgress } from '../../features/current-duty/ui/MobileDutyProgress'
import { MobileDutyBottomNav } from '../../features/current-duty/ui/MobileDutyBottomNav'
import { DutyTaskSelects } from '../../features/current-duty/ui/DutyTaskSelects'
import { TeamMemberTaskGroups } from '../../features/current-duty/ui/TeamMemberTaskGroups'
import { TaskGroups } from '../../features/current-duty/ui/TaskGroups'
import { BuildingPlanPanel } from '../../features/current-duty/building-plan/BuildingPlanPanel'
import { floorPlans } from '../../features/current-duty/building-plan/generated/plans'
import { getFloorLabel } from '../../features/current-duty/building-plan/utils'
import { PageFrame } from '../../shared/ui/PageFrame'
import { EmptyState } from '../../shared/ui/EmptyState'
import { FloatingNotification } from '../../shared/ui/FloatingNotification'
import { useSelectedDormitoryId } from '../../features/dormitories/model/useDormitorySelection'
import segmentedControlClasses from '../../features/current-duty/ui/SegmentedControl.module.css'
import { navigateTo } from '../../app/navigation'
import { useMobileHeaderContent } from '../../app/MobileHeaderContentContext'

type PlanLegendItem = {
    color: string
    label: string
    stroke: string
}

function getPlanLegendItems(activeSelect: DutyTaskSelect): PlanLegendItem[] {
    switch (activeSelect) {
        case 'mine':
            return [
                { label: 'Нет задач', color: '#F0FDFA', stroke: '#99F6E4' },
                { label: 'Ваши задачи', color: '#E0F2FE', stroke: '#0369A1' },
                { label: 'Все задачи выполнены', color: '#DCFCE7', stroke: '#166534' },
            ]
        case 'free':
            return [
                { label: 'Нет задач', color: '#F0FDFA', stroke: '#99F6E4' },
                { label: 'Свободные задачи', color: '#E0F2FE', stroke: '#0369A1' },
            ]
        case 'all':
            return [
                { label: 'Нет задач', color: '#F0FDFA', stroke: '#99F6E4' },
                { label: 'На проверке', color: '#FEF3C7', stroke: '#92400E' },
                { label: 'Выполнены и проверены', color: '#DCFCE7', stroke: '#166534' },
                { label: 'Задачи не взяты или не выполнены', color: '#FEE2E2', stroke: '#991B1B' },
            ]
        case 'verification':
            return [
                { label: 'Нет задач', color: '#F0FDFA', stroke: '#99F6E4' },
                { label: 'На проверке', color: '#FEF3C7', stroke: '#92400E' },
                { label: 'Выполнены и проверены', color: '#DCFCE7', stroke: '#166534' },
                { label: 'Задачи не взяты или не выполнены', color: '#FEE2E2', stroke: '#991B1B' },
            ]
        default:
            return [{ label: 'Нет задач', color: '#F0FDFA', stroke: '#99F6E4' }]
    }
}

type Props = {
    currentUser?: CurrentUser | null
    selectedGroupId?: string
}

const MY_GROUP_OPTION_VALUE = '__my_group__'

export function CurrentDutyPage({ currentUser = null, selectedGroupId: initialGroupId }: Props) {
    const selectedDormitoryId = useSelectedDormitoryId()
    const {
        selectedGroupId,
        duty,
        error,
        loading,
        pendingTaskId,
        selectGroup,
        handleTake,
        handleReturn,
        handleComplete,
        handleOpen,
        handleReopen,
        handleVerify,
        selectGroupForDormitory,
        reloadCurrentDuty,
        notification,
        clearNotification,
        visibleMineTaskIds,
        visibleFreeTaskIds,
    } = useCurrentDuty(initialGroupId, selectedDormitoryId)
    const [createModalOpened, setCreateModalOpened] = useState(false)
    const [displayMode, setDisplayMode] = useState<'list' | 'plan'>('list')
    const [selectedFloorPlanId, setSelectedFloorPlanId] = useState('')
    const [legendOpened, setLegendOpened] = useState(false)
    const isMobile = useMediaQuery('(max-width: 48em)')
    const visibleTabs = duty ? selectVisibleDutyTabs(duty) : []
    const allowedMobileTabs: DutyTaskSelect[] = ['mine', 'all', 'team']
    const [activeSelect, setActiveSelect] = useStoredDutySelect(isMobile ? allowedMobileTabs : visibleTabs)
    const availableFloorPlans = useMemo(() => [...floorPlans].sort((left, right) => left.floor - right.floor), [])
    const analytics = duty ? calculateDutyAnalytics(duty.tasks) : null
    const viewOptions = duty ? selectDutyViewOptions(duty, activeSelect, visibleTabs) : null
    const isCurrentUserInDutyTeam = duty != null && currentUser != null
        ? duty.team_members.some((member) => member.id === currentUser.id)
        : false
    const mobileHeaderContent =
        duty
        && analytics
        && viewOptions
        && duty.has_active_duty
        && duty.tasks.length > 0
        && isCurrentUserInDutyTeam
        && (viewOptions.showAnalytics || viewOptions.isReadOnly) ? (
            <MobileDutyProgress
                analytics={analytics}
                isReadOnly={viewOptions.isReadOnly}
                targetValue={duty.cost_per_resident_goal}
            />
        ) : null

    useMobileHeaderContent(mobileHeaderContent)

    useEffect(() => {
        if (!isMobile) {
            return
        }

        if (isSupportedMobileDutySelect(activeSelect)) {
            return
        }

        setActiveSelect('all')
    }, [activeSelect, isMobile, setActiveSelect])

    useEffect(() => {
        if (availableFloorPlans.length === 0) {
            if (selectedFloorPlanId !== '') {
                setSelectedFloorPlanId('')
            }

            return
        }

        if (!availableFloorPlans.some((plan) => String(plan.floor) === selectedFloorPlanId)) {
            setSelectedFloorPlanId(String(availableFloorPlans[0]?.floor ?? ''))
        }
    }, [availableFloorPlans, selectedFloorPlanId])

    useEffect(() => {
        if (!notification) {
            return
        }

        const timeoutId = window.setTimeout(() => {
            clearNotification()
        }, 5000)

        return () => {
            window.clearTimeout(timeoutId)
        }
    }, [clearNotification, notification])

    const groupSelectOptions = useMemo(() => {
        if (!duty) {
            return []
        }

        const options = duty.groups.map((group) => ({
            value: group.id,
            label: group.name,
        }))

        if (!duty.my_group || duty.my_dormitory_id == null) {
            return options
        }

        return [
            { value: MY_GROUP_OPTION_VALUE, label: 'Моя группа' },
            ...options,
        ]
    }, [duty])

    if (loading) {
        return (
            <Center py="xl">
                <Loader />
            </Center>
        )
    }

    if (error && !duty) {
        return (
            <Box px={{ base: 'md', md: 'xl' }} py="xl">
                <Alert color="red" title="Ошибка">
                    {error}
                </Alert>
            </Box>
        )
    }

    if (!duty) {
        return (
            <Box px={{ base: 'md', md: 'xl' }} py="xl">
                <Alert color="gray">Не удалось загрузить текущее дежурство</Alert>
            </Box>
        )
    }

    const resolvedAnalytics = analytics
    const resolvedViewOptions = viewOptions

    if (!resolvedAnalytics || !resolvedViewOptions) {
        return (
            <Box px={{ base: 'md', md: 'xl' }} py="xl">
                <Alert color="gray">Не удалось подготовить данные дежурства</Alert>
            </Box>
        )
    }

    const displayedTasks = selectTasksForActiveSelect({
        activeSelect,
        duty,
        visibleFreeTaskIds,
        visibleMineTaskIds,
    })
    const mineTasks = selectTasksForMobileMine(duty)
    const planLegendItems = getPlanLegendItems(activeSelect)

    const selectedFloorPlan =
        availableFloorPlans.find((plan) => String(plan.floor) === selectedFloorPlanId) ??
        availableFloorPlans[0] ??
        null

    const canManageGroupDuty = duty.can_manage_duty_settings

    const titleActions = duty ? (
        <Group gap="sm" align="center" wrap="wrap">
            {duty.show_group_select ? (
                <Select
                    aria-label="Группа"
                    autoComplete="off"
                    data={groupSelectOptions}
                    value={selectedGroupId ?? duty.selected_group_id}
                    allowDeselect={false}
                    w={{ base: '100%', sm: 280 }}
                    styles={{
                        input: {
                            minHeight: 42,
                            height: 42,
                        },
                    }}
                    onChange={(value) => {
                        if (!value) {
                            return
                        }

                        if (value === MY_GROUP_OPTION_VALUE && duty.my_group && duty.my_dormitory_id != null) {
                            selectGroupForDormitory(duty.my_group.id, String(duty.my_dormitory_id))
                            return
                        }

                        selectGroup(value)
                    }}
                />
            ) : null}

            {canManageGroupDuty ? (
                <Menu shadow="md" width={292} position="bottom-end">
                    <Menu.Target>
                        <ActionIcon
                            variant="subtle"
                            color="gray"
                            radius="md"
                            size={42}
                            aria-label="Действия дежурства"
                        >
                            <DotsThreeOutlineIcon size={24} weight="fill" />
                        </ActionIcon>
                    </Menu.Target>

                    <Menu.Dropdown>
                        <Menu.Item
                            leftSection={<PlusIcon size={24} />}
                            onClick={() => setCreateModalOpened(true)}
                        >
                            Новое дежурство
                        </Menu.Item>
                        <Menu.Item
                            leftSection={<BookOpenIcon size={24} />}
                            onClick={() => navigateTo(`/app/duties/history?group_id=${encodeURIComponent(duty.selected_group_id)}`)}
                        >
                            История дежурств
                        </Menu.Item>
                        <Menu.Item
                            leftSection={<GearIcon size={24} />}
                            onClick={() => navigateTo(`/app/groups/${duty.selected_group_id}/duty-settings`)}
                        >
                            Настройки дежурства
                        </Menu.Item>
                    </Menu.Dropdown>
                </Menu>
            ) : null}
        </Group>
    ) : null

    const dutyControls = resolvedViewOptions.showControls ? (
        <Box visibleFrom="md">
            <DutyTaskSelects
                activeSelect={activeSelect}
                visibleSelects={visibleTabs}
                rightSection={activeSelect === 'team' ? undefined : (
                    <SegmentedControl
                        value={displayMode}
                        data={[
                            { label: 'Список', value: 'list' },
                            { label: 'План', value: 'plan' },
                        ]}
                        classNames={{
                            control: segmentedControlClasses.control,
                            root: segmentedControlClasses.root,
                            label: segmentedControlClasses.label,
                        }}
                        onChange={(value) => setDisplayMode(value as 'list' | 'plan')}
                    />
                )}
                onChange={setActiveSelect}
            />
        </Box>
    ) : null

    const buildingPlanControls = activeSelect === 'team' || displayMode !== 'plan' ? null : (
        <Group justify="space-between" align="center" gap="md" wrap="wrap">
            <Popover
                opened={legendOpened}
                position="bottom-start"
                withArrow
                shadow="md"
            >
                <Popover.Target>
                    <Text
                        span
                        c="dimmed"
                        style={{ cursor: 'default' }}
                        onMouseEnter={() => setLegendOpened(true)}
                        onMouseLeave={() => setLegendOpened(false)}
                    >
                        ⓘ Обозначения
                    </Text>
                </Popover.Target>
                <Popover.Dropdown
                    onMouseEnter={() => setLegendOpened(true)}
                    onMouseLeave={() => setLegendOpened(false)}
                >
                    <Stack gap="xs">
                        {planLegendItems.map((item) => (
                            <Group key={item.label} gap="xs" wrap="nowrap">
                                <Box
                                    style={{
                                        width: 10,
                                        height: 10,
                                        minWidth: 10,
                                        borderRadius: '50%',
                                        backgroundColor: item.color,
                                        border: `1px solid ${item.stroke}`,
                                    }}
                                />
                                <Text size="sm">{item.label}</Text>
                            </Group>
                        ))}
                    </Stack>
                </Popover.Dropdown>
            </Popover>

            <Select
                aria-label="Этаж"
                value={selectedFloorPlanId}
                data={availableFloorPlans.map((plan) => ({
                    value: String(plan.floor),
                    label: getFloorLabel(plan.floor),
                }))}
                allowDeselect={false}
                w={220}
                onChange={(value) => {
                    if (value) {
                        setSelectedFloorPlanId(value)
                    }
                }}
            />
        </Group>
    )

    const analyticsBlock = resolvedViewOptions.showAnalytics || resolvedViewOptions.isReadOnly ? (
        <Box visibleFrom="md">
            <CurrentDutyAnalytics
                analytics={resolvedAnalytics}
                isReadOnly={resolvedViewOptions.isReadOnly}
                targetValue={duty.cost_per_resident_goal}
            />
        </Box>
    ) : undefined

    const controls = analyticsBlock || dutyControls || buildingPlanControls ? (
        <Stack gap="md">
            {analyticsBlock}
            {dutyControls}
            {buildingPlanControls}
        </Stack>
    ) : undefined
    const shouldShowMobileMineEmptyState = activeSelect === 'mine' && mineTasks.length === 0
    const mobileMineEmptyState = !isCurrentUserInDutyTeam ? (
        <EmptyState
            icon={<CalendarXIcon size={32} />}
            title="Текущее дежурство не Ваше"
            description="Вам не нужно выбирать задачи на текущее дежурство"
        />
    ) : (
        <EmptyState
            title="У вас нет взятых задач"
            description={(
                <>
                    Возьмите задачи в разделе <Text component="span" fw={700}>Задачи</Text>
                </>
            )}
        />
    )

    if (!duty.has_active_duty) {
        return (
            <PageFrame title="Дежурство" titleActions={titleActions} error={error}>
                <EmptyState
                    title="Дежурство не найдено"
                    description="В выбранной группе нет активного дежурства."
                />
                <CreateGroupDutyModal
                    opened={createModalOpened}
                    groupId={duty.selected_group_id}
                    onClose={() => setCreateModalOpened(false)}
                    onCreated={reloadCurrentDuty}
                />
            </PageFrame>
        )
    }

    if (duty.tasks.length === 0) {
        return (
            <PageFrame
                title="Дежурство"
                mobileTitle={formatDutyPeriod(duty.start_date, duty.end_date)}
                mobileSubtitle={null}
                subtitle={formatDutyPeriod(duty.start_date, duty.end_date)}
                titleActions={titleActions}
                error={error}
                notice={duty.notice_message}
                noticeTone={duty.notice_tone}
                controls={controls}
            >
                <EmptyState
                    title="Задачи не найдены"
                    description="Для текущего дежурства задачи ещё не заведены."
                />
                <CreateGroupDutyModal
                    opened={createModalOpened}
                    groupId={duty.selected_group_id}
                    onClose={() => setCreateModalOpened(false)}
                    onCreated={reloadCurrentDuty}
                />
            </PageFrame>
        )
    }

    return (
        <>
            {notification ? (
                <FloatingNotification
                    key={notification.id}
                    color={notification.color}
                    title={notification.title}
                    message={notification.message}
                    onClose={clearNotification}
                />
            ) : null}

            <PageFrame
                title="Дежурство"
                mobileTitle={formatDutyPeriod(duty.start_date, duty.end_date)}
                mobileSubtitle={null}
                subtitle={formatDutyPeriod(duty.start_date, duty.end_date)}
                titleActions={titleActions}
                error={error}
                notice={duty.notice_message}
                noticeTone={duty.notice_tone}
                controls={controls}
            >
                {displayMode === 'plan' && activeSelect !== 'team' ? (
                    <>
                        {selectedFloorPlan ? (
                            <BuildingPlanPanel
                                key={selectedFloorPlan.floor}
                                activeSelect={activeSelect}
                                actionMode={resolvedViewOptions.actionMode}
                                allTasks={duty.tasks}
                                floorPlan={selectedFloorPlan}
                                isReadOnly={resolvedViewOptions.isReadOnly}
                                pendingTaskId={pendingTaskId}
                                tasks={displayedTasks}
                                onTake={handleTake}
                                onReturn={handleReturn}
                                onComplete={handleComplete}
                                onOpen={handleOpen}
                                onReopen={handleReopen}
                                onVerify={handleVerify}
                            />
                        ) : (
                            <EmptyState
                                title="План здания не найден"
                                description="Для выбранного общежития план здания не настроен."
                            />
                        )}
                        <Box hiddenFrom="md">
                            <TaskGroups
                                isReadOnly={resolvedViewOptions.isReadOnly}
                                actionMode={resolvedViewOptions.actionMode}
                                pendingTaskId={pendingTaskId}
                                tasks={displayedTasks}
                                emptyMessage={resolvedViewOptions.emptyMessage}
                                onTake={handleTake}
                                onReturn={handleReturn}
                                onComplete={handleComplete}
                                onOpen={handleOpen}
                                onReopen={handleReopen}
                                onVerify={handleVerify}
                            />
                        </Box>
                    </>
                ) : activeSelect === 'team' ? (
                    <TeamMemberTaskGroups duty={duty} />
                ) : shouldShowMobileMineEmptyState ? (
                    mobileMineEmptyState
                ) : (
                    <TaskGroups
                        isReadOnly={resolvedViewOptions.isReadOnly}
                        actionMode={activeSelect === 'mine' ? 'default' : resolvedViewOptions.actionMode}
                        pendingTaskId={pendingTaskId}
                        tasks={activeSelect === 'mine' ? mineTasks : displayedTasks}
                        emptyMessage={resolvedViewOptions.emptyMessage}
                        onTake={handleTake}
                        onReturn={handleReturn}
                        onComplete={handleComplete}
                        onOpen={handleOpen}
                        onReopen={handleReopen}
                        onVerify={handleVerify}
                    />
                )}
                <Box hiddenFrom="md" h={198} />
                <CreateGroupDutyModal
                    opened={createModalOpened}
                    groupId={duty.selected_group_id}
                    onClose={() => setCreateModalOpened(false)}
                    onCreated={reloadCurrentDuty}
                />
            </PageFrame>

            <MobileDutyBottomNav
                activeSelect={activeSelect === 'mine' || activeSelect === 'all' || activeSelect === 'team' ? activeSelect : 'all'}
                onChange={setActiveSelect}
            />
        </>
    )
}
