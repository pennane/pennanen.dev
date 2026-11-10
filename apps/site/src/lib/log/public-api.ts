import { z } from 'zod'
import { requiredEnv } from './env'
import { isTimeSpent } from './present'

export const LOG_CACHE_TAG = 'log'

const scoresSchema = z
  .record(z.string())
  .nullish()
  .transform((scores) => scores ?? {})

const rankSchema = z
  .object({ place: z.number(), of: z.number() })
  .nullish()
  .transform((rank) => rank ?? undefined)

const momentSchema = z.object({
  at: z.string(),
  note: z.string(),
  refs: z
    .array(z.object({ kind: z.string(), slug: z.string(), title: z.string() }))
    .nullish()
    .transform((refs) => refs ?? [])
})

const publicItemSchema = z.object({
  kind: z.string(),
  slug: z.string(),
  fields: z
    .record(z.unknown())
    .nullish()
    .transform((fields) => fields ?? {}),
  status: z.string(),
  scores: scoresSchema,
  rank: rankSchema,
  tags: z
    .array(z.string())
    .nullish()
    .transform((tags) => tags ?? []),
  blurb: z.string(),
  timeSpent: z
    .string()
    .nullish()
    .transform((value) => (isTimeSpent(value) ? value : undefined)),
  updatedAt: z.string(),
  moments: z
    .array(momentSchema)
    .nullish()
    .transform((moments) => moments ?? [])
})

const publicKindSchema = z.object({ name: z.string(), untriaged: z.number() })

export type PublicItem = z.infer<typeof publicItemSchema>
export type PublicMoment = z.infer<typeof momentSchema>

const NOT_FOUND = 404

const STALE_AFTER_SECONDS = 300

const fetchPublic = (path: string) =>
  fetch(requiredEnv('LOG_API_URL') + path, {
    next: { tags: [LOG_CACHE_TAG], revalidate: STALE_AFTER_SECONDS }
  })

const failure = (response: Response) =>
  new Error(`log api ${response.url} answered ${response.status}`)

const divisionSchema = z.object({ key: z.string(), name: z.string() })

const presentationSchema = z.object({
  axes: z.array(
    z.object({ key: z.string(), name: z.string(), divisions: z.array(divisionSchema) })
  ),
  kinds: z.record(
    z.object({
      statusesShown: z.array(z.string()),
      axes: z.array(z.string())
    })
  ),
  tagTitles: z.record(z.string())
})

export const logPublic = {
  presentation: async () => {
    const response = await fetchPublic('/public/presentation')
    if (!response.ok) throw failure(response)
    return presentationSchema.parse(await response.json())
  },
  kind: async (kind: string) => {
    const response = await fetchPublic(`/public/kinds/${encodeURIComponent(kind)}`)
    if (!response.ok) throw failure(response)
    return publicKindSchema.parse(await response.json())
  },
  items: async (kind: string) => {
    const response = await fetchPublic(`/public/items?${new URLSearchParams({ kind })}`)
    if (!response.ok) throw failure(response)
    return z.array(publicItemSchema).parse(await response.json())
  },
  item: async (kind: string, slug: string) => {
    const response = await fetchPublic(
      `/public/items/${encodeURIComponent(kind)}/${encodeURIComponent(slug)}`
    )
    if (response.status === NOT_FOUND) return undefined
    if (!response.ok) throw failure(response)
    return publicItemSchema.parse(await response.json())
  }
}
