import { z } from 'zod'
import { requiredEnv } from './env'
import { isTimeSpent } from './present'
import { DEFAULT_OVERVIEW_COUNT, DEFAULT_VIEWS, FIELD_TYPES, FIELD_USES, SHAPES } from './types'

export const LOG_CACHE_TAG = 'log'

const scoresSchema = z
  .record(z.string())
  .nullish()
  .transform((scores) => scores ?? {})

const rankSchema = z
  .object({ place: z.number(), of: z.number() })
  .nullish()
  .transform((rank) => rank ?? undefined)

const partsSchema = z
  .object({
    title: z.string(),
    items: z.array(
      z.object({
        kind: z.string(),
        fields: z
          .record(z.unknown())
          .nullish()
          .transform((fields) => fields ?? {}),
        scores: scoresSchema
      })
    )
  })
  .nullish()
  .transform((parts) => parts ?? undefined)

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
  order: z
    .number()
    .nullish()
    .transform((order) => order ?? undefined),
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
  prose: z
    .string()
    .nullish()
    .transform((prose) => prose ?? ''),
  proseWords: z
    .number()
    .nullish()
    .transform((words) => words ?? 0),
  takes: z
    .array(z.object({ source: z.string(), values: z.record(z.unknown()), fetchedAt: z.string() }))
    .nullish()
    .transform((takes) => takes ?? []),
  parts: partsSchema
})

const publicKindSchema = z.object({ name: z.string(), untriaged: z.number() })

export type PublicItem = z.infer<typeof publicItemSchema>

export type PublicParts = NonNullable<PublicItem['parts']>

const NOT_FOUND = 404

const UNTIL_THE_LOG_SAYS_IT_CHANGED = false

const fetchPublic = (path: string) =>
  fetch(requiredEnv('LOG_API_URL') + path, {
    next: { tags: [LOG_CACHE_TAG], revalidate: UNTIL_THE_LOG_SAYS_IT_CHANGED }
  })

const failure = (response: Response) =>
  new Error(`log api ${response.url} answered ${response.status}`)

const divisionSchema = z.object({ key: z.string(), name: z.string() })

const presentationSchema = z.object({
  axes: z.array(
    z.object({
      key: z.string(),
      name: z.string(),
      divisions: z.array(divisionSchema),
      reading: z
        .object({
          sort: z.enum(['outside', 'crowd', 'activity']),
          source: z.string(),
          field: z.string(),
          floors: z.array(z.number())
        })
        .optional()
    })
  ),
  kinds: z.record(
    z.object({
      title: z
        .string()
        .nullish()
        .transform((title) => title ?? ''),
      description: z
        .string()
        .nullish()
        .transform((description) => description ?? ''),
      statusesShown: z.array(z.string()),
      axes: z.array(z.string()),
      defaultView: z.enum(DEFAULT_VIEWS).catch('auto'),
      axesWithoutTiers: z.array(z.string()).catch([]),
      axesWithoutColumns: z.array(z.string()).catch([]),
      fieldUses: z.record(z.array(z.enum(FIELD_USES))).catch({}),
      fieldNames: z.record(z.string()).catch({}),
      overview: z
        .object({ sort: z.string(), count: z.number() })
        .catch({ sort: '', count: DEFAULT_OVERVIEW_COUNT })
    })
  ),
  fields: z
    .record(
      z.array(
        z.object({
          key: z.string(),
          name: z.string(),
          type: z.enum(FIELD_TYPES).catch('url'),
          take: z.string().optional(),
          defaultUses: z.array(z.enum(FIELD_USES))
        })
      )
    )
    .catch({}),
  pictures: z.record(z.array(z.object({ key: z.string(), shape: z.enum(SHAPES) }))).catch({}),
  tagTitles: z.record(z.string()),
  readables: z
    .array(
      z.object({ source: z.string(), field: z.string(), from: z.enum(['take', 'description']) })
    )
    .catch([]),
  kindOrder: z.array(z.string()).catch([])
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
