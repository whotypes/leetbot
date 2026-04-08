import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import { Check, ChevronDown, ChevronUp, Search, X } from 'lucide-react'
import * as React from 'react'

interface CompanySelectorProps {
  companies: string[]
  selectedCompany: string
  onCompanyChange: (company: string) => void
  onCompanyPreview?: (company: string) => void
  compact?: boolean
}

const ITEM_HEIGHT = 36
const SCROLLBAR_WIDTH = 16

const formatLabel = (company: string) =>
  company.charAt(0).toUpperCase() + company.slice(1).replace(/-/g, ' ')

interface VirtualizedListProps {
  items: Array<{ value: string; label: string; searchLabel: string }>
  selectedValue: string
  onSelect: (value: string) => void
  search: string
  maxHeight: number
}

const VirtualizedList = ({
  items,
  selectedValue,
  onSelect,
  search,
  maxHeight,
}: VirtualizedListProps) => {
  const containerRef = React.useRef<HTMLDivElement>(null)
  const [virtualPosition, setVirtualPosition] = React.useState(0)
  const [focusedIndex, setFocusedIndex] = React.useState(-1)
  const [isDragging, setIsDragging] = React.useState(false)
  const trackRef = React.useRef<HTMLDivElement>(null)
  const thumbRef = React.useRef<HTMLDivElement>(null)

  const filteredItems = React.useMemo(() => {
    if (!search) return items
    const lowerSearch = search.toLowerCase()
    return items.filter((item) => item.searchLabel.toLowerCase().includes(lowerSearch))
  }, [items, search])

  const itemsToShow = React.useMemo(() => {
    return Math.floor(maxHeight / ITEM_HEIGHT)
  }, [maxHeight])

  const maxPosition = React.useMemo(() => {
    return Math.max(0, filteredItems.length - itemsToShow)
  }, [filteredItems.length, itemsToShow])

  React.useEffect(() => {
    setVirtualPosition(0)
    setFocusedIndex(filteredItems.length > 0 ? 0 : -1)
  }, [search, filteredItems.length])

  React.useEffect(() => {
    if (selectedValue && !search) {
      const selectedIndex = filteredItems.findIndex((item) => item.value === selectedValue)
      if (selectedIndex !== -1) {
        const targetPosition = Math.max(0, Math.min(selectedIndex - Math.floor(itemsToShow / 2), maxPosition))
        setVirtualPosition(targetPosition)
      }
    }
  }, [selectedValue, filteredItems, itemsToShow, maxPosition, search])

  const displayedItems = React.useMemo(() => {
    const start = Math.floor(virtualPosition)
    const end = Math.min(start + itemsToShow + 1, filteredItems.length)
    return filteredItems.slice(start, end).map((item, i) => ({
      ...item,
      absoluteIndex: start + i,
    }))
  }, [virtualPosition, itemsToShow, filteredItems])

  const movePosition = React.useCallback(
    (delta: number) => {
      setVirtualPosition((prev) => {
        const newPos = prev + delta
        return Math.max(0, Math.min(newPos, maxPosition))
      })
    },
    [maxPosition]
  )

  React.useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const handleWheel = (e: WheelEvent) => {
      e.preventDefault()
      const delta = e.deltaY / ITEM_HEIGHT
      movePosition(delta)
    }

    container.addEventListener('wheel', handleWheel, { passive: false })

    let lastTouchY = 0
    let velocity = 0
    let animationFrame: number | null = null

    const applyMomentum = () => {
      if (Math.abs(velocity) > 0.1) {
        movePosition(velocity)
        velocity *= 0.92
        animationFrame = requestAnimationFrame(applyMomentum)
      }
    }

    const handleTouchStart = (e: TouchEvent) => {
      lastTouchY = e.touches[0].clientY
      velocity = 0
      if (animationFrame) {
        cancelAnimationFrame(animationFrame)
      }
    }

    const handleTouchMove = (e: TouchEvent) => {
      e.preventDefault()
      const touchY = e.touches[0].clientY
      const deltaY = lastTouchY - touchY
      velocity = deltaY / ITEM_HEIGHT / 2
      lastTouchY = touchY
      movePosition(velocity)
    }

    const handleTouchEnd = () => {
      if (Math.abs(velocity) > 0.1) {
        animationFrame = requestAnimationFrame(applyMomentum)
      }
    }

    container.addEventListener('touchstart', handleTouchStart, { passive: false })
    container.addEventListener('touchmove', handleTouchMove, { passive: false })
    container.addEventListener('touchend', handleTouchEnd, { passive: false })

    return () => {
      container.removeEventListener('wheel', handleWheel)
      container.removeEventListener('touchstart', handleTouchStart)
      container.removeEventListener('touchmove', handleTouchMove)
      container.removeEventListener('touchend', handleTouchEnd)
      if (animationFrame) cancelAnimationFrame(animationFrame)
    }
  }, [movePosition])

  const handleKeyDown = React.useCallback(
    (e: React.KeyboardEvent) => {
      const pageSize = itemsToShow

      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault()
          if (focusedIndex < filteredItems.length - 1) {
            const newIndex = focusedIndex + 1
            setFocusedIndex(newIndex)
            if (newIndex >= virtualPosition + itemsToShow - 1) {
              movePosition(1)
            }
          }
          break
        case 'ArrowUp':
          e.preventDefault()
          if (focusedIndex > 0) {
            const newIndex = focusedIndex - 1
            setFocusedIndex(newIndex)
            if (newIndex < virtualPosition) {
              movePosition(-1)
            }
          } else if (focusedIndex === -1 && filteredItems.length > 0) {
            setFocusedIndex(0)
          }
          break
        case 'PageDown':
          e.preventDefault()
          {
            const newIndex = Math.min(focusedIndex + pageSize, filteredItems.length - 1)
            setFocusedIndex(newIndex)
            movePosition(pageSize)
          }
          break
        case 'PageUp':
          e.preventDefault()
          {
            const newIndex = Math.max(focusedIndex - pageSize, 0)
            setFocusedIndex(newIndex)
            movePosition(-pageSize)
          }
          break
        case 'Home':
          e.preventDefault()
          setFocusedIndex(0)
          setVirtualPosition(0)
          break
        case 'End':
          e.preventDefault()
          setFocusedIndex(filteredItems.length - 1)
          setVirtualPosition(maxPosition)
          break
        case 'Enter':
          e.preventDefault()
          if (focusedIndex >= 0 && focusedIndex < filteredItems.length) {
            onSelect(filteredItems[focusedIndex].value)
          }
          break
      }
    },
    [focusedIndex, filteredItems, itemsToShow, virtualPosition, maxPosition, movePosition, onSelect]
  )

  const scrollPercentage = React.useMemo(() => {
    if (maxPosition === 0) return 0
    return (virtualPosition / maxPosition) * 100
  }, [virtualPosition, maxPosition])

  const thumbHeight = React.useMemo(() => {
    if (filteredItems.length === 0) return 0
    const ratio = itemsToShow / filteredItems.length
    return Math.max(20, Math.min(ratio * maxHeight, maxHeight))
  }, [itemsToShow, filteredItems.length, maxHeight])

  const thumbPosition = React.useMemo(() => {
    const trackHeight = maxHeight - thumbHeight
    return (scrollPercentage / 100) * trackHeight
  }, [scrollPercentage, maxHeight, thumbHeight])

  const handleTrackClick = (e: React.MouseEvent) => {
    if (e.target === thumbRef.current || !trackRef.current) return
    const rect = trackRef.current.getBoundingClientRect()
    const trackHeight = rect.height - thumbHeight
    let clickPosition = (e.clientY - rect.top - thumbHeight / 2) / trackHeight
    clickPosition = Math.max(0, Math.min(1, clickPosition))
    setVirtualPosition(Math.floor(clickPosition * maxPosition))
  }

  const handleDrag = React.useCallback(
    (e: MouseEvent) => {
      if (!isDragging || !trackRef.current) return
      const rect = trackRef.current.getBoundingClientRect()
      const trackHeight = rect.height - thumbHeight
      let percentage = (e.clientY - rect.top - thumbHeight / 2) / trackHeight
      percentage = Math.max(0, Math.min(1, percentage))
      setVirtualPosition(Math.floor(percentage * maxPosition))
    },
    [isDragging, thumbHeight, maxPosition]
  )

  const handleDragEnd = React.useCallback(() => {
    setIsDragging(false)
  }, [])

  React.useEffect(() => {
    if (isDragging) {
      window.addEventListener('pointermove', handleDrag)
      window.addEventListener('pointerup', handleDragEnd)
      return () => {
        window.removeEventListener('pointermove', handleDrag)
        window.removeEventListener('pointerup', handleDragEnd)
      }
    }
  }, [isDragging, handleDrag, handleDragEnd])

  const handleFocus = React.useCallback(() => {
    if (focusedIndex === -1 && filteredItems.length > 0) {
      const selectedIdx = filteredItems.findIndex((item) => item.value === selectedValue)
      setFocusedIndex(selectedIdx !== -1 ? selectedIdx : 0)
    }
  }, [focusedIndex, filteredItems, selectedValue])

  const highlightMatch = (text: string, query: string) => {
    if (!query) return text
    const lowerText = text.toLowerCase()
    const lowerQuery = query.toLowerCase()
    const index = lowerText.indexOf(lowerQuery)
    if (index === -1) return text
    return (
      <>
        {text.slice(0, index)}
        <span className="bg-yellow-200 dark:bg-yellow-800">{text.slice(index, index + query.length)}</span>
        {text.slice(index + query.length)}
      </>
    )
  }

  const showScrollbar = filteredItems.length > itemsToShow

  if (filteredItems.length === 0) {
    return (
      <div className="flex items-center justify-center py-6 text-sm text-muted-foreground">
        No companies found
      </div>
    )
  }

  return (
    <div
      className="flex"
      data-list-container
      onKeyDown={handleKeyDown}
      onFocus={handleFocus}
      tabIndex={0}
    >
      <div
        ref={containerRef}
        className="flex-1 overflow-hidden outline-none"
        style={{ height: Math.min(filteredItems.length * ITEM_HEIGHT, maxHeight) }}
      >
        {displayedItems.map((item) => {
          const isSelected = item.value === selectedValue
          const isFocused = item.absoluteIndex === focusedIndex
          return (
            <div
              key={item.value}
              role="option"
              aria-selected={isSelected}
              tabIndex={-1}
              className={cn(
                'relative flex cursor-pointer select-none items-center rounded-sm py-1.5 pl-2 pr-8 text-sm outline-none transition-colors',
                isSelected && 'bg-accent',
                isFocused && 'bg-accent/50',
                !isSelected && !isFocused && 'hover:bg-accent/30'
              )}
              style={{ height: ITEM_HEIGHT }}
              onClick={() => onSelect(item.value)}
              onMouseEnter={() => setFocusedIndex(item.absoluteIndex)}
            >
              <span className="absolute right-2 flex h-3.5 w-3.5 items-center justify-center">
                {isSelected && <Check className="h-4 w-4" />}
              </span>
              <span className="truncate">{highlightMatch(item.label, search)}</span>
            </div>
          )
        })}
      </div>

      {showScrollbar && (
        <div className="flex flex-col" style={{ width: SCROLLBAR_WIDTH }}>
          <button
            type="button"
            className="flex h-6 items-center justify-center hover:bg-accent/50"
            onClick={() => setVirtualPosition(0)}
            aria-label="Scroll to top"
          >
            <ChevronUp className="h-3 w-3" />
          </button>
          <div
            ref={trackRef}
            className="relative flex-1 cursor-pointer"
            onClick={handleTrackClick}
            style={{ height: maxHeight - 48 }}
          >
            <div
              ref={thumbRef}
              className="absolute left-1 right-1 cursor-grab rounded bg-border hover:bg-muted-foreground/50 active:cursor-grabbing"
              style={{
                height: thumbHeight,
                top: thumbPosition,
              }}
              onPointerDown={(e) => {
                e.preventDefault()
                setIsDragging(true)
              }}
            />
          </div>
          <button
            type="button"
            className="flex h-6 items-center justify-center hover:bg-accent/50"
            onClick={() => setVirtualPosition(maxPosition)}
            aria-label="Scroll to bottom"
          >
            <ChevronDown className="h-3 w-3" />
          </button>
        </div>
      )}
    </div>
  )
}

interface MobileCompanySelectorProps {
  companies: string[]
  selectedCompany: string
  onCompanyChange: (company: string) => void
}

export const MobileCompanySelector = ({
  companies,
  selectedCompany,
  onCompanyChange,
}: MobileCompanySelectorProps) => {
  const [isOpen, setIsOpen] = React.useState(false)
  const [search, setSearch] = React.useState('')
  const triggerRef = React.useRef<HTMLButtonElement>(null)
  const dropdownRef = React.useRef<HTMLDivElement>(null)
  const inputRef = React.useRef<HTMLInputElement>(null)

  const items = React.useMemo(() => {
    return companies.map((company) => ({
      value: company,
      label: formatLabel(company),
      searchLabel: `${company} ${formatLabel(company)}`,
    }))
  }, [companies])

  const handleSelect = React.useCallback(
    (value: string) => {
      onCompanyChange(value)
      setIsOpen(false)
      setSearch('')
    },
    [onCompanyChange]
  )

  React.useEffect(() => {
    if (isOpen) {
      const timer = setTimeout(() => inputRef.current?.focus(), 0)
      return () => clearTimeout(timer)
    }
  }, [isOpen])

  React.useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(e.target as Node) &&
        triggerRef.current &&
        !triggerRef.current.contains(e.target as Node)
      ) {
        setIsOpen(false)
        setSearch('')
      }
    }

    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setIsOpen(false)
        setSearch('')
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

  const selectedLabel = React.useMemo(() => {
    if (!selectedCompany) return null
    return formatLabel(selectedCompany)
  }, [selectedCompany])

  return (
    <div className="relative">
      <button
        ref={triggerRef}
        type="button"
        role="combobox"
        aria-expanded={isOpen}
        aria-haspopup="listbox"
        aria-label="Select company"
        className={cn(
          'inline-flex h-8 items-center gap-1.5 rounded-md border border-input bg-background px-2.5 text-sm font-medium shadow-sm transition-colors',
          'hover:bg-accent hover:text-accent-foreground',
          'focus:outline-none focus:ring-1 focus:ring-ring',
          'active:scale-[0.98]',
          !selectedCompany && 'text-muted-foreground'
        )}
        onClick={() => setIsOpen(!isOpen)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            setIsOpen(!isOpen)
          }
        }}
      >
        <span className="max-w-[140px] truncate">
          {selectedLabel ?? 'Company'}
        </span>
        <ChevronDown className="h-3.5 w-3.5 shrink-0 opacity-60" />
      </button>

      {isOpen && (
        <div
          ref={dropdownRef}
          className="absolute left-0 top-full z-50 mt-1 w-[280px] overflow-hidden rounded-md border bg-popover text-popover-foreground shadow-lg animate-in fade-in-0 zoom-in-95"
          role="listbox"
        >
          <div className="flex items-center border-b px-3 py-2">
            <Search className="mr-2 h-4 w-4 shrink-0 opacity-50" />
            <input
              ref={inputRef}
              type="text"
              className="flex h-8 w-full rounded-md bg-transparent text-sm outline-none placeholder:text-muted-foreground"
              placeholder="Search company..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'ArrowDown' || e.key === 'ArrowUp' || e.key === 'Enter') {
                  e.preventDefault()
                  dropdownRef.current?.querySelector('[data-list-container]')?.dispatchEvent(
                    new KeyboardEvent('keydown', { key: e.key, bubbles: true })
                  )
                }
              }}
            />
            {search && (
              <button
                type="button"
                className="ml-2 rounded p-0.5 hover:bg-accent"
                onClick={() => setSearch('')}
                aria-label="Clear search"
              >
                <X className="h-4 w-4 opacity-50" />
              </button>
            )}
          </div>
          <div className="p-1">
            <VirtualizedList
              items={items}
              selectedValue={selectedCompany}
              onSelect={handleSelect}
              search={search}
              maxHeight={280}
            />
          </div>
        </div>
      )}
    </div>
  )
}

export const CompanySelector = ({
  companies,
  selectedCompany,
  onCompanyChange,
  compact = false,
}: CompanySelectorProps) => {
  const [isOpen, setIsOpen] = React.useState(false)
  const [search, setSearch] = React.useState('')
  const triggerRef = React.useRef<HTMLButtonElement>(null)
  const dropdownRef = React.useRef<HTMLDivElement>(null)
  const inputRef = React.useRef<HTMLInputElement>(null)

  const items = React.useMemo(() => {
    return companies.map((company) => ({
      value: company,
      label: formatLabel(company),
      searchLabel: `${company} ${formatLabel(company)}`,
    }))
  }, [companies])

  const handleSelect = React.useCallback(
    (value: string) => {
      onCompanyChange(value)
      setIsOpen(false)
      setSearch('')
    },
    [onCompanyChange]
  )

  React.useEffect(() => {
    if (isOpen) {
      const timer = setTimeout(() => inputRef.current?.focus(), 0)
      return () => clearTimeout(timer)
    }
  }, [isOpen])

  React.useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(e.target as Node) &&
        triggerRef.current &&
        !triggerRef.current.contains(e.target as Node)
      ) {
        setIsOpen(false)
        setSearch('')
      }
    }

    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setIsOpen(false)
        setSearch('')
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

  const selectedLabel = React.useMemo(() => {
    if (!selectedCompany) return null
    return formatLabel(selectedCompany)
  }, [selectedCompany])

  return (
    <div className="relative">
      <div className={compact ? '' : 'space-y-2'}>
        {!compact && <Label htmlFor="company-select">Company</Label>}
        <button
          ref={triggerRef}
          id="company-select"
          type="button"
          role="combobox"
          aria-expanded={isOpen}
          aria-haspopup="listbox"
          className={cn(
            'flex h-9 w-full items-center justify-between whitespace-nowrap rounded-md border border-input bg-background px-3 py-2 text-sm shadow-sm ring-offset-background focus:outline-none focus:ring-1 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
            !selectedCompany && 'text-muted-foreground'
          )}
          onClick={() => setIsOpen(!isOpen)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.preventDefault()
              setIsOpen(!isOpen)
            } else if (e.key === 'ArrowDown' && !isOpen) {
              e.preventDefault()
              setIsOpen(true)
            }
          }}
        >
          <span className="truncate">{selectedLabel ?? 'Select a company'}</span>
          <ChevronDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </button>
      </div>

      {isOpen && (
        <div
          ref={dropdownRef}
          className="absolute left-0 top-full z-50 mt-1 w-full overflow-hidden rounded-md border bg-popover text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95"
          role="listbox"
        >
          <div className="flex items-center border-b px-3 py-2">
            <Search className="mr-2 h-4 w-4 shrink-0 opacity-50" />
            <input
              ref={inputRef}
              type="text"
              className="flex h-7 w-full rounded-md bg-transparent text-sm outline-none placeholder:text-muted-foreground"
              placeholder="Search company..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'ArrowDown' || e.key === 'ArrowUp' || e.key === 'Enter') {
                  e.preventDefault()
                  dropdownRef.current?.querySelector('[data-list-container]')?.dispatchEvent(
                    new KeyboardEvent('keydown', { key: e.key, bubbles: true })
                  )
                }
              }}
            />
            {search && (
              <button
                type="button"
                className="ml-2 rounded p-0.5 hover:bg-accent"
                onClick={() => setSearch('')}
                aria-label="Clear search"
              >
                <X className="h-4 w-4 opacity-50" />
              </button>
            )}
          </div>
          <div className="p-1">
            <VirtualizedList
              items={items}
              selectedValue={selectedCompany}
              onSelect={handleSelect}
              search={search}
              maxHeight={300}
            />
          </div>
        </div>
      )}
    </div>
  )
}
