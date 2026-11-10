import { createHash, timingSafeEqual } from 'crypto'
import { revalidateTag } from 'next/cache'
import { z } from 'zod'
import { requiredEnv } from '../../../lib/log/env'

const bodySchema = z.object({ tags: z.array(z.string().min(1)).min(1) })

const digest = (text: string) => createHash('sha256').update(text).digest()

const carriesSecret = (request: Request) => {
  const expected = `Bearer ${requiredEnv('REVALIDATE_SECRET')}`
  const given = request.headers.get('authorization') ?? ''
  return timingSafeEqual(digest(given), digest(expected))
}

export async function POST(request: Request) {
  if (!carriesSecret(request)) {
    return Response.json({ error: 'unauthorized' }, { status: 401 })
  }
  const body = bodySchema.safeParse(await request.json().catch(() => null))
  if (!body.success) {
    return Response.json({ error: 'expected {"tags":["…"]}' }, { status: 400 })
  }
  body.data.tags.forEach((tag) => revalidateTag(tag))
  return Response.json({ revalidated: body.data.tags })
}
