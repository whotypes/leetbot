import { useEffect, useState } from 'react'
import { CompanySelector } from './components/CompanySelector'
import { LoadingSpinner } from './components/LoadingSpinner'
import { ProblemsTable } from './components/ProblemsTable'
import { ThemeToggle } from './components/ThemeToggle'
import { TimeframeSelector } from './components/TimeframeSelector'
import { useTheme } from './hooks/useTheme'

interface Problem {
  id: number
  url: string
  title: string
  difficulty: string
  acceptance: number
  frequency: number
}

interface APIResponse {
  success: boolean
  data?: {
    company: string
    timeframe: string
    problems: Problem[]
    count: number
  }
  error?: string
}

function App() {
  const { theme, toggleTheme } = useTheme()
  const [companies, setCompanies] = useState<string[]>([])
  const [selectedCompany, setSelectedCompany] = useState<string>(() => {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('selectedCompany') || ''
    }
    return ''
  })
  const [timeframes, setTimeframes] = useState<string[]>([])
  const [selectedTimeframe, setSelectedTimeframe] = useState<string>(() => {
    if (typeof window !== 'undefined') {
      return localStorage.getItem('selectedTimeframe') || ''
    }
    return ''
  })
  const [problems, setProblems] = useState<Problem[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string>('')

  const handleCompanyChange = (company: string) => {
    setSelectedCompany(company)
    if (typeof window !== 'undefined') {
      localStorage.setItem('selectedCompany', company)
    }
  }

  const handleTimeframeChange = (timeframe: string) => {
    setSelectedTimeframe(timeframe)
    if (typeof window !== 'undefined') {
      localStorage.setItem('selectedTimeframe', timeframe)
    }
  }

  useEffect(() => {
    const loadCompanies = async () => {
      try {
        const response = await fetch('/api/companies')
        const data = await response.json()
        if (data.success) {
          setCompanies(data.data.companies)
        } else {
          setError('Failed to load companies')
        }
      } catch {
        setError('Failed to load companies')
      }
    }
    loadCompanies()
  }, [])

  useEffect(() => {
    if (selectedCompany) {
      const loadTimeframes = async () => {
        try {
          const response = await fetch(`/api/companies/${selectedCompany}/timeframes`)
          const data = await response.json()
          if (data.success) {
            setTimeframes(data.data.timeframes)
            setProblems([])

            // only clear timeframe if it's not available for the new company
            if (typeof window !== 'undefined') {
              const savedTimeframe = localStorage.getItem('selectedTimeframe')
              if (savedTimeframe && !data.data.timeframes.includes(savedTimeframe)) {
                setSelectedTimeframe('')
                localStorage.removeItem('selectedTimeframe')
              }
            } else {
              setSelectedTimeframe('')
            }
          } else {
            setError('Failed to load timeframes')
          }
        } catch {
          setError('Failed to load timeframes')
        }
      }
      loadTimeframes()
    }
  }, [selectedCompany])

  useEffect(() => {
    if (selectedCompany && selectedTimeframe) {
      const loadProblems = async () => {
        setLoading(true)
        setError('')
        try {
          const response = await fetch(`/api/companies/${selectedCompany}/timeframes/${selectedTimeframe}/problems`)
          const data: APIResponse = await response.json()
          if (data.success && data.data) {
            setProblems(data.data.problems)
          } else {
            setError(data.error || 'Failed to load problems')
          }
        } catch {
          setError('Failed to load problems')
        } finally {
          setLoading(false)
        }
      }
      loadProblems()
    }
  }, [selectedCompany, selectedTimeframe])

  return (
    <div className="min-h-screen" style={{ backgroundColor: 'var(--color-background)' }}>
      <div className="container mx-auto px-4 py-8">
        <header className="mb-8 flex items-start justify-between">
          <div>
            <h1 className="text-4xl font-bold mb-2" style={{ color: 'var(--color-content)' }}>Leetbot.org - a leetcode problem data explorer</h1>
            <p className='max-w-4xl' style={{ color: 'var(--color-tertiary)' }}>See all of the problems that have been asked at your favorite companies. <br /><a href="https://discord.com/oauth2/authorize?client_id=1431162839187460126&permissions=2147559424&integration_type=0&scope=bot+applications.commands" target="_blank" rel="noopener noreferrer" className="hover:underline text-fuchsia-400">Add leetbot to your discord servers</a> to expose the problems in your own communities.</p>
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

        {loading && <LoadingSpinner />}

        {!loading && problems.length > 0 && (
          <ProblemsTable problems={problems} />
        )}

        {!loading && selectedCompany && selectedTimeframe && problems.length === 0 && !error && (
          <div className="rounded-lg p-4 border" style={{ backgroundColor: '#fffbeb', borderColor: '#fbbf24' }}>
            <p style={{ color: '#d97706' }}>No problems found for the selected company and timeframe.</p>
          </div>
        )}
      </div>
    </div>
  )
}

export default App
