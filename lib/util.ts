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

export const getAbsoluteURL = (path: string) => {
    const baseURL = process.env.VERCEL_URL ? `https://${process.env.VERCEL_URL}` : 'http://localhost:3000'
    return baseURL + path
}

export function wrapText(context: any, text: string, x: number, y: number, maxWidth: number, lineHeight: number) {
    const words = text.split(' ')
    let line = ''

    for (var n = 0; n < words.length; n++) {
        let testLine = line + words[n] + ' '
        const metrics = context.measureText(testLine)
        const testWidth = metrics.width
        if (testWidth > maxWidth && n > 0) {
            context.fillText(line, x, y)
            line = words[n] + ' '
            y += lineHeight
        } else {
            line = testLine
        }
    }
    context.fillText(line, x, y)
    return { x, y }
}
