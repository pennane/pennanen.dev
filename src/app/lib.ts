export function parseDateString(date: unknown) {
  if (!date) return null
  if (typeof date === 'number') return new Date(date)
  if (typeof date !== 'string') return undefined
  const [day, month, year] = date.split('.')
  if (!year) return null
  return new Date(parseInt(year), parseInt(month) - 1, parseInt(day))
}

const formatter = new Intl.DateTimeFormat('en-GB', {
  year: 'numeric',
  month: 'short',
  day: 'numeric'
})

export function formatDate(date: number | Date | undefined | null) {
  if (!date) return ''
  return formatter.format(date)
}
