function getRussianPlural(count: number, forms: [string, string, string]): string {
    const mod100 = count % 100

    if (mod100 >= 11 && mod100 <= 14) {
        return forms[2]
    }

    const mod10 = count % 10

    if (mod10 === 1) {
        return forms[0]
    }

    if (mod10 >= 2 && mod10 <= 4) {
        return forms[1]
    }

    return forms[2]
}

export function formatEstimatedDuration(cost: number): string {
    const totalMinutes = Math.max(0, Math.round(cost * 5))
    const hours = Math.floor(totalMinutes / 60)
    const minutes = totalMinutes % 60

    if (hours === 0) {
        return `≈ ${totalMinutes} ${getRussianPlural(totalMinutes, ['минута', 'минуты', 'минут'])}`
    }

    const hourPart = `${hours} ${getRussianPlural(hours, ['час', 'часа', 'часов'])}`

    if (minutes === 0) {
        return `≈ ${hourPart}`
    }

    return `≈ ${hourPart} ${minutes} ${getRussianPlural(minutes, ['минута', 'минуты', 'минут'])}`
}
