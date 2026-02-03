import { Label } from '@/components/ui/label'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'
import { Check, ChevronDown } from 'lucide-react'
import * as React from 'react'

interface TimeframeSelectorProps {
  timeframes: string[]
  selectedTimeframe: string
  onTimeframeChange: (timeframe: string) => void
  disabled: boolean
  compact?: boolean
}

interface MobileTimeframeSelectorProps {
  timeframes: string[]
  selectedTimeframe: string
  onTimeframeChange: (timeframe: string) => void
  disabled?: boolean
}

const timeframeLabels: Record<string, string> = {
  'all': 'All Time',
  'thirty-days': 'Last 30 Days',
  'three-months': 'Last 3 Months',
  'six-months': 'Last 6 Months',
  'more-than-six-months': 'More than 6 Months'
}

const shortTimeframeLabels: Record<string, string> = {
  'all': 'All Time',
  'thirty-days': '30 Days',
  'three-months': '3 Months',
  'six-months': '6 Months',
  'more-than-six-months': '6+ Months'
}

export const MobileTimeframeSelector = ({
  timeframes,
  selectedTimeframe,
  onTimeframeChange,
  disabled = false,
}: MobileTimeframeSelectorProps) => {
  const [isOpen, setIsOpen] = React.useState(false)
  const triggerRef = React.useRef<HTMLButtonElement>(null)
  const dropdownRef = React.useRef<HTMLDivElement>(null)

  const handleSelect = (value: string) => {
    onTimeframeChange(value)
    setIsOpen(false)
  }

  React.useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(e.target as Node) &&
        triggerRef.current &&
        !triggerRef.current.contains(e.target as Node)
      ) {
        setIsOpen(false)
      }
    }

    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setIsOpen(false)
        triggerRef.current?.focus()
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside)
      document.addEventListener('keydown', handleEscape)
      return () => {
        document.removeEventListener('mousedown', handleClickOutside)
        document.removeEventListener('keydown', handleEscape)
      }
    }
  }, [isOpen])

  const selectedLabel = shortTimeframeLabels[selectedTimeframe] || selectedTimeframe || 'Timeframe'

  return (
    <div className="relative">
      <button
        ref={triggerRef}
        type="button"
        role="combobox"
        aria-expanded={isOpen}
        aria-haspopup="listbox"
        aria-label="Select timeframe"
        disabled={disabled}
        className={cn(
          'inline-flex h-8 items-center gap-1.5 rounded-md border border-input bg-background px-2.5 text-sm font-medium shadow-sm transition-colors',
          'hover:bg-accent hover:text-accent-foreground',
          'focus:outline-none focus:ring-1 focus:ring-ring',
          'active:scale-[0.98]',
          'disabled:pointer-events-none disabled:opacity-50',
          !selectedTimeframe && 'text-muted-foreground'
        )}
        onClick={() => setIsOpen(!isOpen)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            setIsOpen(!isOpen)
          }
        }}
      >
        <span className="max-w-[100px] truncate">{selectedLabel}</span>
        <ChevronDown className="h-3.5 w-3.5 shrink-0 opacity-60" />
      </button>

      {isOpen && (
        <div
          ref={dropdownRef}
          className="absolute left-0 top-full z-50 mt-1 min-w-[160px] overflow-hidden rounded-md border bg-popover text-popover-foreground shadow-lg animate-in fade-in-0 zoom-in-95"
          role="listbox"
        >
          <div className="p-1">
            {timeframes.map((timeframe) => {
              const isSelected = timeframe === selectedTimeframe
              return (
                <div
                  key={timeframe}
                  role="option"
                  aria-selected={isSelected}
                  tabIndex={0}
                  className={cn(
                    'relative flex cursor-pointer select-none items-center rounded-sm py-2 pl-2 pr-8 text-sm outline-none transition-colors',
                    isSelected && 'bg-accent',
                    !isSelected && 'hover:bg-accent/50'
                  )}
                  onClick={() => handleSelect(timeframe)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                      e.preventDefault()
                      handleSelect(timeframe)
                    }
                  }}
                >
                  <span className="absolute right-2 flex h-3.5 w-3.5 items-center justify-center">
                    {isSelected && <Check className="h-4 w-4" />}
                  </span>
                  <span>{timeframeLabels[timeframe] || timeframe}</span>
                </div>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}

export const TimeframeSelector = ({ timeframes, selectedTimeframe, onTimeframeChange, disabled, compact = false }: TimeframeSelectorProps) => {
  const handleValueChange = (value: string) => {
    onTimeframeChange(value)
  }

  return (
    <div className={compact ? '' : 'space-y-2'}>
      {!compact && <Label htmlFor="timeframe-select">Timeframe</Label>}
      <Select value={selectedTimeframe || undefined} onValueChange={handleValueChange} disabled={disabled}>
        <SelectTrigger id="timeframe-select" className="w-full h-9" disabled={disabled}>
          <SelectValue placeholder="Timeframe" />
        </SelectTrigger>
        <SelectContent>
          {timeframes.map((timeframe) => (
            <SelectItem key={timeframe} value={timeframe}>
              {timeframeLabels[timeframe] || timeframe}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  )
}
