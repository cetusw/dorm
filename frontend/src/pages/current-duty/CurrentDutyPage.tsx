import { useEffect, useMemo, useState } from 'react'

import { BookOpenIcon, DotsThreeOutlineIcon, GearIcon, PlusIcon } from '@phosphor-icons/react'
import { ActionIcon, Alert, Box, Center, Group, Loader, Menu, Popover, SegmentedControl, Select, Stack, Text } from '@mantine/core'

import type { DutyTaskSelect } from '../../features/current-duty/model/types'
import {
    calculateDutyAnalytics,
    selectVisibleDutyTabs,
    selectDutyViewOptions,
    selectTasksForActiveSelect,
} from '../../features/current-duty/model/selectors'
import { CreateGroupDutyModal } from '../../features/current-duty/ui/CreateGroupDutyModal'
import { useCurrentDuty } from '../../features/current-duty/model/useCurrentDuty'
import { useStoredDutySelect } from '../../features/current-duty/model/useStoredDutySelect'
import { formatDutyPeriod } from '../../features/current-duty/model/utils'
import { CurrentDutyAnalytics } from '../../features/current-duty/ui/CurrentDutyAnalytics'
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
    selectedGroupId?: string
}

const MY_GROUP_OPTION_VALUE = '__my_group__'

export function CurrentDutyPage({ selectedGroupId: initialGroupId }: Props) {
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
    const visibleTabs = duty ? selectVisibleDutyTabs(duty) : []
    const [activeSelect, setActiveSelect] = useStoredDutySelect(visibleTabs)
    const availableFloorPlans = useMemo(() => [...floorPlans].sort((left, right) => left.floor - right.floor), [])

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

    const viewOptions = selectDutyViewOptions(duty, activeSelect, visibleTabs)
    const displayedTasks = selectTasksForActiveSelect({
        activeSelect,
        duty,
        visibleFreeTaskIds,
        visibleMineTaskIds,
    })
    const planLegendItems = getPlanLegendItems(activeSelect)
    const analytics = calculateDutyAnalytics(duty.tasks)

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

    const dutyControls = viewOptions.showControls ? (
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
    ) : null

    const buildingPlanControls = activeSelect === 'team' ? null : (
        <Box>
            {displayMode === 'plan' && (
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
            )}
        </Box>
    )

    const analyticsBlock = viewOptions.showAnalytics || viewOptions.isReadOnly ? (
        <CurrentDutyAnalytics
            analytics={analytics}
            isReadOnly={viewOptions.isReadOnly}
            targetValue={duty.cost_per_resident_goal}
        />
    ) : undefined

    const controls = analyticsBlock || dutyControls || buildingPlanControls ? (
        <Stack gap="md">
            {analyticsBlock}
            {dutyControls}
            {buildingPlanControls}
        </Stack>
    ) : undefined

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
                                actionMode={viewOptions.actionMode}
                                allTasks={duty.tasks}
                                floorPlan={selectedFloorPlan}
                                isReadOnly={viewOptions.isReadOnly}
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
                                isReadOnly={viewOptions.isReadOnly}
                                actionMode={viewOptions.actionMode}
                                pendingTaskId={pendingTaskId}
                                tasks={displayedTasks}
                                emptyMessage={viewOptions.emptyMessage}
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
                ) : (
                    <TaskGroups
                        isReadOnly={viewOptions.isReadOnly}
                        actionMode={viewOptions.actionMode}
                        pendingTaskId={pendingTaskId}
                        tasks={displayedTasks}
                        emptyMessage={viewOptions.emptyMessage}
                        onTake={handleTake}
                        onReturn={handleReturn}
                        onComplete={handleComplete}
                        onOpen={handleOpen}
                        onReopen={handleReopen}
                        onVerify={handleVerify}
                    />
                )}
                <CreateGroupDutyModal
                    opened={createModalOpened}
                    groupId={duty.selected_group_id}
                    onClose={() => setCreateModalOpened(false)}
                    onCreated={reloadCurrentDuty}
                />
            </PageFrame>
        </>
    )
}
