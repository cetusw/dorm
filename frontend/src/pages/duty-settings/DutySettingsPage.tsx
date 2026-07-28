import { useEffect, useMemo, useState } from 'react'

import { CaretLeftIcon } from '@phosphor-icons/react'
import { Alert, Box, Center, Group, Loader, SegmentedControl, Select, Stack } from '@mantine/core'

import {
    createDutySettingsTask,
    deleteDutySettingsTask,
    excludeDutySettingsTask,
    getDutySettings,
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
import { TaskFormModal } from '../../features/task-catalog/ui/TaskFormModal'
import { ConfirmActionModal } from '../../shared/ui/ConfirmActionModal'
import segmentedControlClasses from '../../features/current-duty/ui/SegmentedControl.module.css'
import { PageFrame } from '../../shared/ui/PageFrame'
import { navigateTo } from '../../app/navigation'

type Props = {
    groupId: string
}

export function DutySettingsPage({ groupId }: Props) {
    const { data, error, forbidden, loading, reload } = useDutySettings(groupId)
    const [mainTab, setMainTab] = useState<DutySettingsMainTab>('tasks')
    const [viewMode, setViewMode] = useState<DutySettingsViewMode>('list')
    const [selectedFloorPlanId, setSelectedFloorPlanId] = useState('')
    const [selectedAreaId, setSelectedAreaId] = useState<string | null>(null)
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
    const selectedArea = data?.areas.find((area) => String(area.id) === selectedAreaId) ?? null
    const selectedFloorPlan = availableFloorPlans.find((plan) => String(plan.floor) === selectedFloorPlanId) ?? null
    const hasTaskModalOpen = taskAreaId !== null || editingTaskId !== null || excludingTaskId !== null || deletingTaskId !== null

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

    let content = null

    if (loading) {
        content = (
            <Center py="xl">
                <Loader />
            </Center>
        )
    } else if (forbidden) {
        content = <Alert color="red">{error ?? 'У вас нет доступа к настройкам дежурства этой группы.'}</Alert>
    } else if (error && !data) {
        content = <Alert color="red">{error}</Alert>
    } else if (!data) {
        content = <Alert color="gray">Не удалось загрузить настройки дежурства.</Alert>
    } else if (mainTab === 'teams') {
        content = (
            <DutySettingsTeamsTab
                groupId={groupId}
                teams={data.teams}
                activeDutyTeamId={data.active_duty_team_id}
                onReload={() => reload({ silent: true })}
            />
        )
    } else if (mainTab !== 'tasks') {
        content = <Box h={120} />
    } else if (data.task_editor_state !== 'active' || !data.active_duty) {
        content = <Alert color="gray">{data.task_editor_alert}</Alert>
    } else {
        const floorAreas = selectedFloorPlan ? data.areas.filter((area) => area.floor === selectedFloorPlan.floor) : []

        content = (
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
    }

    return (
        <PageFrame
            topContent={(
                <button
                    type="button"
                    onClick={() => navigateTo('/app')}
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
                    <DutySettingsCreateTaskModal
                        opened={taskAreaId !== null && editingTaskId === null}
                        onClose={handleCloseTaskModal}
                        onCreate={async (request) => {
                            await createDutySettingsTask(groupId, Number(taskAreaId), request)
                            await reload({ silent: true })
                        }}
                    />

                    <TaskFormModal
                        opened={editingTaskId !== null}
                        mode="edit"
                        appearance="settings"
                        taskId={editingTaskId}
                        dormitoryId={String(data.group.dormitory_id)}
                        initialAreaId={taskAreaId == null ? null : String(taskAreaId)}
                        hideAreaField={false}
                        loadAreas={async () => ({
                            areas: data.areas.map((area) => ({
                                id: area.id,
                                name: area.name,
                                floor: area.floor,
                                group: {
                                    id: data.group.id,
                                    name: data.group.name,
                                },
                            })),
                        })}
                        loadTask={async (taskId) => {
                            const freshData = await getDutySettings(groupId)
                            for (const area of freshData.areas) {
                                const task = area.tasks.find((item) => item.id === taskId)
                                if (task) {
                                    return {
                                        id: task.id,
                                        title: task.title,
                                        cost: task.cost,
                                        frequency: task.frequency,
                                        area: {
                                            id: area.id,
                                            name: area.name,
                                        },
                                    }
                                }
                            }

                            throw new Error('Не удалось загрузить данные задачи')
                        }}
                        updateTaskRequest={async (taskId, request) => {
                            const updated = await updateDutySettingsTask(groupId, taskId, request)
                            await reload({ silent: true })
                            return updated
                        }}
                        onClose={handleCloseTaskModal}
                        onSaved={async () => {}}
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
