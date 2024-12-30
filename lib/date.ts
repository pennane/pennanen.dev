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
		'December',
	]
	if (i > 11) return monthNames[11]
	else if (i < 0) return monthNames[0]
	return monthNames[i]
}
