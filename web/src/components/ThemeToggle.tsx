import { Button } from '@/components/ui/button'
import { Moon, Sun } from 'lucide-react'
import { memo } from 'react'

interface ThemeToggleProps {
  theme: 'light' | 'dark'
  onToggle: () => void
}

export const ThemeToggle = memo(({ theme, onToggle }: ThemeToggleProps) => {
  const handleClick = () => {
    onToggle()
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLButtonElement>) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      onToggle()
    }
  }

  return (
    <Button
      variant="outline"
      size="icon"
      onClick={handleClick}
      onKeyDown={handleKeyDown}
      aria-label={`Switch to ${theme === 'light' ? 'dark' : 'light'} mode`}
    >
      {theme === 'light' ? (
        <Moon className="h-4 w-4" />
      ) : (
        <Sun className="h-4 w-4" />
      )}
    </Button>
  )
})
