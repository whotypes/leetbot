interface TimeframeSelectorProps {
  timeframes: string[]
  selectedTimeframe: string
  onTimeframeChange: (timeframe: string) => void
  disabled: boolean
}

const timeframeLabels: Record<string, string> = {
  'all': 'All Time',
  'thirty-days': 'Last 30 Days',
  'three-months': 'Last 3 Months',
  'six-months': 'Last 6 Months',
  'more-than-six-months': 'More than 6 Months'
}

export const TimeframeSelector = ({ timeframes, selectedTimeframe, onTimeframeChange, disabled }: TimeframeSelectorProps) => {
  return (
    <div>
      <label htmlFor="timeframe-select" className="block text-sm font-medium mb-2" style={{ color: 'var(--color-content)' }}>
        Timeframe
      </label>
      <select
        id="timeframe-select"
        value={selectedTimeframe}
        onChange={(e) => onTimeframeChange(e.target.value)}
        disabled={disabled}
        className={`w-full p-4 border rounded-md focus:outline-none focus:ring-2 ${
          disabled ? 'cursor-not-allowed opacity-60' : ''
        }`}
        style={{
          borderColor: 'var(--color-muted)',
          backgroundColor: disabled ? 'var(--color-muted)' : 'var(--color-surface)',
          color: 'var(--color-content)',
          '--tw-ring-color': 'var(--color-primary)'
        } as React.CSSProperties & { '--tw-ring-color': string }}
      >
        <option value="">Select a timeframe</option>
        {timeframes.map((timeframe) => (
          <option key={timeframe} value={timeframe}>
            {timeframeLabels[timeframe] || timeframe}
          </option>
        ))}
      </select>
    </div>
  )
}
