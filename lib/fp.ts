export const isString = (text: unknown): text is string =>
	typeof text === 'string' || text instanceof String

export const isNumber = (number: unknown): number is number =>
	typeof number === 'number' && !isNaN(number)
