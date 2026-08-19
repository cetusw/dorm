import type { DutyTaskSelect, ResidentCurrentDuty, ResidentDutyTask } from './types'

export function isSupportedMobileDutySelect(select: DutyTaskSelect): boolean {
    return select === 'mine' || select === 'all' || select === 'team'
}

export function selectTasksForMobileMine(duty: ResidentCurrentDuty): ResidentDutyTask[] {
    return duty.tasks.filter((task) => task.is_mine)
}
