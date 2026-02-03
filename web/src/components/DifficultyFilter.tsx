import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { cn } from '@/lib/utils'
import { Filter, X } from 'lucide-react'

export type Difficulty = 'easy' | 'medium' | 'hard'

interface DifficultyFilterProps {
  selectedDifficulties: Difficulty[]
  onDifficultyChange: (difficulties: Difficulty[]) => void
}

const difficulties: { value: Difficulty; label: string; className: string }[] = [
  {
    value: 'easy',
    label: 'Easy',
    className: 'text-green-600 dark:text-green-400',
  },
  {
    value: 'medium',
    label: 'Medium',
    className: 'text-yellow-600 dark:text-yellow-400',
  },
  {
    value: 'hard',
    label: 'Hard',
    className: 'text-red-600 dark:text-red-400',
  },
]

export const DifficultyFilter = ({
  selectedDifficulties,
  onDifficultyChange,
}: DifficultyFilterProps) => {
  const handleToggle = (difficulty: Difficulty) => {
    if (selectedDifficulties.includes(difficulty)) {
      onDifficultyChange(selectedDifficulties.filter((d) => d !== difficulty))
    } else {
      onDifficultyChange([...selectedDifficulties, difficulty])
    }
  }

  const handleClearAll = () => {
    onDifficultyChange([])
  }

  const hasSelection = selectedDifficulties.length > 0

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className={cn(
            'h-9 gap-2',
            hasSelection && 'border-primary'
          )}
          aria-label="Filter by difficulty"
        >
          <Filter className="h-4 w-4" />
          <span className="hidden sm:inline">Difficulty</span>
          {hasSelection && (
            <Badge
              variant="secondary"
              className="ml-1 h-5 min-w-5 rounded-full px-1.5 text-xs"
            >
              {selectedDifficulties.length}
            </Badge>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-48 p-3" align="start">
        <div className="flex items-center justify-between mb-3">
          <span className="text-sm font-medium">Filter by difficulty</span>
          {hasSelection && (
            <Button
              variant="ghost"
              size="sm"
              className="h-6 px-2 text-xs"
              onClick={handleClearAll}
              aria-label="Clear all filters"
            >
              <X className="h-3 w-3 mr-1" />
              Clear
            </Button>
          )}
        </div>
        <div className="space-y-2">
          {difficulties.map((difficulty) => (
            <div key={difficulty.value} className="flex items-center space-x-2">
              <Checkbox
                id={`difficulty-${difficulty.value}`}
                checked={selectedDifficulties.includes(difficulty.value)}
                onCheckedChange={() => handleToggle(difficulty.value)}
              />
              <Label
                htmlFor={`difficulty-${difficulty.value}`}
                className={cn('cursor-pointer font-normal', difficulty.className)}
              >
                {difficulty.label}
              </Label>
            </div>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  )
}
