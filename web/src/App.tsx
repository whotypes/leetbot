import { QueryClient, useQuery, useQueryClient } from '@tanstack/react-query'
import { RefreshCw } from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import type { Difficulty } from './components/DifficultyFilter'
import { Navbar } from './components/Navbar'
import { ProblemsDataTable } from './components/ProblemsDataTable'
import { Alert, AlertDescription } from './components/ui/alert'
import { Button } from './components/ui/button'
import { useLocalStorage } from './hooks/useLocalStorage'
import { useTheme } from './hooks/useTheme'
import type { AllProblemsData, APIResponse, CompanyInfo, Problem } from './types'

const formatSlugAsDisplayName = (slug: string) =>
  slug.charAt(0).toUpperCase() + slug.slice(1).replace(/-/g, ' ')

/** Accepts both catalog objects and legacy string[] from older API builds. */
function normalizeCompaniesPayload(payload: unknown): {
  dataLastUpdated: string
  companies: CompanyInfo[]
} {
  if (!payload || typeof payload !== 'object') {
    return { dataLastUpdated: '', companies: [] }
  }
  const p = payload as { dataLastUpdated?: unknown; companies?: unknown }
  const updated =
    typeof p.dataLastUpdated === 'string' ? p.dataLastUpdated : ''
  const raw = p.companies
  if (!Array.isArray(raw)) {
    return { dataLastUpdated: updated, companies: [] }
  }
  const companies: CompanyInfo[] = []
  for (const item of raw) {
    if (typeof item === 'string') {
      const slug = item.toLowerCase().trim()
      if (!slug) continue
      companies.push({
        slug,
        name: formatSlugAsDisplayName(slug),
        hasData: true,
      })
      continue
    }
    if (item && typeof item === 'object') {
      const o = item as Partial<CompanyInfo>
      const slug = String(o.slug ?? '')
        .toLowerCase()
        .trim()
      if (!slug) continue
      const rawName = o.name != null ? String(o.name).trim() : ''
      companies.push({
        slug,
        name: rawName || formatSlugAsDisplayName(slug),
        hasData: typeof o.hasData === 'boolean' ? o.hasData : true,
      })
    }
  }
  return { dataLastUpdated: updated, companies }
}

const fetchCompanies = async (): Promise<{
  dataLastUpdated: string
  companies: CompanyInfo[]
}> => {
  const response = await fetch('/api/companies')
  const data = await response.json()
  if (!data.success) {
    throw new Error('Failed to load companies')
  }
  return normalizeCompaniesPayload(data.data)
}

const fetchTimeframes = async (company: string): Promise<{ timeframes: string[] }> => {
  const response = await fetch(`/api/companies/${company}/timeframes`)
  const data = await response.json()
  if (!data.success) {
    throw new Error('Failed to load timeframes')
  }
  return data.data
}

const normalizeTimeframe = (timeframe: string): string => {
  const normalized = timeframe.toLowerCase().trim().replace(/\s+/g, '-')
  const mapping: Record<string, string> = {
    '30': 'thirty-days',
    '30days': 'thirty-days',
    '30-days': 'thirty-days',
    'thirty': 'thirty-days',
    'thirtydays': 'thirty-days',
    'thirty-days': 'thirty-days',
    '30d': 'thirty-days',
    '90': 'three-months',
    '90days': 'three-months',
    '90-days': 'three-months',
    'three': 'three-months',
    'threemonths': 'three-months',
    'three-months': 'three-months',
    '3months': 'three-months',
    '3-months': 'three-months',
    '3mo': 'three-months',
    '90d': 'three-months',
    '180': 'six-months',
    '180days': 'six-months',
    '180-days': 'six-months',
    'six': 'six-months',
    'sixmonths': 'six-months',
    'six-months': 'six-months',
    '6months': 'six-months',
    '6-months': 'six-months',
    '6mo': 'six-months',
    'all': 'all',
    'alltime': 'all',
    'all-time': 'all',
    'everything': 'all',
    'more-than-six-months': 'more-than-six-months',
    'morethan6months': 'more-than-six-months',
    'more-than-6-months': 'more-than-six-months',
    '>6mo': 'more-than-six-months',
    '>6months': 'more-than-six-months',
  }
  return mapping[normalized] || 'all'
}

const fetchAllProblems = async (): Promise<AllProblemsData> => {
  const response = await fetch('/api/all-problems')
  const data = await response.json()
  if (!data.success) {
    throw new Error('Failed to load all problems')
  }
  return data.data
}

const fetchProblems = async ({
  company,
  timeframe,
  queryClient,
}: {
  company: string
  timeframe: string
  queryClient: QueryClient
}): Promise<{
  company: string
  timeframe: string
  problems: Problem[]
  count: number
  companyHasLocalData?: boolean
  emptyTimeframe?: boolean
}> => {
  const normalizedCompany = company.toLowerCase().trim()
  const normalizedTimeframe = normalizeTimeframe(timeframe)
  const catalog = queryClient.getQueryData<{
    dataLastUpdated: string
    companies: CompanyInfo[]
  }>(['companies'])
  const companyMeta = catalog?.companies.find((c) => c.slug === normalizedCompany)

  const allProblemsData = queryClient.getQueryData<AllProblemsData>(['all-problems'])

  if (allProblemsData) {
    const companyData = allProblemsData[normalizedCompany]
    if (companyData) {
      const problems = companyData[normalizedTimeframe]
      if (problems !== undefined) {
        return {
          company: normalizedCompany,
          timeframe: normalizedTimeframe,
          problems,
          count: problems.length,
          companyHasLocalData: companyMeta?.hasData,
          emptyTimeframe:
            companyMeta?.hasData && problems.length === 0 ? true : undefined,
        }
      }
    }
  }

  const response = await fetch(
    `/api/companies/${encodeURIComponent(company)}/timeframes/${encodeURIComponent(timeframe)}/problems`
  )
  const data: APIResponse = await response.json()
  if (!data.success) {
    throw new Error(data.error || 'Failed to load problems')
  }
  return {
    company: data.data!.company,
    timeframe: data.data!.timeframe,
    problems: data.data!.problems,
    count: data.data!.count,
    companyHasLocalData: data.data!.companyHasLocalData,
    emptyTimeframe: data.data!.emptyTimeframe,
  }
}

function App() {
  const { theme, toggleTheme } = useTheme()
  const queryClient = useQueryClient()
  const hasClearedCache = useRef(false)

  const [selectedCompany, setSelectedCompany] = useLocalStorage<string>('selectedCompany', 'google')
  const [selectedTimeframe, setSelectedTimeframe] = useLocalStorage<string>('selectedTimeframe', 'all')
  const [previewCompany, setPreviewCompany] = useState<string>('')
  const [selectedDifficulties, setSelectedDifficulties] = useState<Difficulty[]>([])
  const [searchQuery, setSearchQuery] = useState<string>('')

  useQuery({
    queryKey: ['all-problems'],
    queryFn: fetchAllProblems,
    staleTime: 1000 * 60 * 60,
    gcTime: 1000 * 60 * 60 * 24,
    refetchOnMount: false,
    refetchOnWindowFocus: false,
  })

  const { data: companiesData, error: companiesError } = useQuery({
    queryKey: ['companies'],
    queryFn: fetchCompanies,
    staleTime: 1000 * 60 * 10,
  })

  const activeCompany = previewCompany || selectedCompany

  const { data: timeframesData, error: timeframesError } = useQuery({
    queryKey: ['timeframes', activeCompany],
    queryFn: () => fetchTimeframes(activeCompany),
    enabled: !!activeCompany,
  })

  const activeTimeframe = selectedTimeframe
  const { data: problemsData, isLoading: problemsLoading, error: problemsError } = useQuery({
    queryKey: ['problems', activeCompany, activeTimeframe],
    queryFn: () => fetchProblems({ company: activeCompany, timeframe: activeTimeframe, queryClient }),
    enabled: !!activeCompany && !!activeTimeframe,
    retry: (failureCount, error) => {
      if (error instanceof Error && error.message.includes('No problems found')) {
        return false
      }
      return failureCount < 2
    },
    staleTime: 1000 * 60 * 2,
  })

  const companies = companiesData?.companies || []
  const dataLastUpdated = companiesData?.dataLastUpdated
  const timeframes = timeframesData?.timeframes || []
  const problems = problemsData?.problems || []
  const error = companiesError?.message || timeframesError?.message || problemsError?.message || ''

  const activeCompanyInfo = useMemo(
    () => companies.find((c) => c.slug === activeCompany),
    [companies, activeCompany]
  )
  const noLocalData =
    (activeCompanyInfo && activeCompanyInfo.hasData === false) ||
    (problemsData && problemsData.companyHasLocalData === false)
  const showLastUpdatedLine =
    typeof dataLastUpdated === 'string' &&
    dataLastUpdated.length > 0 &&
    new Date(dataLastUpdated).getUTCFullYear() >= 2015

  const filteredProblems = useMemo(() => {
    let result = problems

    if (selectedDifficulties.length > 0) {
      result = result.filter((p) =>
        selectedDifficulties.includes(p.difficulty.toLowerCase() as Difficulty)
      )
    }

    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase().trim()
      result = result.filter(
        (p) =>
          p.title.toLowerCase().includes(query) ||
          p.id.toString().includes(query)
      )
    }

    return result
  }, [problems, selectedDifficulties, searchQuery])

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

  useEffect(() => {
    if (selectedCompany && timeframesData) {
      const currentTimeframes = timeframesData.timeframes || []
      if (selectedTimeframe && !currentTimeframes.includes(selectedTimeframe)) {
        const fallback = currentTimeframes.includes('all')
          ? 'all'
          : currentTimeframes[0] || 'all'
        setSelectedTimeframe(fallback)
      }
    }
  }, [selectedCompany, timeframesData, selectedTimeframe, setSelectedTimeframe])

  useEffect(() => {
    if (!hasClearedCache.current && (selectedCompany || selectedTimeframe)) {
      hasClearedCache.current = true
      queryClient.invalidateQueries({ queryKey: ['problems'] })
      queryClient.invalidateQueries({ queryKey: ['timeframes'] })
    }
  }, [queryClient, selectedCompany, selectedTimeframe])

  const isProd = import.meta.env.PROD
  const discordInviteUrl = isProd
    ? 'https://discord.com/oauth2/authorize?client_id=1431162839187460126&permissions=277025736768&integration_type=0&scope=applications.commands+bot'
    : 'https://discord.com/oauth2/authorize?client_id=1431596971767894036&permissions=277025736768&integration_type=0&scope=applications.commands+bot'

  return (
    <div className="flex h-screen flex-col bg-background">
      <Navbar
        companies={companies}
        selectedCompany={selectedCompany}
        onCompanyChange={handleCompanyChange}
        onCompanyPreview={handleCompanyPreview}
        timeframes={timeframes}
        selectedTimeframe={selectedTimeframe}
        onTimeframeChange={handleTimeframeChange}
        selectedDifficulties={selectedDifficulties}
        onDifficultyChange={setSelectedDifficulties}
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        theme={theme}
        onThemeToggle={toggleTheme}
        discordInviteUrl={discordInviteUrl}
        dataLastUpdated={dataLastUpdated}
        showDataLastUpdated={!!showLastUpdatedLine}
      />

      <main className="flex-1 overflow-hidden">
        {error && (
          <div className="px-4 py-3">
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          </div>
        )}

        {(problemsLoading || filteredProblems.length > 0) && (
          <ProblemsDataTable
            problems={filteredProblems}
            isLoading={problemsLoading}
            companies={companies}
            selectedCompany={selectedCompany}
            onCompanyChange={handleCompanyChange}
            timeframes={timeframes}
            selectedTimeframe={selectedTimeframe}
            onTimeframeChange={handleTimeframeChange}
          />
        )}

        {!problemsLoading && selectedCompany && selectedTimeframe && filteredProblems.length === 0 && !error && (
          <div className="flex h-full items-center justify-center px-4">
            <Alert className="max-w-md">
              <div className="flex items-center justify-between gap-4">
                <div>
                  <AlertDescription>
                    {noLocalData
                      ? "There's no leetbot data for this company yet."
                      : problems.length === 0
                        ? problemsData?.emptyTimeframe
                          ? 'No problems for this company in the selected timeframe in my dataset.'
                          : 'No problems found for the selected company and timeframe.'
                        : 'No problems match your current filters.'}
                  </AlertDescription>
                  <p className="text-sm mt-2 text-muted-foreground">
                    {noLocalData
                      ? 'Companies still appear in the list when LeetCode has a company page, even if there are no problems for that company yet.'
                      : problems.length === 0
                        ? 'This can mean the timeframe is empty, or the data is still being updated.'
                        : 'Try adjusting your difficulty or search filters.'}
                  </p>
                </div>
                {problems.length === 0 && !noLocalData && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      queryClient.invalidateQueries({ queryKey: ['problems'] })
                      queryClient.invalidateQueries({ queryKey: ['timeframes'] })
                      queryClient.invalidateQueries({ queryKey: ['companies'] })
                      queryClient.invalidateQueries({ queryKey: ['all-problems'] })
                    }}
                    aria-label="Refresh data"
                  >
                    <RefreshCw className="h-4 w-4 mr-2" />
                    Refresh
                  </Button>
                )}
              </div>
            </Alert>
          </div>
        )}

        {selectedCompany && !timeframesData && !timeframesError && !problemsLoading && (
          <div className="flex h-full items-center justify-center">
            <div className="flex items-center gap-2">
              <div className="h-4 w-4 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent" />
              <p className="text-sm text-muted-foreground">Loading timeframes...</p>
            </div>
          </div>
        )}
      </main>
    </div>
  )
}

export default App
