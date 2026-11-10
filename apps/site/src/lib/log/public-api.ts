import { z } from 'zod'
import { requiredEnv } from './env'

export const LOG_CACHE_TAG = 'log'

const scoresSchema = z
  .record(z.number())
  .nullish()
  .transform((scores) => scores ?? {})

const momentSchema = z.object({
  at: z.string(),
  note: z.string(),
  scores: scoresSchema,
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
  tags: z
    .array(z.string())
    .nullish()
    .transform((tags) => tags ?? []),
  blurb: z.string(),
  updatedAt: z.string(),
  moments: z
    .array(momentSchema)
    .nullish()
    .transform((moments) => moments ?? [])
})

export type PublicItem = z.infer<typeof publicItemSchema>
export type PublicMoment = z.infer<typeof momentSchema>

const NOT_FOUND = 404

const fetchPublic = (path: string) =>
  fetch(requiredEnv('LOG_API_URL') + path, {
    cache: 'force-cache',
    next: { tags: [LOG_CACHE_TAG] }
  })

const failure = (response: Response) =>
  new Error(`log api ${response.url} answered ${response.status}`)

export const logPublic = {
  items: async (kind: string) => {
    const response = await fetchPublic(
      `/public/items?${new URLSearchParams({ kind })}`
    )
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
