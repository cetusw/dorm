import { useEffect, useMemo, useState } from 'react'

import { CaretLeftIcon } from '@phosphor-icons/react'
import { Alert, Box, Center, Group, Loader, Popover, SegmentedControl, Select, Stack, Text } from '@mantine/core'

import { navigateTo } from '../../app/navigation'
import { calculateDutyAnalytics, selectDutyViewOptions, selectTasksForActiveSelect } from '../../features/current-duty/model/selectors'
import { useStoredDutySelect } from '../../features/current-duty/model/useStoredDutySelect'
import { formatDutyPeriod } from '../../features/current-duty/model/utils'
import { BuildingPlanPanel } from '../../features/current-duty/building-plan/BuildingPlanPanel'
import { floorPlans } from '../../features/current-duty/building-plan/generated/plans'
import { getFloorLabel } from '../../features/current-duty/building-plan/utils'
import { CurrentDutyAnalytics } from '../../features/current-duty/ui/CurrentDutyAnalytics'
import { DutyTaskSelects } from '../../features/current-duty/ui/DutyTaskSelects'
import { TaskGroups } from '../../features/current-duty/ui/TaskGroups'
import { TeamMemberTaskGroups } from '../../features/current-duty/ui/TeamMemberTaskGroups'
import { useDutyDetails } from '../../features/duty-history/model/useDutyDetails'
import segmentedControlClasses from '../../features/current-duty/ui/SegmentedControl.module.css'
import { PageFrame } from '../../shared/ui/PageFrame'
import type { DutyTaskSelect } from '../../features/current-duty/model/types'

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
    dutyId: string
}

export function DutyHistoryDetailsPage({ dutyId }: Props) {
    const { duty, loading, error } = useDutyDetails(dutyId)
    const visibleTabs = duty ? duty.visible_tabs : []
    const [activeSelect, setActiveSelect] = useStoredDutySelect(visibleTabs)
    const [displayMode, setDisplayMode] = useState<'list' | 'plan'>('list')
    const [selectedFloorPlanId, setSelectedFloorPlanId] = useState('')
    const [legendOpened, setLegendOpened] = useState(false)
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

    if (loading) {
        return <Center py="xl"><Loader /></Center>
    }

    if (error || !duty) {
        return (
            <PageFrame title="Дежурство" error={error ?? 'Не удалось загрузить дежурство'}>
                <Alert color="gray">Дежурство недоступно</Alert>
            </PageFrame>
        )
    }

    const selectedGroupId = new URLSearchParams(window.location.search).get('group_id') ?? duty.selected_group_id
    const viewOptions = selectDutyViewOptions(duty, activeSelect, visibleTabs)
    const displayedTasks = selectTasksForActiveSelect({
        activeSelect,
        duty,
        visibleFreeTaskIds: [],
        visibleMineTaskIds: [],
    })
    const analytics = calculateDutyAnalytics(duty.tasks)
    const planLegendItems = getPlanLegendItems(activeSelect)
    const selectedFloorPlan =
        availableFloorPlans.find((plan) => String(plan.floor) === selectedFloorPlanId) ??
        availableFloorPlans[0] ??
        null

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
    ) : undefined

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
    const noopTaskAction = async () => false

    return (
        <PageFrame
            topContent={(
                <button
                    type="button"
                    onClick={() => navigateTo(`/app/duties/history?group_id=${encodeURIComponent(selectedGroupId)}`)}
                    style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '4px',
                        width: 'fit-content',
                        padding: 0,
                        border: 0,
                        background: 'transparent',
                        color: 'var(--mantine-color-gray-7)',
                        cursor: 'pointer',
                        fontSize: '1rem',
                        lineHeight: 1.5,
                        fontWeight: 500,
                    }}
                >
                    <CaretLeftIcon size={20} />
                    <span>История дежурств</span>
                </button>
            )}
            title={duty.team || 'Глава команды не назначен'}
            subtitle={formatDutyPeriod(duty.start_date, duty.end_date)}
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
                            isReadOnly
                            pendingTaskId={null}
                            tasks={displayedTasks}
                            onTake={noopTaskAction}
                            onReturn={noopTaskAction}
                            onComplete={noopTaskAction}
                            onOpen={noopTaskAction}
                            onReopen={noopTaskAction}
                            onVerify={noopTaskAction}
                        />
                    ) : (
                        <Alert color="gray">Для выбранного общежития план здания не настроен.</Alert>
                    )}
                    <Box hiddenFrom="md">
                        <TaskGroups
                            tasks={displayedTasks}
                            actionMode={viewOptions.actionMode}
                            emptyMessage={viewOptions.emptyMessage}
                            isReadOnly
                            pendingTaskId={null}
                            onTake={noopTaskAction}
                            onReturn={noopTaskAction}
                            onComplete={noopTaskAction}
                            onOpen={noopTaskAction}
                            onVerify={noopTaskAction}
                            onReopen={noopTaskAction}
                        />
                    </Box>
                </>
            ) : activeSelect === 'team' ? (
                <TeamMemberTaskGroups duty={duty} />
            ) : (
                <TaskGroups
                    tasks={displayedTasks}
                    actionMode={viewOptions.actionMode}
                    emptyMessage={viewOptions.emptyMessage}
                    isReadOnly
                    pendingTaskId={null}
                    onTake={noopTaskAction}
                    onReturn={noopTaskAction}
                    onComplete={noopTaskAction}
                    onOpen={noopTaskAction}
                    onVerify={noopTaskAction}
                    onReopen={noopTaskAction}
                />
            )}
        </PageFrame>
    )
}
