// noinspection JSUnusedGlobalSymbols

export const isString = (text: any): text is string => typeof text === 'string' || text instanceof String
export const isNumber = (number: any): number is number => !isNaN(number)
export const dateToFinnishLocale = (date: Date) => {
    return date.toLocaleDateString('fi-Fi', {
        timeZone: 'Europe/Helsinki'
    })
}
export const monthIndexToName = (number: number) => {
    const i = Math.round(number)
    const monthNames = [
        'January',
        'February',
        'March',
        'April',
        'May',
        'June',
        'July',
        'August',
        'September',
        'October',
        'November',
        'December'
    ]
    if (i > 11) return monthNames[11]
    else if (i < 0) return monthNames[0]
    return monthNames[i]
}
