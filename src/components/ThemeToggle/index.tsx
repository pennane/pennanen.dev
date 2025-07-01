'use client'

import { Moon, Sun } from 'lucide-react'
import { useEffect, useState } from 'react'
import styles from './theme-toggle.module.css'

const ICON_SIZE = 18

export const ThemeToggle = () => {
  const [theme, setTheme] = useState<'light' | 'dark'>('light')
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
    const savedTheme = localStorage.getItem('theme') as 'light' | 'dark' | null
    const systemTheme = window.matchMedia('(prefers-color-scheme: dark)')
      .matches
      ? 'dark'
      : 'light'
    const initialTheme = savedTheme || systemTheme

    setTheme(initialTheme)
    document.documentElement.setAttribute('data-color-scheme', initialTheme)
  }, [])

  const toggle = () => {
    const newTheme = theme === 'light' ? 'dark' : 'light'
    setTheme(newTheme)
    localStorage.setItem('theme', newTheme)
    document.documentElement.setAttribute('data-color-scheme', newTheme)
  }

  if (!mounted) {
    return (
      <button
        className={`${styles.toggle} ${styles.notMounted}`}
        aria-label="Toggle theme"
      >
        <Sun height={ICON_SIZE} />
      </button>
    )
  }

  return (
    <button
      className={styles.toggle}
      onClick={toggle}
      aria-label={`Switch to ${theme === 'light' ? 'dark' : 'light'} theme`}
      title={`Switch to ${theme === 'light' ? 'dark' : 'light'} theme`}
    >
      <div className={styles.iconContainer}>
        {theme === 'light' ? (
          <Sun height={ICON_SIZE} />
        ) : (
          <Moon height={ICON_SIZE} />
        )}
      </div>
    </button>
  )
}
