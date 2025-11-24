import { useQuery, useQueryClient } from '@tanstack/react-query'
import { RefreshCw } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { CompanySelector } from './components/CompanySelector'
import { LoadingSpinner } from './components/LoadingSpinner'
import { ProblemsTable } from './components/ProblemsTable'
import { ThemeToggle } from './components/ThemeToggle'
import { TimeframeSelector } from './components/TimeframeSelector'
import { Alert, AlertDescription } from './components/ui/alert'
import { Button } from './components/ui/button'
import { Card, CardContent } from './components/ui/card'
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

  const [selectedCompany, setSelectedCompany] = useLocalStorage<string>('selectedCompany', 'google')
  const [selectedTimeframe, setSelectedTimeframe] = useLocalStorage<string>('selectedTimeframe', 'all')
  const [previewCompany, setPreviewCompany] = useState<string>('')

  // Query for companies - runs once on mount
  const { data: companiesData, error: companiesError } = useQuery({
    queryKey: ['companies'],
    queryFn: fetchCompanies,
    staleTime: 1000 * 60 * 10, // 10 minutes - companies don't change often
  })

  // Use preview company for fetching if available, otherwise use selected company
  const activeCompany = previewCompany || selectedCompany

  // Query for timeframes - runs when company changes (including preview)
  const { data: timeframesData, error: timeframesError } = useQuery({
    queryKey: ['timeframes', activeCompany],
    queryFn: () => fetchTimeframes(activeCompany),
    enabled: !!activeCompany,
  })

  // Query for problems - runs when both company and timeframe are selected
  // Only use preview company if timeframe is also selected
  const activeTimeframe = selectedTimeframe
  const { data: problemsData, isLoading: problemsLoading, error: problemsError } = useQuery({
    queryKey: ['problems', activeCompany, activeTimeframe],
    queryFn: () => fetchProblems({ company: activeCompany, timeframe: activeTimeframe }),
    enabled: !!activeCompany && !!activeTimeframe,
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
    setPreviewCompany('')
  }

  const handleCompanyPreview = (company: string) => {
    if (company) {
      setPreviewCompany(company)
    } else {
      setPreviewCompany('')
    }
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
    <div className="min-h-screen bg-background">
      <div className="container mx-auto px-4 py-8">
        <header className="mb-8 flex items-start justify-between">
          <div>
            <h1 className="text-4xl font-bold mb-2 text-foreground">Leetbot.org - a leetcode problem data explorer</h1>
            <p className='max-w-4xl text-muted-foreground'>See all of the problems that have been asked at your favorite companies. <br /><a href={discordInviteUrl} target="_blank" rel="noopener noreferrer" className="hover:underline text-fuchsia-400">Add leetbot to your discord servers</a> to expose the problems in your own communities.</p>
          </div>
          <ThemeToggle theme={theme} onToggle={toggleTheme} />
        </header>

        <Card className="mb-8">
          <CardContent className="pt-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <CompanySelector
                companies={companies}
                selectedCompany={selectedCompany}
                onCompanyChange={handleCompanyChange}
                onCompanyPreview={handleCompanyPreview}
              />
              <TimeframeSelector
                timeframes={timeframes}
                selectedTimeframe={selectedTimeframe}
                onTimeframeChange={handleTimeframeChange}
                disabled={!selectedCompany}
              />
            </div>
          </CardContent>
        </Card>

        {error && (
          <Alert variant="destructive" className="mb-6">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <div className="min-h-[600px]">
          {problemsLoading && <LoadingSpinner />}

          {!problemsLoading && problems.length > 0 && (
            <ProblemsTable problems={problems} />
          )}

          {!problemsLoading && selectedCompany && selectedTimeframe && problems.length === 0 && !error && (
            <Alert className="mb-6">
              <div className="flex items-center justify-between w-full">
                <div>
                  <AlertDescription>
                    No problems found for the selected company and timeframe.
                  </AlertDescription>
                  <p className="text-sm mt-2 text-muted-foreground">
                    This could mean no problems were asked in this timeframe, or the data is still being updated.
                  </p>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    queryClient.invalidateQueries({ queryKey: ['problems'] })
                    queryClient.invalidateQueries({ queryKey: ['timeframes'] })
                    queryClient.invalidateQueries({ queryKey: ['companies'] })
                  }}
                  aria-label="Refresh data"
                >
                  <RefreshCw className="h-4 w-4 mr-2" />
                  Refresh
                </Button>
              </div>
            </Alert>
          )}
        </div>

        {selectedCompany && !timeframesData && !timeframesError && (
          <Card className="mb-6">
            <CardContent className="pt-6">
              <div className="flex items-center gap-2">
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent" />
                <p className="text-sm text-muted-foreground">Loading timeframes...</p>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
}

export default App
