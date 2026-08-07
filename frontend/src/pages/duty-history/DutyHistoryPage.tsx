import { CaretLeftIcon } from '@phosphor-icons/react'
import { Box, Center, Loader, Select, Stack, Table, Text } from '@mantine/core'

import { navigateTo } from '../../app/navigation'
import { buildReadonlyDutyProgressModel, type DutyAnalytics } from '../../features/current-duty/model/selectors'
import { formatDutyPeriodFull } from '../../features/current-duty/model/utils'
import { MemberDutyProgressCard } from '../../features/current-duty/ui/MemberDutyProgressCard'
import { useDutyHistory } from '../../features/duty-history/model/useDutyHistory'
import { AppTable, appTableClasses } from '../../shared/ui/AppTable'
import { EmptyState } from '../../shared/ui/EmptyState'
import { PageFrame } from '../../shared/ui/PageFrame'
import classes from './DutyHistoryPage.module.css'

function toDutyAnalytics(total: number, taken: number, completed: number, verified: number): DutyAnalytics {
    return {
        myVerifiedTasksCount: 0,
        totalTasksCount: total,
        totalTakenTasksCount: taken,
        totalCompletedTasksCount: completed,
        totalVerifiedTasksCount: verified,
        takenCostSum: 0,
        completedCostSum: 0,
        verifiedCostSum: 0,
        takenTasksCount: 0,
        completedTasksCount: 0,
    }
}

function ProgressTooltip({
    takenCost,
    completed,
    verified,
    totalCost,
    totalTasks,
}: {
    takenCost: number
    completed: number
    verified: number
    totalCost: number
    totalTasks: number
}) {
    return (
        <Stack gap={2}>
            <Text size="sm">Взято {takenCost} из {totalCost} баллов</Text>
            <Text size="sm">Выполнено {completed} из {totalTasks} задач</Text>
            <Text size="sm">Проверено {verified} из {totalTasks} задач</Text>
        </Stack>
    )
}

export function DutyHistoryPage() {
    const params = new URLSearchParams(window.location.search)
    const initialGroupId = params.get('group_id') ?? undefined
    const { data, loading, error, selectGroup } = useDutyHistory(initialGroupId)

    if (loading) {
        return <Center py="xl"><Loader /></Center>
    }

    if (error || !data) {
        return (
            <PageFrame title="История дежурств" error={error ?? 'Не удалось загрузить историю'}>
                <EmptyState
                    title="История не найдена"
                    description="Не удалось открыть историю дежурств."
                />
            </PageFrame>
        )
    }

    return (
        <PageFrame
            topContent={(
                <button
                    type="button"
                    onClick={() => navigateTo(`/app/tasks?group_id=${encodeURIComponent(data.selected_group_id)}`)}
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
            title="История дежурств"
            controls={data.show_group_select ? (
                <Select
                    aria-label="Группа"
                    data={data.groups.map((group) => ({ value: group.id, label: group.name }))}
                    value={data.selected_group_id}
                    allowDeselect={false}
                    onChange={(value) => {
                        if (value) {
                            selectGroup(value)
                        }
                    }}
                />
            ) : undefined}
        >
            <AppTable minWidth={640}>
                <Table.Thead>
                    <Table.Tr>
                        <Table.Th>Период</Table.Th>
                        <Table.Th>Команда</Table.Th>
                        <Table.Th w={320}>Прогресс</Table.Th>
                    </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                    {data.duties.map((item) => (
                        <Table.Tr
                            key={item.id}
                            className={`${appTableClasses.interactiveRow} ${appTableClasses.compactRow}`}
                            tabIndex={0}
                            onClick={() => navigateTo(`/app/duties/${item.id}?group_id=${encodeURIComponent(data.selected_group_id)}`)}
                            onKeyDown={(event) => {
                                if (event.key === 'Enter' || event.key === ' ') {
                                    event.preventDefault()
                                    navigateTo(`/app/duties/${item.id}?group_id=${encodeURIComponent(data.selected_group_id)}`)
                                }
                            }}
                        >
                            <Table.Td>{formatDutyPeriodFull(item.start_date, item.end_date)}</Table.Td>
                            <Table.Td>{item.team_leader_name}</Table.Td>
                            <Table.Td className={classes.progressCell}>
                                <Box maw={288}>
                                    <MemberDutyProgressCard
                                        progress={buildReadonlyDutyProgressModel(toDutyAnalytics(
                                            item.progress.total_tasks_count,
                                            item.progress.taken_tasks_count,
                                            item.progress.completed_tasks_count,
                                            item.progress.verified_tasks_count,
                                        ))}
                                        showTitle={false}
                                        withContainer={false}
                                        tooltip={(
                                            <ProgressTooltip
                                                takenCost={item.progress.taken_cost_sum}
                                                completed={item.progress.completed_tasks_count}
                                                verified={item.progress.verified_tasks_count}
                                                totalCost={item.progress.total_cost_sum}
                                                totalTasks={item.progress.total_tasks_count}
                                            />
                                        )}
                                    />
                                </Box>
                            </Table.Td>
                        </Table.Tr>
                    ))}
                </Table.Tbody>
            </AppTable>
        </PageFrame>
    )
}
