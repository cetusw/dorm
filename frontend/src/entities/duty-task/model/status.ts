export type DutyTaskStatus = 'free' | 'assigned' | 'completed' | 'verified'

export type DutyTaskStatusConfig = {
    label: string
    textColor: string
    backgroundColor: string
}

const DUTY_TASK_STATUS_CONFIG: Record<DutyTaskStatus, DutyTaskStatusConfig> = {
    free: {
        label: 'Свободна',
        textColor: '#000000',
        backgroundColor: '#EEF2F1',
    },
    assigned: {
        label: 'Взята',
        textColor: '#0369A1',
        backgroundColor: '#E0F2FE',
    },
    completed: {
        label: 'На проверке',
        textColor: '#92400E',
        backgroundColor: '#FEF3C7',
    },
    verified: {
        label: 'Проверена',
        textColor: '#166534',
        backgroundColor: '#DCFCE7',
    },
}

export function getDutyTaskStatusConfig(status: DutyTaskStatus): DutyTaskStatusConfig {
    return DUTY_TASK_STATUS_CONFIG[status]
}
