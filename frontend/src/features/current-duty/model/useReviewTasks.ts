import { useEffect, useState } from 'react'

import { selectReviewVisibleTaskIds } from './selectors'
import type { ResidentCurrentDuty, DutyTaskSelect } from './types'

type Props = {
    activeSelect: DutyTaskSelect
    duty: ResidentCurrentDuty | null
    onVerify: (taskId: string) => Promise<boolean>
    onReopen: (taskId: string) => Promise<boolean>
}

export function useReviewTasks({ activeSelect, duty, onVerify, onReopen }: Props) {
    const [reviewVisibleTaskIds, setReviewVisibleTaskIds] = useState<string[]>([])

    useEffect(() => {
        if (!duty || activeSelect !== 'review') {
            return
        }

        setReviewVisibleTaskIds(selectReviewVisibleTaskIds(duty.tasks))
    }, [activeSelect, duty])

    async function handleVerify(taskId: string) {
        return runReviewAction(taskId, onVerify)
    }

    async function handleReopen(taskId: string) {
        return runReviewAction(taskId, onReopen)
    }

    async function runReviewAction(taskId: string, action: (currentTaskId: string) => Promise<boolean>) {
        const success = await action(taskId)
        if (!success) {
            return false
        }

        setReviewVisibleTaskIds((current) => current.filter((currentTaskId) => currentTaskId !== taskId))
        return true
    }

    return {
        reviewVisibleTaskIds,
        handleReopen,
        handleVerify,
    }
}
