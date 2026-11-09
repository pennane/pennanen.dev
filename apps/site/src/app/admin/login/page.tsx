'use client'

import { useActionState } from 'react'
import styles from '../admin.module.css'
import { logIn } from './actions'

export default function LoginPage() {
  const [state, action, pending] = useActionState(logIn, undefined)
  return (
    <form action={action} className={styles.login}>
      <h1>things as ap saw them</h1>
      <input
        name="password"
        type="password"
        autoComplete="current-password"
        placeholder="password"
        autoFocus
        required
      />
      <button type="submit" disabled={pending}>
        Enter
      </button>
      {state?.error && <p className={styles.error}>{state.error}</p>}
    </form>
  )
}
