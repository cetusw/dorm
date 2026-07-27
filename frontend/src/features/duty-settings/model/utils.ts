import type { AreaListItem } from '../../areas/model/types'
import type { TaskListItem } from '../../task-catalog/model/types'
import type { DutySettingsArea, DutySettingsTask, DutySettingsTeamLeader } from './types'

export function formatDutySettingsFloorLabel(floor: number | null): string {
    return floor == null ? 'Без этажа' : `${floor} этаж`
}

export function formatDutySettingsAreaTitle(area: DutySettingsArea): string {
    return `${formatDutySettingsFloorLabel(area.floor)} · ${area.name}`
}

export function formatDutySettingsAreaOption(area: Pick<DutySettingsArea, 'floor' | 'name'>): string {
    return `${formatDutySettingsFloorLabel(area.floor)}. ${area.name}`
}

export function formatDutySettingsLeaderName(leader: DutySettingsTeamLeader | null): string {
    return leader?.name ?? 'Без главы'
}

export function formatLastCompletedAt(value: string | null): string {
    if (!value) {
        return '—'
    }

    const date = new Date(value)
    if (Number.isNaN(date.getTime())) {
        return '—'
    }

    return new Intl.DateTimeFormat('ru-RU', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
    }).format(date)
}

export function toAreaListItem(area: DutySettingsArea, groupId: string, groupName: string): AreaListItem {
    return {
        id: area.id,
        name: area.name,
        floor: area.floor,
        group: {
            id: groupId,
            name: groupName,
        },
    }
}

export function toTaskListItem(task: DutySettingsTask, area: DutySettingsArea): TaskListItem {
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
