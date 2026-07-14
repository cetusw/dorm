import { Alert, Box, Center, Loader } from '@mantine/core'

import {
    calculateDutyAnalytics,
    selectDutyViewOptions,
    selectTasksForTab,
} from '../../features/current-duty/model/selectors'
import { useCurrentDuty } from '../../features/current-duty/model/useCurrentDuty'
import { useReviewTasks } from '../../features/current-duty/model/useReviewTasks'
import { useStoredDutyTab } from '../../features/current-duty/model/useStoredDutyTab'
import { formatDutyPeriod } from '../../features/current-duty/model/utils'
import { CurrentDutyAnalytics } from '../../features/current-duty/ui/CurrentDutyAnalytics'
import { DutyTaskSelects } from '../../features/current-duty/ui/DutyTaskSelects'
import { TaskGroups } from '../../features/current-duty/ui/TaskGroups'
import { PageFrame } from '../../shared/ui/PageFrame'

export function CurrentDutyPage() {
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
        handleVerify,
        visibleMineTaskIds,
        visibleFreeTaskIds,
    } = useCurrentDuty()
    const [activeTab, setActiveTab] = useStoredDutyTab(duty?.visible_tabs ?? [])
    const { reviewVisibleTaskIds, handleVerify: handleReviewVerify } = useReviewTasks({
        activeTab,
        duty,
        onVerify: handleVerify,
    })

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

    const viewOptions = selectDutyViewOptions(duty, activeTab)
    const displayedTasks = selectTasksForTab({
        activeTab,
        duty,
        reviewVisibleTaskIds,
        visibleFreeTaskIds,
        visibleMineTaskIds,
    })
    const analytics = calculateDutyAnalytics(duty.tasks)

    const controls = viewOptions.showControls ? (
        <DutyTaskSelects
            activeSelect={activeTab}
            groups={duty.groups}
            selectedGroupId={selectedGroupId ?? duty.selected_group_id}
            showGroupSelect={duty.show_group_select}
            visibleSelects={duty.visible_tabs}
            onGroupChange={selectGroup}
            onChange={setActiveTab}
        />
    ) : undefined

    const analyticsBlock = viewOptions.showAnalytics || viewOptions.isReadOnly ? (
        <CurrentDutyAnalytics
            analytics={analytics}
            isReadOnly={viewOptions.isReadOnly}
            targetValue={duty.cost_per_resident_goal}
        />
    ) : undefined

    if (!duty.has_active_duty) {
        return (
            <PageFrame title="Дежурство" error={error} controls={controls}>
                <Alert color="gray">В выбранной группе сейчас нет активного дежурства.</Alert>
            </PageFrame>
        )
    }

    if (duty.tasks.length === 0) {
        return (
            <PageFrame
                title="Дежурство"
                subtitle={formatDutyPeriod(duty.start_date, duty.end_date)}
                error={error}
                notice={duty.notice_message}
                controls={controls}
            >
                <Alert color="gray">На текущее дежурство не заведены задачи.</Alert>
            </PageFrame>
        )
    }

    return (
        <PageFrame
            title="Дежурство"
            subtitle={formatDutyPeriod(duty.start_date, duty.end_date)}
            error={error}
            notice={duty.notice_message}
            analytics={analyticsBlock}
            controls={controls}
        >
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
                onVerify={handleReviewVerify}
            />
        </PageFrame>
    )
}
