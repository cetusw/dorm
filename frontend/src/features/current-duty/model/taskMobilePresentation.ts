import type { ResidentDutyTask } from './types'
import type { TaskRowActionMode } from './taskActions'

export type MobileTaskCardPresentation = {
    showCheckbox: boolean
    showStatus: boolean
    showAssignee: boolean
    assigneeLabel: string | null
}

type Params = {
    isReadOnly: boolean
    mode: TaskRowActionMode
    task: ResidentDutyTask
}

function getAssigneeLabel(task: ResidentDutyTask): string | null {
    if (!task.assignee_name) {
        return null
    }

    return task.is_mine ? 'Вы' : task.assignee_name
}

export function getMobileTaskCardPresentation({
    isReadOnly,
    mode,
    task,
}: Params): MobileTaskCardPresentation {
    const showCheckbox = !isReadOnly && task.is_mine && (task.can_complete || task.can_open)
    const assigneeLabel = getAssigneeLabel(task)
    const showVerificationMeta = mode === 'verification' && task.status === 'completed' && (task.can_verify || task.can_review_open)

    if (task.status === 'free') {
        return {
            showCheckbox,
            showStatus: false,
            showAssignee: false,
            assigneeLabel: null,
        }
    }

    if (showVerificationMeta) {
        return {
            showCheckbox,
            showStatus: true,
            showAssignee: Boolean(assigneeLabel),
            assigneeLabel,
        }
    }

    if (task.is_mine && task.status !== 'verified') {
        return {
            showCheckbox,
            showStatus: false,
            showAssignee: false,
            assigneeLabel: null,
        }
    }

    return {
        showCheckbox,
        showStatus: true,
        showAssignee: !task.is_mine && Boolean(assigneeLabel),
        assigneeLabel,
    }
}
