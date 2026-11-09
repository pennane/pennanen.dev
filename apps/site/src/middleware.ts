import { NextResponse, type NextRequest } from 'next/server'
import { isValidSession, SESSION_COOKIE } from './lib/admin/session'

const LOGIN_PATH = '/admin/login'

export async function middleware(request: NextRequest) {
  if (request.nextUrl.pathname === LOGIN_PATH) return NextResponse.next()
  if (await isValidSession(request.cookies.get(SESSION_COOKIE)?.value)) {
    return NextResponse.next()
  }
  return NextResponse.redirect(new URL(LOGIN_PATH, request.url))
}

export const config = {
  matcher: ['/admin', '/admin/:path*']
}
