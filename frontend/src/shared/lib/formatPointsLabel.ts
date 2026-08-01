function getRussianPluralForm(
    value: number,
    one: string,
    few: string,
    many: string,
): string {
    const absoluteValue = Math.abs(value)
    const lastTwoDigits = absoluteValue % 100

    if (lastTwoDigits >= 11 && lastTwoDigits <= 14) {
        return many
    }

    const lastDigit = absoluteValue % 10

    if (lastDigit === 1) {
        return one
    }

    if (lastDigit >= 2 && lastDigit <= 4) {
        return few
    }

    return many
}

export function formatPointsLabel(value: number): string {
    return `${value} ${getRussianPluralForm(value, 'балл', 'балла', 'баллов')}`
}
