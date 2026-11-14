import { cookies } from 'next/headers'
import { NextResponse } from 'next/server'
import { requireAdmin } from '../../../../lib/admin/require-admin'
import { logAdmin } from '../../../../lib/log/admin-api'
import { SPOTIFY_STATE_COOKIE, SPOTIFY_STATE_MAX_AGE_SECONDS, spotifyRedirectUri } from '../state'

export const dynamic = 'force-dynamic'

export async function GET(request: Request) {
  await requireAdmin()
  const state = crypto.randomUUID()
  const url = await logAdmin.spotifyAuthorizeURL(spotifyRedirectUri(request), state)
  ;(await cookies()).set(SPOTIFY_STATE_COOKIE, state, {
    httpOnly: true,
    secure: true,
    sameSite: 'lax',
    maxAge: SPOTIFY_STATE_MAX_AGE_SECONDS,
    path: '/admin/spotify'
  })
  return NextResponse.redirect(url)
}
