import type { ResidentDutyTask } from './types'

export type TaskRowActionMode = 'default' | 'verification'

export type TaskActionHandler = (taskId: string) => Promise<boolean>

export type TaskActionKind = 'take' | 'return' | 'reopen' | 'verify'

export type TaskActionSpec = {
    kind: TaskActionKind
    label: string
    tone: 'default' | 'danger'
}

type TaskActionParams = {
    isReadOnly: boolean
    mode: TaskRowActionMode
    task: ResidentDutyTask
}

export function getLeftSwipeActionSpec({
    isReadOnly,
    task,
}: TaskActionParams): TaskActionSpec | null {
    if (isReadOnly) {
        return null
    }

    if (task.status === 'completed' && task.can_review_open) {
        return {
            kind: 'reopen',
            label: 'Переоткрыть',
            tone: 'danger',
        }
    }

    if (task.can_return) {
        return {
            kind: 'return',
            label: 'Вернуть',
            tone: 'default',
        }
    }

    return null
}

export function getRightSwipeActionSpec({
    isReadOnly,
    task,
}: TaskActionParams): TaskActionSpec | null {
    if (isReadOnly) {
        return null
    }

    if (task.status === 'completed' && task.can_verify) {
        return {
            kind: 'verify',
            label: 'Подтвердить',
            tone: 'default',
        }
    }

    if (task.can_take) {
        return {
            kind: 'take',
            label: 'Взять',
            tone: 'default',
        }
    }

    return null
}

export function getDrawerActionSpecs({
    isReadOnly,
    task,
}: TaskActionParams): TaskActionSpec[] {
    if (isReadOnly) {
        return []
    }

    if (task.status === 'completed' && (task.can_review_open || task.can_verify)) {
        const actions: TaskActionSpec[] = []

        if (task.can_review_open) {
            actions.push({
                kind: 'reopen',
                label: 'Переоткрыть',
                tone: 'danger',
            })
        }

        if (task.can_verify) {
            actions.push({
                kind: 'verify',
                label: 'Подтвердить',
                tone: 'default',
            })
        }

        return actions
    }

    if (task.can_take) {
        return [{
            kind: 'take',
            label: 'Взять',
            tone: 'default',
        }]
    }

    if (task.can_return) {
        return [{
            kind: 'return',
            label: 'Вернуть',
            tone: 'default',
        }]
    }

    return []
}
