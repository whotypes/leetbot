import { Badge } from '@/components/ui/badge'
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table'
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'
import {
    flexRender,
    getCoreRowModel,
    getSortedRowModel,
    useReactTable,
    type ColumnDef,
    type SortingState,
} from '@tanstack/react-table'
import { useVirtualizer } from '@tanstack/react-virtual'
import { ArrowDown, ArrowUp, ArrowUpDown, ExternalLink } from 'lucide-react'
import * as React from 'react'
import type { Problem } from '../types'
import { MobileCompanySelector } from './CompanySelector'
import { MobileTimeframeSelector } from './TimeframeSelector'

interface ProblemsDataTableProps {
  problems: Problem[]
  isLoading?: boolean
  companies?: string[]
  selectedCompany?: string
  onCompanyChange?: (company: string) => void
  timeframes?: string[]
  selectedTimeframe?: string
  onTimeframeChange?: (timeframe: string) => void
}

const getDifficultyClassName = (difficulty: string): string => {
  switch (difficulty.toLowerCase()) {
    case 'easy':
      return 'bg-green-100 text-green-800 border-green-200 dark:bg-green-900/30 dark:text-green-400 dark:border-green-800'
    case 'medium':
      return 'bg-yellow-100 text-yellow-800 border-yellow-200 dark:bg-yellow-900/30 dark:text-yellow-400 dark:border-yellow-800'
    case 'hard':
      return 'bg-red-100 text-red-800 border-red-200 dark:bg-red-900/30 dark:text-red-400 dark:border-red-800'
    default:
      return ''
  }
}

const SortIcon = ({ isSorted }: { isSorted: false | 'asc' | 'desc' }) => {
  if (isSorted === 'asc') return <ArrowUp className="ml-1 h-4 w-4" />
  if (isSorted === 'desc') return <ArrowDown className="ml-1 h-4 w-4" />
  return <ArrowUpDown className="ml-1 h-4 w-4 opacity-50" />
}

const columns: ColumnDef<Problem>[] = [
  {
    accessorKey: 'id',
    header: ({ column }) => (
      <button
        type="button"
        className="flex items-center font-medium hover:text-foreground transition-colors"
        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
        aria-label={`Sort by ID ${column.getIsSorted() === 'asc' ? 'descending' : 'ascending'}`}
      >
        ID
        <SortIcon isSorted={column.getIsSorted()} />
      </button>
    ),
    cell: ({ row }) => (
      <span className="font-mono text-muted-foreground">{row.getValue('id')}</span>
    ),
    size: 80,
  },
  {
    accessorKey: 'title',
    header: ({ column }) => (
      <button
        type="button"
        className="flex items-center font-medium hover:text-foreground transition-colors"
        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
        aria-label={`Sort by title ${column.getIsSorted() === 'asc' ? 'descending' : 'ascending'}`}
      >
        Title
        <SortIcon isSorted={column.getIsSorted()} />
      </button>
    ),
    cell: ({ row }) => {
      const problem = row.original
      return (
        <a
          href={problem.url}
          target="_blank"
          rel="noopener noreferrer"
          className="group inline-flex items-center gap-1.5 text-primary hover:underline"
          tabIndex={0}
        >
          {problem.title}
          <ExternalLink className="h-3 w-3 opacity-0 group-hover:opacity-100 transition-opacity" />
        </a>
      )
    },
    size: 400,
  },
  {
    accessorKey: 'difficulty',
    header: ({ column }) => (
      <button
        type="button"
        className="flex items-center font-medium hover:text-foreground transition-colors"
        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
        aria-label={`Sort by difficulty ${column.getIsSorted() === 'asc' ? 'descending' : 'ascending'}`}
      >
        Difficulty
        <SortIcon isSorted={column.getIsSorted()} />
      </button>
    ),
    cell: ({ row }) => {
      const difficulty = row.getValue('difficulty') as string
      return (
        <Badge variant="outline" className={getDifficultyClassName(difficulty)}>
          {difficulty}
        </Badge>
      )
    },
    sortingFn: (rowA, rowB) => {
      const order = { easy: 0, medium: 1, hard: 2 }
      const a = (rowA.getValue('difficulty') as string).toLowerCase()
      const b = (rowB.getValue('difficulty') as string).toLowerCase()
      return (order[a as keyof typeof order] ?? 3) - (order[b as keyof typeof order] ?? 3)
    },
    size: 120,
  },
  {
    accessorKey: 'acceptance',
    header: ({ column }) => (
      <button
        type="button"
        className="flex items-center font-medium hover:text-foreground transition-colors"
        onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
        aria-label={`Sort by acceptance ${column.getIsSorted() === 'asc' ? 'descending' : 'ascending'}`}
      >
        Acceptance
        <SortIcon isSorted={column.getIsSorted()} />
      </button>
    ),
    cell: ({ row }) => (
      <span className="tabular-nums">{(row.getValue('acceptance') as number).toFixed(1)}%</span>
    ),
    size: 120,
  },
  {
    accessorKey: 'frequency',
    header: ({ column }) => (
      <div className="flex items-center gap-1">
        <button
          type="button"
          className="flex items-center font-medium hover:text-foreground transition-colors"
          onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}
          aria-label={`Sort by frequency ${column.getIsSorted() === 'asc' ? 'descending' : 'ascending'}`}
        >
          Frequency
          <SortIcon isSorted={column.getIsSorted()} />
        </button>
        <TooltipProvider delayDuration={100}>
          <Tooltip>
            <TooltipTrigger asChild>
              <a
                href="https://leetcode.com/discuss/post/1912580/answered-leetcodes-lists-and-frequency-c-u6vi/comments/1337781/"
                target="_blank"
                rel="noopener noreferrer"
                className="text-muted-foreground hover:text-foreground transition-colors"
                onClick={(e) => e.stopPropagation()}
                tabIndex={0}
                aria-label="What is frequency?"
              >
                (?)
              </a>
            </TooltipTrigger>
            <TooltipContent>
              <p>What is this?</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>
    ),
    cell: ({ row }) => {
      const frequency = row.getValue('frequency') as number
      return (
        <div className="flex items-center gap-2">
          <div className="h-2 w-16 rounded-full bg-muted overflow-hidden">
            <div
              className="h-full bg-primary transition-all"
              style={{ width: `${Math.min(frequency, 100)}%` }}
            />
          </div>
          <span className="tabular-nums text-muted-foreground">{frequency.toFixed(1)}%</span>
        </div>
      )
    },
    size: 180,
  },
]

const ROW_HEIGHT = 48
const SKELETON_ROW_COUNT = 10

const SkeletonCell = ({ width, className }: { width: string; className?: string }) => (
  <div className={cn('h-4 rounded bg-muted', className)} style={{ width }} />
)

const SkeletonRow = ({ index }: { index: number }) => (
  <TableRow
    className={cn(
      'transition-colors',
      index % 2 === 0 ? 'bg-background' : 'bg-muted/30'
    )}
    style={{ height: `${ROW_HEIGHT}px` }}
  >
    <TableCell style={{ width: 80 }}>
      <SkeletonCell width="40px" />
    </TableCell>
    <TableCell style={{ width: 400 }}>
      <SkeletonCell width={`${180 + (index % 3) * 60}px`} />
    </TableCell>
    <TableCell style={{ width: 120 }}>
      <SkeletonCell width="60px" className="h-5 rounded-full" />
    </TableCell>
    <TableCell style={{ width: 120 }}>
      <SkeletonCell width="50px" />
    </TableCell>
    <TableCell style={{ width: 180 }}>
      <div className="flex items-center gap-2">
        <div className="h-2 w-16 rounded-full bg-muted" />
        <SkeletonCell width="45px" className="h-4" />
      </div>
    </TableCell>
  </TableRow>
)

export const ProblemsDataTable = ({
  problems,
  isLoading = false,
  companies = [],
  selectedCompany = '',
  onCompanyChange,
  timeframes = [],
  selectedTimeframe = '',
  onTimeframeChange,
}: ProblemsDataTableProps) => {
  const [sorting, setSorting] = React.useState<SortingState>([])
  const tableContainerRef = React.useRef<HTMLDivElement>(null)

  const table = useReactTable({
    data: problems,
    columns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  })

  const { rows } = table.getRowModel()

  const rowVirtualizer = useVirtualizer({
    count: rows.length,
    estimateSize: () => ROW_HEIGHT,
    getScrollElement: () => tableContainerRef.current,
    overscan: 10,
  })

  const virtualRows = rowVirtualizer.getVirtualItems()
  const totalSize = rowVirtualizer.getTotalSize()

  const paddingTop = virtualRows.length > 0 ? virtualRows[0]?.start ?? 0 : 0
  const paddingBottom =
    virtualRows.length > 0 ? totalSize - (virtualRows[virtualRows.length - 1]?.end ?? 0) : 0

  if (!isLoading && problems.length === 0) {
    return (
      <div className="flex h-full items-center justify-center text-muted-foreground">
        No problems to display
      </div>
    )
  }

  const showMobileFilters = companies.length > 0 && onCompanyChange && onTimeframeChange

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center justify-between border-b px-4 py-3">
        {showMobileFilters && (
          <div className="flex items-center gap-2 sm:hidden">
            <MobileCompanySelector
              companies={companies}
              selectedCompany={selectedCompany}
              onCompanyChange={onCompanyChange}
            />
            <MobileTimeframeSelector
              timeframes={timeframes}
              selectedTimeframe={selectedTimeframe}
              onTimeframeChange={onTimeframeChange}
              disabled={!selectedCompany}
            />
          </div>
        )}
        <div className={cn(
          "text-sm text-muted-foreground",
          showMobileFilters && "ml-auto sm:ml-0"
        )}>
          {isLoading ? (
            <div className="h-4 w-24 rounded bg-muted" />
          ) : (
            <>
              <span className="font-medium text-foreground">{problems.length.toLocaleString()}</span>{' '}
              {problems.length === 1 ? 'problem' : 'problems'}
            </>
          )}
        </div>
      </div>

      <div
        ref={tableContainerRef}
        className="flex-1 overflow-auto"
      >
        <Table>
          <TableHeader className="sticky top-0 z-10 bg-background">
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id} className="hover:bg-transparent">
                {headerGroup.headers.map((header) => (
                  <TableHead
                    key={header.id}
                    style={{ width: header.getSize() }}
                    className="bg-background"
                  >
                    {header.isPlaceholder
                      ? null
                      : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {isLoading ? (
              Array.from({ length: SKELETON_ROW_COUNT }).map((_, index) => (
                <SkeletonRow key={index} index={index} />
              ))
            ) : (
              <>
                {paddingTop > 0 && (
                  <tr>
                    <td style={{ height: `${paddingTop}px` }} />
                  </tr>
                )}
                {virtualRows.map((virtualRow) => {
                  const row = rows[virtualRow.index]
                  return (
                    <TableRow
                      key={row.id}
                      data-index={virtualRow.index}
                      className={cn(
                        'transition-colors',
                        virtualRow.index % 2 === 0 ? 'bg-background' : 'bg-muted/30'
                      )}
                      style={{ height: `${ROW_HEIGHT}px` }}
                    >
                      {row.getVisibleCells().map((cell) => (
                        <TableCell key={cell.id} style={{ width: cell.column.getSize() }}>
                          {flexRender(cell.column.columnDef.cell, cell.getContext())}
                        </TableCell>
                      ))}
                    </TableRow>
                  )
                })}
                {paddingBottom > 0 && (
                  <tr>
                    <td style={{ height: `${paddingBottom}px` }} />
                  </tr>
                )}
              </>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
