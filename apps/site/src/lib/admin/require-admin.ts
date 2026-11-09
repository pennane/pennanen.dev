import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'
import { isValidSession, SESSION_COOKIE } from './session'

export const requireAdmin = async () => {
  const session = (await cookies()).get(SESSION_COOKIE)?.value
  if (!(await isValidSession(session))) redirect('/admin/login')
}
