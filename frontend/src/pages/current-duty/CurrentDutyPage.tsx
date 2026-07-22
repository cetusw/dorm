import { useState } from 'react'

import { Alert, Box, Button, Center, Group, Loader, Popover, SegmentedControl, Select, Stack, Text } from '@mantine/core'

import type { CurrentUser } from '../../features/current-user/model/types'
import type { DutyTaskSelect } from '../../features/current-duty/model/types'
import {
    calculateDutyAnalytics,
    selectVisibleDutyTabs,
    selectTeamMemberTaskGroups,
    selectDutyViewOptions,
    selectTasksForActiveSelect,
} from '../../features/current-duty/model/selectors'
import { CreateDutyWeekModal } from '../../features/current-duty/ui/CreateDutyWeekModal'
import { useSelectedDormitoryId } from '../../features/dormitories/model/useDormitorySelection'
import { useCurrentDuty } from '../../features/current-duty/model/useCurrentDuty'
import { useStoredDutySelect } from '../../features/current-duty/model/useStoredDutySelect'
import { formatDutyPeriod } from '../../features/current-duty/model/utils'
import { CurrentDutyAnalytics } from '../../features/current-duty/ui/CurrentDutyAnalytics'
import { DutyTaskSelects } from '../../features/current-duty/ui/DutyTaskSelects'
import { TeamMemberTaskGroups } from '../../features/current-duty/ui/TeamMemberTaskGroups'
import { TaskGroups } from '../../features/current-duty/ui/TaskGroups'
import { BuildingPlanPanel } from '../../features/current-duty/building-plan/BuildingPlanPanel'
import { floorPlans } from '../../features/current-duty/building-plan/plans'
import { PageFrame } from '../../shared/ui/PageFrame'
import segmentedControlClasses from '../../features/current-duty/ui/SegmentedControl.module.css'

type Props = {
    currentUser: CurrentUser | null
}

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
        case 'review':
            return [
                { label: 'Нет задач', color: '#F0FDFA', stroke: '#99F6E4' },
                { label: 'На проверке', color: '#FEF3C7', stroke: '#92400E' },
                { label: 'Выполнены и проверены', color: '#DCFCE7', stroke: '#166534' },
                { label: 'Задачи не взяты или не выполнены', color: '#FEE2E2', stroke: '#991B1B' },
            ]
        case 'all':
        default:
            return [
                { label: 'Нет задач', color: '#F0FDFA', stroke: '#99F6E4' },
                { label: 'Есть задачи', color: '#E0F2FE', stroke: '#0369A1' },
            ]
    }
}

export function CurrentDutyPage({ currentUser }: Props) {
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
        reloadCurrentDuty,
        visibleMineTaskIds,
        visibleFreeTaskIds,
    } = useCurrentDuty()
    const selectedDormitoryId = useSelectedDormitoryId()
    const [createModalOpened, setCreateModalOpened] = useState(false)
    const [displayMode, setDisplayMode] = useState<'list' | 'plan'>('list')
    const [selectedFloorPlanId, setSelectedFloorPlanId] = useState(floorPlans[0].id)
    const [legendOpened, setLegendOpened] = useState(false)
    const visibleTabs = duty ? selectVisibleDutyTabs(duty) : []
    const [activeSelect, setActiveSelect] = useStoredDutySelect(visibleTabs)

    const titleActions = currentUser?.can_manage_dormitories ? (
        <Button
            radius="md"
            onClick={() => setCreateModalOpened(true)}
            disabled={!selectedDormitoryId}
        >
            + Новая дежурная неделя
        </Button>
    ) : undefined

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
    const teamTaskGroups = selectTeamMemberTaskGroups(duty)
    const analytics = calculateDutyAnalytics(duty.tasks)

    const selectedFloorPlan = floorPlans.find((plan) => plan.id === selectedFloorPlanId) ?? floorPlans[0]

    const dutyControls = viewOptions.showControls ? (
        <DutyTaskSelects
            activeSelect={activeSelect}
            groups={duty.groups}
            selectedGroupId={selectedGroupId ?? duty.selected_group_id}
            showGroupSelect={duty.show_group_select}
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
            onGroupChange={selectGroup}
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
                        data={floorPlans.map((plan) => ({
                            value: plan.id,
                            label: plan.name,
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
            <PageFrame title="Дежурство" titleActions={titleActions} error={error} controls={controls}>
                <Alert color="gray">В выбранной группе сейчас нет активного дежурства.</Alert>
                {selectedDormitoryId && (
                    <CreateDutyWeekModal
                        opened={createModalOpened}
                        dormitoryId={selectedDormitoryId}
                        onClose={() => setCreateModalOpened(false)}
                        onCreated={reloadCurrentDuty}
                    />
                )}
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
                controls={controls}
            >
                <Alert color="gray">На текущее дежурство не заведены задачи.</Alert>
                {selectedDormitoryId && (
                    <CreateDutyWeekModal
                        opened={createModalOpened}
                        dormitoryId={selectedDormitoryId}
                        onClose={() => setCreateModalOpened(false)}
                        onCreated={reloadCurrentDuty}
                    />
                )}
            </PageFrame>
        )
    }

    return (
        <PageFrame
            title="Дежурство"
            subtitle={formatDutyPeriod(duty.start_date, duty.end_date)}
            titleActions={titleActions}
            error={error}
            notice={duty.notice_message}
            controls={controls}
        >
            {displayMode === 'plan' && activeSelect !== 'team' ? (
                <>
                    <BuildingPlanPanel
                        key={selectedFloorPlan.id}
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
                    <Box hiddenFrom="md">
                        <TaskGroups
                            isReadOnly={viewOptions.isReadOnly}
                            actionMode={viewOptions.actionMode}
                            pendingTaskId={pendingTaskId}
                            tasks={displayedTasks}
                            emptyMessage={viewOptions.emptyMessage}
                            showAssigneeColumn={viewOptions.showAssigneeColumn}
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
                <TeamMemberTaskGroups groups={teamTaskGroups} />
            ) : (
                <TaskGroups
                    isReadOnly={viewOptions.isReadOnly}
                    actionMode={viewOptions.actionMode}
                    pendingTaskId={pendingTaskId}
                    tasks={displayedTasks}
                    emptyMessage={viewOptions.emptyMessage}
                    showAssigneeColumn={viewOptions.showAssigneeColumn}
                    onTake={handleTake}
                    onReturn={handleReturn}
                    onComplete={handleComplete}
                    onOpen={handleOpen}
                    onReopen={handleReopen}
                    onVerify={handleVerify}
                />
            )}
            {selectedDormitoryId && (
                <CreateDutyWeekModal
                    opened={createModalOpened}
                    dormitoryId={selectedDormitoryId}
                    onClose={() => setCreateModalOpened(false)}
                    onCreated={reloadCurrentDuty}
                />
            )}
        </PageFrame>
    )
}
