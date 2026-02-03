import { Label } from '@/components/ui/label'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'

interface TimeframeSelectorProps {
  timeframes: string[]
  selectedTimeframe: string
  onTimeframeChange: (timeframe: string) => void
  disabled: boolean
  compact?: boolean
}

const timeframeLabels: Record<string, string> = {
  'all': 'All Time',
  'thirty-days': 'Last 30 Days',
  'three-months': 'Last 3 Months',
  'six-months': 'Last 6 Months',
  'more-than-six-months': 'More than 6 Months'
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
