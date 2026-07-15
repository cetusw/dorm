const transliterationMap: Record<string, string> = {
    а: 'a',
    б: 'b',
    в: 'v',
    г: 'g',
    д: 'd',
    е: 'e',
    ё: 'e',
    ж: 'zh',
    з: 'z',
    и: 'i',
    й: 'i',
    к: 'k',
    л: 'l',
    м: 'm',
    н: 'n',
    о: 'o',
    п: 'p',
    р: 'r',
    с: 's',
    т: 't',
    у: 'u',
    ф: 'f',
    х: 'kh',
    ц: 'ts',
    ч: 'ch',
    ш: 'sh',
    щ: 'shch',
    ъ: '',
    ы: 'y',
    ь: '',
    э: 'e',
    ю: 'yu',
    я: 'ya',
}

function transliteratePart(value: string): string {
    return value
        .trim()
        .toLowerCase()
        .split('')
        .map((character) => transliterationMap[character] ?? character)
        .join('')
        .replace(/[^a-z0-9\s-]/g, '')
        .replace(/[\s-]+/g, '-')
        .replace(/^-+|-+$/g, '')
}

export function generateResidentLogin(firstName: string, lastName: string): string {
    const firstNamePart = transliteratePart(firstName)
    const lastNamePart = transliteratePart(lastName)

    return [firstNamePart, lastNamePart].filter(Boolean).join('.')
}

export function generateResidentPassword(): string {
    const alphabet = 'abcdefghijklmnopqrstuvwxyz0123456789'
    let password = ''

    for (let index = 0; index < 6; index += 1) {
        const nextIndex = Math.floor(Math.random() * alphabet.length)
        password += alphabet[nextIndex]
    }

    return password
}
