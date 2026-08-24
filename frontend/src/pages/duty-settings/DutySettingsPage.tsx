import { useEffect, useMemo, useState } from 'react'

import { CaretLeftIcon } from '@phosphor-icons/react'
import { Alert, Box, Center, Group, Loader, SegmentedControl, Select, Stack } from '@mantine/core'

import { AreaFormModal } from '../../features/areas/ui/AreaFormModal'
import { DeleteAreaModal } from '../../features/areas/ui/DeleteAreaModal'
import {
    createDutySettingsArea,
    deleteDutySettingsArea,
    createDutySettingsTask,
    updateDutySettingsArea,
    deleteDutySettingsTask,
    excludeDutySettingsTask,
    includeDutySettingsTask,
    updateDutySettingsTask,
} from '../../features/duty-settings/api/dutySettingsApi'
import { useDutySettings } from '../../features/duty-settings/model/useDutySettings'
import type { DutySettingsMainTab, DutySettingsViewMode } from '../../features/duty-settings/model/types'
import { DutySettingsAreaDrawer } from '../../features/duty-settings/ui/DutySettingsAreaDrawer'
import { DutySettingsAreaList } from '../../features/duty-settings/ui/DutySettingsAreaList'
import { DutySettingsCreateTaskModal } from '../../features/duty-settings/ui/DutySettingsCreateTaskModal'
import { DutySettingsPlanPanel } from '../../features/duty-settings/ui/DutySettingsPlanPanel'
import { DutySettingsTaskSummary } from '../../features/duty-settings/ui/DutySettingsTaskSummary'
import { DutySettingsTeamsTab } from '../../widgets/duty-settings-teams/ui/DutySettingsTeamsTab'
import { floorPlans } from '../../features/current-duty/building-plan/generated/plans'
import { getFloorLabel } from '../../features/current-duty/building-plan/utils'
import { ConfirmActionModal } from '../../shared/ui/ConfirmActionModal'
import { EmptyState } from '../../shared/ui/EmptyState'
import segmentedControlClasses from '../../features/current-duty/ui/SegmentedControl.module.css'
import { PageFrame } from '../../shared/ui/PageFrame'
import { navigateTo } from '../../app/navigation'

type Props = {
    groupId: string
}

export function DutySettingsPage({ groupId }: Props) {
    const { data, error, forbidden, loading, reload, changeTeamMembersCount } = useDutySettings(groupId)
    const [mainTab, setMainTab] = useState<DutySettingsMainTab>('tasks')
    const [viewMode, setViewMode] = useState<DutySettingsViewMode>('list')
    const [selectedFloorPlanId, setSelectedFloorPlanId] = useState('')
    const [selectedAreaId, setSelectedAreaId] = useState<string | null>(null)
    const [createAreaOpened, setCreateAreaOpened] = useState(false)
    const [editingAreaId, setEditingAreaId] = useState<number | null>(null)
    const [deletingAreaId, setDeletingAreaId] = useState<number | null>(null)
    const [editingTaskId, setEditingTaskId] = useState<string | null>(null)
    const [taskAreaId, setTaskAreaId] = useState<number | null>(null)
    const [excludingTaskId, setExcludingTaskId] = useState<string | null>(null)
    const [deletingTaskId, setDeletingTaskId] = useState<string | null>(null)
    const availableFloorPlans = useMemo(() => [...floorPlans].sort((left, right) => left.floor - right.floor), [])

    const selectedTask = useMemo(() => {
        if (!data) {
            return null
        }

        for (const area of data.areas) {
            const task = area.tasks.find((item) => item.id === excludingTaskId || item.id === deletingTaskId)
            if (task) {
                return task
            }
        }

        return null
    }, [data, deletingTaskId, excludingTaskId])
    const excludingTask = excludingTaskId ? selectedTask : null
    const deletingTask = deletingTaskId ? selectedTask : null
    const editingTask = useMemo(() => {
        if (!data || !editingTaskId) {
            return null
        }

        for (const area of data.areas) {
            const task = area.tasks.find((item) => item.id === editingTaskId)
            if (task) {
                return {
                    id: task.id,
                    title: task.title,
                    cost: task.cost,
                    recurrenceInterval: task.recurrenceInterval,
                    startSequence: task.startSequence,
                    area: {
                        id: area.id,
                        name: area.name,
                    },
                }
            }
        }

        return null
    }, [data, editingTaskId])
    const selectedArea = data?.areas.find((area) => String(area.id) === selectedAreaId) ?? null
    const deletingArea = data?.areas.find((area) => area.id === deletingAreaId) ?? null
    const selectedFloorPlan = availableFloorPlans.find((plan) => String(plan.floor) === selectedFloorPlanId) ?? null
    const hasTaskModalOpen = createAreaOpened || editingAreaId !== null || deletingAreaId !== null || taskAreaId !== null || editingTaskId !== null || excludingTaskId !== null || deletingTaskId !== null

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

    function handleCreateTask(areaId: number) {
        setTaskAreaId(areaId)
        setEditingTaskId(null)
    }

    function handleEditTask(taskId: string) {
        if (!data) {
            return
        }

        for (const area of data.areas) {
            if (area.tasks.some((task) => task.id === taskId)) {
                setTaskAreaId(area.id)
                setEditingTaskId(taskId)
                return
            }
        }
    }

    function handleCloseTaskModal() {
        setTaskAreaId(null)
        setEditingTaskId(null)
    }

    function handleOpenArea(areaId: string) {
        setSelectedAreaId(areaId)
    }

    function handleCreateArea() {
        setCreateAreaOpened(true)
        setEditingAreaId(null)
    }

    function handleEditArea(areaId: number) {
        setCreateAreaOpened(false)
        setEditingAreaId(areaId)
    }

    function handleCloseAreaFormModal() {
        setCreateAreaOpened(false)
        setEditingAreaId(null)
    }

    const content = loading ? (
            <Center py="xl">
                <Loader />
            </Center>
        ) : forbidden ? (
            <Alert color="red">{error ?? 'У вас нет доступа к настройкам дежурства этой группы.'}</Alert>
        ) : error && !data ? (
            <Alert color="red">{error}</Alert>
        ) : !data ? (
            <EmptyState
                title="Настройки не найдены"
                description="Не удалось загрузить настройки дежурства."
            />
        ) : mainTab === 'teams' ? (
            <DutySettingsTeamsTab
                groupId={groupId}
                teams={data.teams}
                activeDutyTeamId={data.active_duty_team_id}
                onReload={() => reload({ silent: true })}
                onMemberCountChange={changeTeamMembersCount}
            />
        ) : mainTab !== 'tasks' ? (
            <Box h={120} />
        ) : data.task_editor_state !== 'active' || !data.active_duty ? (
            <EmptyState
                title="Дежурство не найдено"
                description={data.task_editor_alert}
            />
        ) : (() => {
        const floorAreas = selectedFloorPlan ? data.areas.filter((area) => area.floor === selectedFloorPlan.floor) : []

        return (
            <Stack gap="xl">
                {viewMode === 'plan' ? (
                    <Group justify="space-between" align="center" gap="md" wrap="wrap">
                        <DutySettingsTaskSummary summary={data.active_duty.summary} />
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
                ) : (
                    <DutySettingsTaskSummary summary={data.active_duty.summary} />
                )}

                {viewMode === 'plan' && selectedFloorPlan ? (
                    <DutySettingsPlanPanel
                        areas={floorAreas}
                        floorPlan={selectedFloorPlan}
                        selectedAreaId={selectedAreaId}
                        onAreaClick={handleOpenArea}
                    />
                ) : (
                    <DutySettingsAreaList
                        areas={data.areas}
                        onCreateArea={handleCreateArea}
                        onEditArea={handleEditArea}
                        onDeleteArea={(areaId) => setDeletingAreaId(areaId)}
                        onCreateTask={handleCreateTask}
                        onEditTask={handleEditTask}
                        onIncludeTask={async (taskId) => {
                            await includeDutySettingsTask(groupId, taskId)
                            await reload({ silent: true })
                        }}
                        onExcludeTask={(taskId) => setExcludingTaskId(taskId)}
                        onDeleteTask={(taskId) => setDeletingTaskId(taskId)}
                    />
                )}
            </Stack>
        )
        })()

    return (
        <PageFrame
            topContent={(
                <button
                    type="button"
                    onClick={() => navigateTo(`/app/tasks?group_id=${encodeURIComponent(groupId)}`)}
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
                    <span>Дежурство</span>
                </button>
            )}
            title="Настройки дежурства"
            controls={(
                <Stack gap="md">
                    <Group justify="space-between" align="center" gap="md" wrap="wrap">
                        <SegmentedControl
                            value={mainTab}
                            data={[
                                { label: 'Задачи', value: 'tasks' },
                                { label: 'Команды', value: 'teams' },
                            ]}
                            classNames={{
                                control: segmentedControlClasses.control,
                                root: segmentedControlClasses.root,
                                label: segmentedControlClasses.label,
                            }}
                            onChange={(value) => setMainTab(value as DutySettingsMainTab)}
                        />

                        {mainTab === 'tasks' ? (
                            <SegmentedControl
                                value={viewMode}
                                data={[
                                    { label: 'Список', value: 'list' },
                                    { label: 'План', value: 'plan' },
                                ]}
                                classNames={{
                                    control: segmentedControlClasses.control,
                                    root: segmentedControlClasses.root,
                                    label: segmentedControlClasses.label,
                                }}
                                onChange={(value) => setViewMode(value as DutySettingsViewMode)}
                            />
                        ) : null}
                    </Group>
                </Stack>
            )}
        >
            {content}

            {data ? (
                <>
                    <AreaFormModal
                        opened={createAreaOpened || editingAreaId !== null}
                        mode={editingAreaId === null ? 'create' : 'edit'}
                        areaId={editingAreaId}
                        dormitoryId={String(data.group.dormitory_id)}
                        fixedGroupId={data.group.id}
                        hideGroupField
                        requireFloor
                        variant="settings"
                        onClose={handleCloseAreaFormModal}
                        onSaved={async () => {
                            await reload({ silent: true })
                        }}
                        loadArea={async (areaId) => {
                            const area = data.areas.find((item) => item.id === areaId)
                            if (!area) {
                                throw new Error('Не удалось загрузить данные территории')
                            }

                            return {
                                id: area.id,
                                name: area.name,
                                floor: area.floor,
                                group: {
                                    id: data.group.id,
                                    name: data.group.name,
                                },
                            }
                        }}
                        createAreaRequest={(request) => createDutySettingsArea(groupId, request)}
                        updateAreaRequest={(areaId, request) => updateDutySettingsArea(groupId, areaId, request)}
                    />

                    <DeleteAreaModal
                        opened={deletingArea !== null}
                        area={deletingArea ? {
                            id: deletingArea.id,
                            name: deletingArea.name,
                            floor: deletingArea.floor,
                            group: {
                                id: data.group.id,
                                name: data.group.name,
                            },
                        } : null}
                        dormitoryId={String(data.group.dormitory_id)}
                        deleteAreaRequest={(areaId) => deleteDutySettingsArea(groupId, areaId)}
                        onClose={() => setDeletingAreaId(null)}
                        onDeleted={async () => {
                            if (selectedArea && deletingArea && selectedArea.id === deletingArea.id) {
                                setSelectedAreaId(null)
                            }
                            setDeletingAreaId(null)
                            await reload({ silent: true })
                        }}
                    />

                    <DutySettingsCreateTaskModal
                        opened={taskAreaId !== null && editingTaskId === null}
                        mode="create"
                        onClose={handleCloseTaskModal}
                        onCreate={async (request) => {
                            await createDutySettingsTask(groupId, Number(taskAreaId), request)
                            await reload({ silent: true })
                        }}
                    />

                    <DutySettingsCreateTaskModal
                        opened={editingTaskId !== null}
                        mode="edit"
                        taskId={editingTaskId}
                        task={editingTask}
                        onUpdate={async (taskId, request) => {
                            await updateDutySettingsTask(groupId, taskId, request)
                            await reload({ silent: true })
                        }}
                        onClose={handleCloseTaskModal}
                    />

                    <ConfirmActionModal
                        opened={excludingTask !== null}
                        title="Исключить задачу"
                        description={excludingTask?.assignee_name
                            ? `Задача назначена пользователю ${excludingTask.assignee_name}. Вы уверены, что хотите исключить её из дежурства?`
                            : `Вы уверены, что хотите исключить задачу «${excludingTask?.title ?? ''}» из дежурства?`}
                        confirmLabel="Исключить"
                        confirmColor="red"
                        onClose={() => setExcludingTaskId(null)}
                        onConfirm={async () => {
                            if (!excludingTask) {
                                return
                            }
                            await excludeDutySettingsTask(groupId, excludingTask.id)
                            await reload({ silent: true })
                        }}
                        errorMessage="Не удалось исключить задачу из дежурства"
                    />

                    <ConfirmActionModal
                        opened={deletingTask !== null}
                        title="Удалить задачу"
                        description={`Вы уверены, что хотите удалить задачу «${deletingTask?.title ?? ''}»? Она будет скрыта из списков и больше не попадет в дежурства.`}
                        confirmLabel="Удалить"
                        confirmColor="red"
                        onClose={() => setDeletingTaskId(null)}
                        onConfirm={async () => {
                            if (!deletingTask) {
                                return
                            }
                            await deleteDutySettingsTask(groupId, deletingTask.id)
                            await reload({ silent: true })
                        }}
                        errorMessage="Не удалось удалить задачу"
                    />

                    <DutySettingsAreaDrawer
                        area={selectedArea}
                        opened={selectedArea !== null}
                        hasNestedModalOpen={hasTaskModalOpen}
                        onClose={() => setSelectedAreaId(null)}
                        onCreateArea={handleCreateArea}
                        onEditArea={handleEditArea}
                        onDeleteArea={(areaId) => setDeletingAreaId(areaId)}
                        onCreateTask={handleCreateTask}
                        onEditTask={handleEditTask}
                        onIncludeTask={async (taskId) => {
                            await includeDutySettingsTask(groupId, taskId)
                            await reload({ silent: true })
                        }}
                        onExcludeTask={(taskId) => setExcludingTaskId(taskId)}
                        onDeleteTask={(taskId) => setDeletingTaskId(taskId)}
                    />
                </>
            ) : null}
        </PageFrame>
    )
}
