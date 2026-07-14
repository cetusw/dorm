import { useEffect, useState } from 'react'

import { selectReviewVisibleTaskIds } from './selectors'
import type { ResidentCurrentDuty, DutyTaskSelect } from './types'

type Props = {
    activeSelect: DutyTaskSelect
    duty: ResidentCurrentDuty | null
    onVerify: (taskId: string) => Promise<boolean>
}

export function useReviewTasks({ activeSelect, duty, onVerify }: Props) {
    const [reviewVisibleTaskIds, setReviewVisibleTaskIds] = useState<string[]>([])

    useEffect(() => {
        if (!duty || activeSelect !== 'review') {
            return
        }

        setReviewVisibleTaskIds(selectReviewVisibleTaskIds(duty.tasks))
    }, [activeSelect, duty])

    async function handleVerify(taskId: string) {
        const success = await onVerify(taskId)
        if (!success) {
            return false
        }

        setReviewVisibleTaskIds((current) => current.filter((currentTaskId) => currentTaskId !== taskId))
        return true
    }

    return {
        reviewVisibleTaskIds,
        handleVerify,
    }
}
