import type { DutyTaskSelect, ResidentCurrentDuty, ResidentDutyTask } from './types'
import { preserveTaskOrder } from './utils'

export function isSupportedMobileDutySelect(select: DutyTaskSelect): boolean {
    return select === 'mine' || select === 'all' || select === 'team'
}

export function selectTasksForMobileMine(
    duty: ResidentCurrentDuty,
    visibleMineTaskIds: string[],
): ResidentDutyTask[] {
    return preserveTaskOrder(
        duty.tasks.filter((task) => visibleMineTaskIds.includes(task.id)),
        visibleMineTaskIds,
    )
}
