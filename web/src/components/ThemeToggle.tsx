import { Button } from '@/components/ui/button'
import type { ThemePreference } from '@/hooks/useTheme'
import { Monitor, Moon, Sun } from 'lucide-react'

interface ThemeToggleProps {
  theme: ThemePreference
  onToggle: () => void
}

export const ThemeToggle = ({ theme, onToggle }: ThemeToggleProps) => {
  const label =
    theme === 'light' ? 'Light mode (click for dark)' :
    theme === 'dark' ? 'Dark mode (click for system)' :
    'System mode (click for light)'

  return (
    <Button
      variant="outline"
      size="icon"
      onClick={onToggle}
      aria-label={label}
      title={label}
    >
      {theme === 'light' && <Sun className="h-4 w-4" />}
      {theme === 'dark' && <Moon className="h-4 w-4" />}
      {theme === 'system' && <Monitor className="h-4 w-4" />}
    </Button>
  )
}
