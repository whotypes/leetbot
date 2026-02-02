import { useEffect, useState } from 'react'

export type ThemePreference = 'light' | 'dark' | 'system'

const getStoredTheme = (): ThemePreference => {
  const stored = localStorage.getItem('theme')
  if (stored === 'light' || stored === 'dark') {
    return stored
  }
  return 'system'
}

const getSystemTheme = (): 'light' | 'dark' => {
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

const applyTheme = (preference: ThemePreference) => {
  const root = document.documentElement
  root.classList.remove('light', 'dark')
  const effectiveTheme = preference === 'system' ? getSystemTheme() : preference
  root.classList.add(effectiveTheme)
}

export const useTheme = () => {
  const [theme, setTheme] = useState<ThemePreference>(getStoredTheme)

  useEffect(() => {
    if (theme !== 'system') return

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')

    const handleChange = () => {
      applyTheme('system')
    }

    mediaQuery.addEventListener('change', handleChange)
    return () => mediaQuery.removeEventListener('change', handleChange)
  }, [theme])

  const toggleTheme = () => {
    const nextTheme: ThemePreference =
      theme === 'light' ? 'dark' :
      theme === 'dark' ? 'system' :
      'light'

    if (nextTheme === 'system') {
      localStorage.removeItem('theme')
    } else {
      localStorage.setItem('theme', nextTheme)
    }

    applyTheme(nextTheme)
    setTheme(nextTheme)
  }

  return { theme, toggleTheme }
}
