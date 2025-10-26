import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'
import { CompanySelector } from './components/CompanySelector'
import { LoadingSpinner } from './components/LoadingSpinner'
import { ProblemsTable } from './components/ProblemsTable'
import { ThemeToggle } from './components/ThemeToggle'
import { TimeframeSelector } from './components/TimeframeSelector'
import { useLocalStorage } from './hooks/useLocalStorage'
import { useTheme } from './hooks/useTheme'
import type { APIResponse, Problem } from './types'

// Query functions
const fetchCompanies = async (): Promise<{ companies: string[] }> => {
  const response = await fetch('/api/companies')
  const data = await response.json()
  if (!data.success) {
    throw new Error('Failed to load companies')
  }
  return data.data
}

const fetchTimeframes = async (company: string): Promise<{ timeframes: string[] }> => {
  const response = await fetch(`/api/companies/${company}/timeframes`)
  const data = await response.json()
  if (!data.success) {
    throw new Error('Failed to load timeframes')
  }
  return data.data
}

const fetchProblems = async ({
  company,
  timeframe,
}: {
  company: string
  timeframe: string
}): Promise<{
  company: string
  timeframe: string
  problems: Problem[]
  count: number
}> => {
  const response = await fetch(`/api/companies/${company}/timeframes/${timeframe}/problems`)
  const data: APIResponse = await response.json()
  if (!data.success) {
    throw new Error(data.error || 'Failed to load problems')
  }
  return data.data!
}


function App() {
  const { theme, toggleTheme } = useTheme()
  const queryClient = useQueryClient()
  const hasClearedCache = useRef(false)

  const [selectedCompany, setSelectedCompany] = useLocalStorage<string>('selectedCompany', '')
  const [selectedTimeframe, setSelectedTimeframe] = useLocalStorage<string>('selectedTimeframe', '')

  // Query for companies - runs once on mount
  const { data: companiesData, error: companiesError } = useQuery({
    queryKey: ['companies'],
    queryFn: fetchCompanies,
    staleTime: 1000 * 60 * 10, // 10 minutes - companies don't change often
  })

  // Query for timeframes - runs when company changes
  const { data: timeframesData, error: timeframesError } = useQuery({
    queryKey: ['timeframes', selectedCompany],
    queryFn: () => fetchTimeframes(selectedCompany),
    enabled: !!selectedCompany,
  })

  // Query for problems - runs when both company and timeframe are selected
  const { data: problemsData, isLoading: problemsLoading, error: problemsError } = useQuery({
    queryKey: ['problems', selectedCompany, selectedTimeframe],
    queryFn: () => fetchProblems({ company: selectedCompany, timeframe: selectedTimeframe }),
    enabled: !!selectedCompany && !!selectedTimeframe,
    retry: (failureCount, error) => {
      // Don't retry if it's a "no problems found" error
      if (error instanceof Error && error.message.includes('No problems found')) {
        return false
      }
      return failureCount < 2
    },
    staleTime: 1000 * 60 * 2, // 2 minutes for problems data
  })

  // Derived state
  const companies = companiesData?.companies || []
  const timeframes = timeframesData?.timeframes || []
  const problems = problemsData?.problems || []
  const error = companiesError?.message || timeframesError?.message || problemsError?.message || ''

  const handleCompanyChange = (company: string) => {
    setSelectedCompany(company)
  }

  const handleTimeframeChange = (timeframe: string) => {
    setSelectedTimeframe(timeframe)
  }

  // Handle timeframe clearing when timeframes change for a company
  useEffect(() => {
    if (selectedCompany && timeframesData) {
      const currentTimeframes = timeframesData.timeframes || []
      if (selectedTimeframe && !currentTimeframes.includes(selectedTimeframe)) {
        setSelectedTimeframe('')
      }
    }
  }, [selectedCompany, timeframesData, selectedTimeframe, setSelectedTimeframe])

  // Clear cache on component mount to prevent stale data issues
  useEffect(() => {
    if (!hasClearedCache.current && (selectedCompany || selectedTimeframe)) {
      hasClearedCache.current = true
      queryClient.invalidateQueries({ queryKey: ['problems'] })
      queryClient.invalidateQueries({ queryKey: ['timeframes'] })
    }
  }, [queryClient, selectedCompany, selectedTimeframe])

  const isProd = import.meta.env.PROD;

  const discordInviteUrl = isProd ? "https://discord.com/oauth2/authorize?client_id=1431162839187460126&permissions=277025736768&integration_type=0&scope=applications.commands+bot" : "https://discord.com/oauth2/authorize?client_id=1431596971767894036&permissions=277025736768&integration_type=0&scope=applications.commands+bot";

  return (
    <div className="min-h-screen" style={{ backgroundColor: 'var(--color-background)' }}>
      <div className="container mx-auto px-4 py-8">
        <header className="mb-8 flex items-start justify-between">
          <div>
            <h1 className="text-4xl font-bold mb-2" style={{ color: 'var(--color-content)' }}>Leetbot.org - a leetcode problem data explorer</h1>
            <p className='max-w-4xl' style={{ color: 'var(--color-tertiary)' }}>See all of the problems that have been asked at your favorite companies. <br /><a href={discordInviteUrl} target="_blank" rel="noopener noreferrer" className="hover:underline text-fuchsia-400">Add leetbot to your discord servers</a> to expose the problems in your own communities.</p>
          </div>
          <ThemeToggle theme={theme} onToggle={toggleTheme} />
        </header>

        <div className="rounded-lg border p-6 mb-8" style={{ backgroundColor: 'var(--color-surface)', borderColor: 'var(--color-muted)' }}>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <CompanySelector
              companies={companies}
              selectedCompany={selectedCompany}
              onCompanyChange={handleCompanyChange}
            />
            <TimeframeSelector
              timeframes={timeframes}
              selectedTimeframe={selectedTimeframe}
              onTimeframeChange={handleTimeframeChange}
              disabled={!selectedCompany}
            />
          </div>
        </div>

        {error && (
          <div className="rounded-lg p-4 mb-6 border" style={{ backgroundColor: '#fef2f2', borderColor: '#fca5a5' }}>
            <p style={{ color: '#dc2626' }}>{error}</p>
          </div>
        )}

        {problemsLoading && <LoadingSpinner />}

        {!problemsLoading && problems.length > 0 && (
          <ProblemsTable problems={problems} />
        )}

        {!problemsLoading && selectedCompany && selectedTimeframe && problems.length === 0 && !error && (
          <div className="rounded-lg p-4 border" style={{ backgroundColor: '#fffbeb', borderColor: '#fbbf24' }}>
            <div className="flex items-center justify-between">
              <div>
                <p style={{ color: '#d97706' }}>No problems found for the selected company and timeframe.</p>
                <p className="text-sm mt-2" style={{ color: '#92400e' }}>
                  This could mean no problems were asked in this timeframe, or the data is still being updated.
                </p>
              </div>
              <button
                onClick={() => {
                  // Clear cache and refetch data
                  queryClient.invalidateQueries({ queryKey: ['problems'] })
                  queryClient.invalidateQueries({ queryKey: ['timeframes'] })
                  queryClient.invalidateQueries({ queryKey: ['companies'] })
                }}
                className="px-3 py-2 text-sm rounded-md border hover:opacity-80"
                style={{
                  backgroundColor: 'var(--color-surface)',
                  borderColor: 'var(--color-muted)',
                  color: 'var(--color-content)'
                }}
                aria-label="Refresh data"
              >
                🔄 Refresh
              </button>
            </div>
          </div>
        )}

        {/* Show loading state for timeframes if company is selected but timeframes are still loading */}
        {selectedCompany && !timeframesData && !timeframesError && (
          <div className="rounded-lg p-4 border" style={{ backgroundColor: 'var(--color-surface)', borderColor: 'var(--color-muted)' }}>
            <p style={{ color: 'var(--color-tertiary)' }}>Loading timeframes...</p>
          </div>
        )}
      </div>
    </div>
  )
}

export default App
