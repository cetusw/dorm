import { useMemo, useState } from 'react'

import { CaretLeftIcon } from '@phosphor-icons/react'
import { Alert, Box, Center, Loader, SegmentedControl, Stack } from '@mantine/core'

import {
    createDutySettingsTask,
    excludeDutySettingsTask,
    getDutySettings,
    includeDutySettingsTask,
    updateDutySettingsTask,
} from '../../features/duty-settings/api/dutySettingsApi'
import { useDutySettings } from '../../features/duty-settings/model/useDutySettings'
import type { DutySettingsMainTab } from '../../features/duty-settings/model/types'
import { DutySettingsAreaList } from '../../features/duty-settings/ui/DutySettingsAreaList'
import { DutySettingsTaskSummary } from '../../features/duty-settings/ui/DutySettingsTaskSummary'
import { DutySettingsTeamsTab } from '../../widgets/duty-settings-teams/ui/DutySettingsTeamsTab'
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
    const [editingTaskId, setEditingTaskId] = useState<string | null>(null)
    const [taskAreaId, setTaskAreaId] = useState<number | null>(null)
    const [excludingTaskId, setExcludingTaskId] = useState<string | null>(null)

    const excludingTask = useMemo(() => {
        if (!data || !excludingTaskId) {
            return null
        }

        for (const area of data.areas) {
            const task = area.tasks.find((item) => item.id === excludingTaskId)
            if (task) {
                return task
            }
        }

        return null
    }, [data, excludingTaskId])

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
        content = (
            <Stack gap="xl">
                <DutySettingsTaskSummary summary={data.active_duty.summary} />
                <DutySettingsAreaList
                    areas={data.areas}
                    onCreateTask={handleCreateTask}
                    onEditTask={handleEditTask}
                    onIncludeTask={async (taskId) => {
                        await includeDutySettingsTask(groupId, taskId)
                        await reload({ silent: true })
                    }}
                    onExcludeTask={(taskId) => setExcludingTaskId(taskId)}
                />
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
            )}
        >
            {content}

            {data ? (
                <>
                    <TaskFormModal
                        opened={taskAreaId !== null}
                        mode={editingTaskId === null ? 'create' : 'edit'}
                        taskId={editingTaskId}
                        dormitoryId={String(data.group.dormitory_id)}
                        initialAreaId={taskAreaId == null ? null : String(taskAreaId)}
                        hideAreaField={editingTaskId === null}
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
                        createTaskRequest={async (request) => {
                            const created = await createDutySettingsTask(groupId, Number(taskAreaId), request)
                            await reload({ silent: true })
                            return created
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
                </>
            ) : null}
        </PageFrame>
    )
}
