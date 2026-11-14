import { cookies } from 'next/headers'
import { requireAdmin } from '../../../../lib/admin/require-admin'
import { logAdmin } from '../../../../lib/log/admin-api'
import { SPOTIFY_STATE_COOKIE, settingsRedirect, spotifyRedirectUri } from '../state'

export const dynamic = 'force-dynamic'

export async function GET(request: Request) {
  await requireAdmin()
  const jar = await cookies()
  const expectedState = jar.get(SPOTIFY_STATE_COOKIE)?.value
  jar.delete({ name: SPOTIFY_STATE_COOKIE, path: '/admin/spotify' })

  const params = new URL(request.url).searchParams
  const refusal = params.get('error')
  if (refusal) return settingsRedirect(request, refusal)

  const code = params.get('code')
  if (!expectedState || params.get('state') !== expectedState || !code) {
    return settingsRedirect(request, 'state-mismatch')
  }

  try {
    await logAdmin.connectSpotify(code, spotifyRedirectUri(request))
  } catch (error) {
    return settingsRedirect(request, error instanceof Error ? error.message : 'failed')
  }
  return settingsRedirect(request, 'connected')
}
