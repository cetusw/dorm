export function formatRedemptionWeight(weight: number): string {
    const normalized = Number.isInteger(weight) ? String(weight) : String(weight).replace('.', ',')
    const absolute = Math.abs(weight)
    if (!Number.isInteger(absolute)) return `${normalized} предупреждения`
    const remainder100 = absolute % 100
    const remainder10 = absolute % 10
    if (remainder100 >= 11 && remainder100 <= 14) return `${normalized} предупреждений`
    if (remainder10 === 1) return `${normalized} предупреждение`
    if (remainder10 >= 2 && remainder10 <= 4) return `${normalized} предупреждения`
    return `${normalized} предупреждений`
}

export function formatDeadline(value: string): string {
    const [year, month, day] = value.split('-')
    return year && month && day ? `${day}.${month}.${year}` : value
}

export function formatArea(area: { name: string; floor: number | null }): string {
    return area.floor === null ? area.name : `${area.floor} этаж. ${area.name}`
}
