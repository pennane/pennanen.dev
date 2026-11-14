export const SPOTIFY_STATE_COOKIE = 'spotify-state'
export const SPOTIFY_STATE_MAX_AGE_SECONDS = 10 * 60

export const spotifyRedirectUri = (request: Request) =>
  `${new URL(request.url).origin}/admin/spotify/callback`

export const settingsRedirect = (request: Request, outcome: string) =>
  Response.redirect(new URL(`/admin/settings?spotify=${encodeURIComponent(outcome)}`, request.url))
