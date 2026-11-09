'use server'

import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'
import { createHash, timingSafeEqual } from 'node:crypto'
import { createSession, SESSION_COOKIE } from '../../../lib/admin/session'

export type LoginState = { error: string } | undefined

const digest = (text: string) => createHash('sha256').update(text).digest()

const isAdminPassword = (attempt: string) => {
  const password = process.env.ADMIN_PASSWORD
  if (!password) return false
  return timingSafeEqual(digest(attempt), digest(password))
}

export async function logIn(
  _previous: LoginState,
  form: FormData
): Promise<LoginState> {
  if (!isAdminPassword(String(form.get('password') ?? ''))) {
    return { error: 'Nope.' }
  }
  const session = await createSession()
  ;(await cookies()).set(SESSION_COOKIE, session.value, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'strict',
    path: '/admin',
    expires: session.expiresAt
  })
  redirect('/admin/triage')
}

export async function logOut() {
  ;(await cookies()).delete({ name: SESSION_COOKIE, path: '/admin' })
  redirect('/admin/login')
}
