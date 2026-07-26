import { useEffect, useMemo, useState } from 'react'

import { CaretLeftIcon } from '@phosphor-icons/react'
import { Alert, Box, Center, Group, Loader, SegmentedControl, Select, Stack } from '@mantine/core'

import { deleteDutySettingsArea, deleteDutySettingsTask } from '../../features/duty-settings/api/dutySettingsApi'
import { useDutySettings } from '../../features/duty-settings/model/useDutySettings'
import type { DutySettingsMainTab, DutySettingsViewMode } from '../../features/duty-settings/model/types'
import { toAreaListItem, toTaskListItem } from '../../features/duty-settings/model/utils'
import { DutySettingsAreaDrawer } from '../../features/duty-settings/ui/DutySettingsAreaDrawer'
import { DutySettingsAreaList } from '../../features/duty-settings/ui/DutySettingsAreaList'
import { DutySettingsPlanPanel } from '../../features/duty-settings/ui/DutySettingsPlanPanel'
import { floorPlans } from '../../features/current-duty/building-plan/generated/plans'
import { getFloorLabel } from '../../features/current-duty/building-plan/utils'
import { AreaFormModal } from '../../features/areas/ui/AreaFormModal'
import { DeleteAreaModal } from '../../features/areas/ui/DeleteAreaModal'
import { TaskFormModal } from '../../features/task-catalog/ui/TaskFormModal'
import { DeleteTaskModal } from '../../features/task-catalog/ui/DeleteTaskModal'
import segmentedControlClasses from '../../features/current-duty/ui/SegmentedControl.module.css'
import { PageFrame } from '../../shared/ui/PageFrame'
import {
    createDutySettingsArea,
    createDutySettingsTask,
    updateDutySettingsArea,
    updateDutySettingsTask,
} from '../../features/duty-settings/api/dutySettingsApi'
import { navigateTo } from '../../app/navigation'

type Props = {
    groupId: string
}

export function DutySettingsPage({ groupId }: Props) {
    const {
        data,
        error,
        forbidden,
        loading,
        appendArea,
        replaceArea,
        removeArea,
        appendTask,
        updateTask,
        removeTask,
    } = useDutySettings(groupId)
    const [mainTab, setMainTab] = useState<DutySettingsMainTab>('tasks')
    const [viewMode, setViewMode] = useState<DutySettingsViewMode>('list')
    const [selectedFloorPlanId, setSelectedFloorPlanId] = useState('')
    const [selectedAreaId, setSelectedAreaId] = useState<string | null>(null)
    const [editingAreaId, setEditingAreaId] = useState<number | null>(null)
    const [deletingAreaId, setDeletingAreaId] = useState<number | null>(null)
    const [editingTaskId, setEditingTaskId] = useState<string | null>(null)
    const [deletingTaskId, setDeletingTaskId] = useState<string | null>(null)
    const [creatingAreaOpened, setCreatingAreaOpened] = useState(false)
    const [taskAreaId, setTaskAreaId] = useState<number | null>(null)

    const availableFloorPlans = useMemo(() => [...floorPlans].sort((left, right) => left.floor - right.floor), [])

    useEffect(() => {
        if (availableFloorPlans.length === 0) {
            setSelectedFloorPlanId('')
            return
        }

        if (!availableFloorPlans.some((plan) => String(plan.floor) === selectedFloorPlanId)) {
            setSelectedFloorPlanId(String(availableFloorPlans[0]?.floor ?? ''))
        }
    }, [availableFloorPlans, selectedFloorPlanId])

    useEffect(() => {
        if (!data || selectedAreaId === null) {
            return
        }

        if (!data.areas.some((area) => String(area.id) === selectedAreaId)) {
            setSelectedAreaId(null)
        }
    }, [data, selectedAreaId])

    const selectedFloorPlan = availableFloorPlans.find((plan) => String(plan.floor) === selectedFloorPlanId) ?? null
    const selectedArea = data?.areas.find((area) => String(area.id) === selectedAreaId) ?? null
    const deletingArea = data?.areas.find((area) => area.id === deletingAreaId) ?? null
    const deletingTaskContext = useMemo(() => {
        if (!data || !deletingTaskId) {
            return null
        }

        for (const area of data.areas) {
            const task = area.tasks.find((item) => item.id === deletingTaskId)
            if (task) {
                return { area, task }
            }
        }

        return null
    }, [data, deletingTaskId])

    function handleCreateArea() {
        setEditingAreaId(null)
        setCreatingAreaOpened(true)
    }

    function handleEditArea(areaId: number) {
        setCreatingAreaOpened(false)
        setEditingAreaId(areaId)
    }

    function handleCloseAreaModal() {
        setCreatingAreaOpened(false)
        setEditingAreaId(null)
    }

    function handleCreateTask(areaId: number) {
        setTaskAreaId(areaId)
        setEditingTaskId(null)
    }

    function handleEditTask(taskId: string) {
        let areaId: number | null = null
        for (const area of data?.areas ?? []) {
            if (area.tasks.some((task) => task.id === taskId)) {
                areaId = area.id
                break
            }
        }
        setTaskAreaId(areaId)
        setEditingTaskId(taskId)
    }

    function handleCloseTaskModal() {
        setTaskAreaId(null)
        setEditingTaskId(null)
    }

    const controls = (
        <Stack gap="md">
            <Group justify="space-between" align="center" gap="md" wrap="wrap">
                <SegmentedControl
                    value={mainTab}
                    data={[
                        { label: 'Задачи', value: 'tasks' },
                        { label: 'Команды', value: 'teams' },
                        { label: 'Следующее дежурство', value: 'next-duty' },
                    ]}
                    classNames={{
                        control: segmentedControlClasses.control,
                        root: segmentedControlClasses.root,
                        label: segmentedControlClasses.label,
                    }}
                    onChange={(value) => setMainTab(value as DutySettingsMainTab)}
                />

                {mainTab === 'tasks' && (
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
                )}
            </Group>

            {mainTab === 'tasks' && viewMode === 'plan' && (
                <Group justify="flex-end">
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
        </Stack>
    )

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
    } else if (mainTab !== 'tasks') {
        content = <Box h={120} />
    } else if (viewMode === 'plan' && selectedFloorPlan) {
        const floorAreas = data.areas.filter((area) => area.floor === selectedFloorPlan.floor)
        content = (
            <Stack gap="xl">
                <Box visibleFrom="md">
                    <DutySettingsPlanPanel
                        areas={floorAreas}
                        floorPlan={selectedFloorPlan}
                        selectedAreaId={selectedAreaId}
                        onAreaClick={setSelectedAreaId}
                    />
                </Box>

                <Box hiddenFrom="md">
                    <DutySettingsAreaList
                        areas={data.areas}
                        onCreateArea={handleCreateArea}
                        onEditArea={handleEditArea}
                        onDeleteArea={setDeletingAreaId}
                        onCreateTask={handleCreateTask}
                        onEditTask={handleEditTask}
                        onDeleteTask={setDeletingTaskId}
                    />
                </Box>
            </Stack>
        )
    } else {
        content = (
            <DutySettingsAreaList
                areas={data.areas}
                onCreateArea={handleCreateArea}
                onEditArea={handleEditArea}
                onDeleteArea={setDeletingAreaId}
                onCreateTask={handleCreateTask}
                onEditTask={handleEditTask}
                onDeleteTask={setDeletingTaskId}
            />
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
            controls={controls}
        >
            {content}

            {data && (
                <>
                    <AreaFormModal
                        opened={creatingAreaOpened || editingAreaId !== null}
                        mode={editingAreaId === null ? 'create' : 'edit'}
                        areaId={editingAreaId}
                        dormitoryId={String(data.group.dormitory_id)}
                        fixedGroupId={data.group.id}
                        hideGroupField
                        requireFloor
                        loadGroups={async () => ({ groups: [] })}
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
                        createAreaRequest={async (request) => {
                            const area = await createDutySettingsArea(groupId, request)
                            appendArea({
                                id: area.id,
                                name: area.name,
                                floor: area.floor,
                                tasks: [],
                            })
                            return area
                        }}
                        updateAreaRequest={async (areaId, request) => {
                            const area = await updateDutySettingsArea(groupId, areaId, request)
                            replaceArea({
                                id: area.id,
                                name: area.name,
                                floor: area.floor,
                                tasks: data.areas.find((item) => item.id === area.id)?.tasks ?? [],
                            })
                            return area
                        }}
                        onClose={handleCloseAreaModal}
                        onSaved={async () => {}}
                    />

                    <DeleteAreaModal
                        opened={deletingArea !== null}
                        area={deletingArea ? toAreaListItem(deletingArea, data.group.id, data.group.name) : null}
                        dormitoryId={String(data.group.dormitory_id)}
                        deleteAreaRequest={(areaId) => deleteDutySettingsArea(groupId, areaId)}
                        onClose={() => setDeletingAreaId(null)}
                        onDeleted={async () => {
                            removeArea(deletingArea?.id ?? 0)
                            if (selectedAreaId === String(deletingArea?.id ?? '')) {
                                setSelectedAreaId(null)
                            }
                        }}
                    />

                    <TaskFormModal
                        opened={taskAreaId !== null}
                        mode={editingTaskId === null ? 'create' : 'edit'}
                        taskId={editingTaskId}
                        dormitoryId={String(data.group.dormitory_id)}
                        initialAreaId={taskAreaId == null ? null : String(taskAreaId)}
                        hideAreaField={editingTaskId === null}
                        loadAreas={async () => ({
                            areas: data.areas.map((area) => toAreaListItem(area, data.group.id, data.group.name)),
                        })}
                        loadTask={async (taskId) => {
                            for (const area of data.areas) {
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
                        createTaskRequest={async (request) => {
                            const created = await createDutySettingsTask(groupId, Number(taskAreaId), request)
                            appendTask(Number(taskAreaId), {
                                id: created.id,
                                title: created.title,
                                cost: created.cost,
                                frequency: created.frequency,
                                last_completed_at: null,
                            })
                            return created
                        }}
                        updateTaskRequest={async (taskId, request) => {
                            const updated = await updateDutySettingsTask(groupId, taskId, request)
                            updateTask(taskId, {
                                id: updated.id,
                                title: updated.title,
                                cost: updated.cost,
                                frequency: updated.frequency,
                                last_completed_at: data.areas
                                    .flatMap((area) => area.tasks)
                                    .find((task) => task.id === taskId)?.last_completed_at ?? null,
                            }, updated.area.id)
                            return updated
                        }}
                        onClose={handleCloseTaskModal}
                        onSaved={async () => {}}
                    />

                    <DeleteTaskModal
                        opened={deletingTaskContext !== null}
                        task={deletingTaskContext ? toTaskListItem(deletingTaskContext.task, deletingTaskContext.area) : null}
                        dormitoryId={String(data.group.dormitory_id)}
                        deleteTaskRequest={(taskId) => deleteDutySettingsTask(groupId, taskId)}
                        onClose={() => setDeletingTaskId(null)}
                        onDeleted={async () => {
                            if (deletingTaskContext) {
                                removeTask(deletingTaskContext.task.id)
                            }
                        }}
                    />

                    <DutySettingsAreaDrawer
                        area={selectedArea}
                        opened={selectedArea !== null}
                        onClose={() => setSelectedAreaId(null)}
                        onCreateTask={handleCreateTask}
                        onEditTask={handleEditTask}
                        onDeleteTask={setDeletingTaskId}
                    />
                </>
            )}
        </PageFrame>
    )
}
