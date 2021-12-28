export const isString = (text: any): text is string => typeof text === 'string' || text instanceof String
export const isNumber = (number: any): number is number => !isNaN(number)
