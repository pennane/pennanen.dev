import { cookies } from 'next/headers'
import { requireAdmin } from '../../../../lib/admin/require-admin'
import { LogApiError, logAdmin } from '../../../../lib/log/admin-api'
import { SPOTIFY_STATE_COOKIE, settingsRedirect, spotifyRedirectUri } from '../state'

export const dynamic = 'force-dynamic'

export async function GET(request: Request) {
  await requireAdmin()
  const jar = await cookies()
  const expectedState = jar.get(SPOTIFY_STATE_COOKIE)?.value
  jar.delete({ name: SPOTIFY_STATE_COOKIE, path: '/admin/spotify' })

  const params = new URL(request.url).searchParams
  if (params.get('error')) return settingsRedirect(request, 'refused')

  const code = params.get('code')
  if (!expectedState || params.get('state') !== expectedState || !code) {
    return settingsRedirect(request, 'failed')
  }

  try {
    await logAdmin.connectSpotify(code, spotifyRedirectUri(request))
  } catch (error) {
    console.error('spotify connect failed', error)
    const rejected = error instanceof LogApiError && error.status === 400
    return settingsRedirect(request, rejected ? 'refused' : 'failed')
  }
  return settingsRedirect(request, 'connected')
}
