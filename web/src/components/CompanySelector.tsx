import { Button } from '@/components/ui/button'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import { Label } from '@/components/ui/label'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { cn } from '@/lib/utils'
import { Check, ChevronsUpDown } from 'lucide-react'
import * as React from 'react'

interface CompanySelectorProps {
  companies: string[]
  selectedCompany: string
  onCompanyChange: (company: string) => void
  onCompanyPreview?: (company: string) => void
}

export const CompanySelector = ({ companies, selectedCompany, onCompanyChange, onCompanyPreview }: CompanySelectorProps) => {
  const [open, setOpen] = React.useState(false)
  const triggerRef = React.useRef<HTMLButtonElement>(null)
  const [popoverWidth, setPopoverWidth] = React.useState<number | undefined>(undefined)
  const [highlightedValue, setHighlightedValue] = React.useState<string>('')
  const isNavigatingRef = React.useRef(false)
  const lastTypingTimeRef = React.useRef<number>(0)

  React.useEffect(() => {
    if (triggerRef.current) {
      setPopoverWidth(triggerRef.current.offsetWidth)
    }
  }, [open])

  React.useEffect(() => {
    if (!open) {
      setHighlightedValue('')
      isNavigatingRef.current = false
      if (onCompanyPreview) {
        onCompanyPreview('')
      }
    }
  }, [open, onCompanyPreview])

  const selectedCompanyData = selectedCompany
    ? companies.find((c) => c === selectedCompany)
    : null

  const selectedLabel = selectedCompanyData
    ? selectedCompanyData.charAt(0).toUpperCase() + selectedCompanyData.slice(1).replace(/-/g, ' ')
    : null

  const handleValueChange = (value: string) => {
    setHighlightedValue(value)

    if (!value || !open) return

    const now = Date.now()
    const timeSinceLastTyping = now - lastTypingTimeRef.current

    const company = companies.find((c) => {
      const label = c.charAt(0).toUpperCase() + c.slice(1).replace(/-/g, ' ')
      const itemValue = `${c} ${label}`
      return value === itemValue
    })

    if (company) {
      const isExactMatch = companies.some((c) => {
        const label = c.charAt(0).toUpperCase() + c.slice(1).replace(/-/g, ' ')
        return value === `${c} ${label}`
      })

      if (isNavigatingRef.current || (isExactMatch && timeSinceLastTyping > 150)) {
        if (onCompanyPreview && company !== selectedCompany) {
          onCompanyPreview(company)
        }
      }
    }
  }

  const handleInputKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
      isNavigatingRef.current = true
    } else if (e.key === 'Enter') {
      isNavigatingRef.current = false
      const highlightedCompany = companies.find((c) => {
        const label = c.charAt(0).toUpperCase() + c.slice(1).replace(/-/g, ' ')
        return highlightedValue === `${c} ${label}`
      })
      if (highlightedCompany) {
        onCompanyChange(highlightedCompany)
        setOpen(false)
      }
    } else if (e.key.length === 1 || e.key === 'Backspace' || e.key === 'Delete') {
      isNavigatingRef.current = false
      lastTypingTimeRef.current = Date.now()
    }
  }

  const handleCommandKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
      isNavigatingRef.current = true
    } else if (e.key === 'Enter') {
      isNavigatingRef.current = false
      const highlightedCompany = companies.find((c) => {
        const label = c.charAt(0).toUpperCase() + c.slice(1).replace(/-/g, ' ')
        return highlightedValue === `${c} ${label}`
      })
      if (highlightedCompany) {
        onCompanyChange(highlightedCompany)
        setOpen(false)
      }
    }
  }

  return (
    <div className="space-y-2">
      <Label htmlFor="company-select">Company</Label>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            ref={triggerRef}
            id="company-select"
            variant="outline"
            role="combobox"
            aria-expanded={open}
            className="w-full justify-between"
          >
            {selectedLabel || 'Select a company...'}
            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent
          className="p-0"
          align="start"
          sideOffset={4}
          style={{ width: popoverWidth ? `${popoverWidth}px` : undefined }}
        >
          <Command
            value={highlightedValue}
            onValueChange={handleValueChange}
            onKeyDown={handleCommandKeyDown}
          >
            <CommandInput
              placeholder="Search company..."
              onKeyDown={handleInputKeyDown}
            />
            <CommandList>
              <CommandEmpty>No company found.</CommandEmpty>
              <CommandGroup>
                {companies.map((company) => {
                  const label = company.charAt(0).toUpperCase() + company.slice(1).replace(/-/g, ' ')
                  const itemValue = `${company} ${label}`
                  const handleSelect = () => {
                    isNavigatingRef.current = false
                    onCompanyChange(company)
                    setOpen(false)
                  }
                  return (
                    <CommandItem
                      key={company}
                      value={itemValue}
                      onSelect={handleSelect}
                      onMouseDown={(e) => {
                        e.preventDefault()
                        isNavigatingRef.current = false
                        handleSelect()
                      }}
                    >
                      <Check
                        className={cn(
                          'mr-2 h-4 w-4',
                          selectedCompany === company ? 'opacity-100' : 'opacity-0'
                        )}
                      />
                      {label}
                    </CommandItem>
                  )
                })}
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
    </div>
  )
}
