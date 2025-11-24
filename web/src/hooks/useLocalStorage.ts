import { useState } from 'react'

type SetValue<T> = T | ((val: T) => T)

function useLocalStorage<T>(
  key: string,
  initialValue: T,
): [T, (value: SetValue<T>) => void] {
  // State to store our value
  // Pass initial state function to useState so logic is only executed once
  const [storedValue, setStoredValue] = useState<T>(() => {
    if (typeof window === 'undefined') {
      return initialValue
    }
    try {
      // Get from local storage by key
      const item = window.localStorage.getItem(key)
      // Parse stored json or if none return initialValue
      return item ? JSON.parse(item) : initialValue
    } catch (error) {
      // If error also return initialValue
      console.warn(`Error reading localStorage key "${key}":`, error)
      return initialValue
    }
  })

  // Return a wrapped version of useState's setter function that persists the new value to localStorage.
  const setValue = (value: SetValue<T>) => {
    try {
      // Allow value to be a function so we have the same API as useState
      const valueToStore = value instanceof Function ? value(storedValue) : value
      // Save state immediately for responsive UI
      setStoredValue(valueToStore)
      // Defer localStorage write to avoid blocking the main thread
      if (typeof window !== 'undefined') {
        if ('requestIdleCallback' in window) {
          requestIdleCallback(() => {
            window.localStorage.setItem(key, JSON.stringify(valueToStore))
          })
        } else {
          setTimeout(() => {
            window.localStorage.setItem(key, JSON.stringify(valueToStore))
          }, 0)
        }
      }
    } catch (error) {
      // A more advanced implementation would handle the error case
      console.warn(`Error setting localStorage key "${key}":`, error)
    }
  }

  return [storedValue, setValue]
}

export { useLocalStorage }
