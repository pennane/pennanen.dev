export const SESSION_COOKIE = 'admin_session'
export const SESSION_DAYS = 90

const encoder = new TextEncoder()

const signingKey = () => {
  const secret = process.env.ADMIN_SESSION_SECRET
  if (!secret) throw new Error('ADMIN_SESSION_SECRET is not set')
  return crypto.subtle.importKey(
    'raw',
    encoder.encode(secret),
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign', 'verify']
  )
}

const toBase64Url = (bytes: ArrayBuffer) =>
  btoa(String.fromCharCode(...new Uint8Array(bytes)))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '')

const fromBase64Url = (text: string) =>
  Uint8Array.from(atob(text.replace(/-/g, '+').replace(/_/g, '/')), (c) => c.charCodeAt(0))

const payloadFor = (expiresAt: number) => encoder.encode(`admin:${expiresAt}`)

export const createSession = async (now = Date.now()) => {
  const expiresAt = now + SESSION_DAYS * 24 * 60 * 60 * 1000
  const signature = await crypto.subtle.sign('HMAC', await signingKey(), payloadFor(expiresAt))
  return { value: `${expiresAt}.${toBase64Url(signature)}`, expiresAt }
}

export const isValidSession = async (value: string | undefined, now = Date.now()) => {
  const [expiresAtText, signature] = value?.split('.') ?? []
  const expiresAt = Number(expiresAtText)
  if (!signature || !Number.isFinite(expiresAt) || expiresAt < now) {
    return false
  }
  try {
    return await crypto.subtle.verify(
      'HMAC',
      await signingKey(),
      fromBase64Url(signature),
      payloadFor(expiresAt)
    )
  } catch {
    return false
  }
}
