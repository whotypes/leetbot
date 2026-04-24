export interface Problem {
  id: number
  url: string
  title: string
  difficulty: string
  acceptance: number
  frequency: number
}

export interface CompanyInfo {
  slug: string
  name: string
  hasData: boolean
}

export interface APIResponse {
  success: boolean
  data?: {
    company: string
    timeframe: string
    problems: Problem[]
    count: number
    companyHasLocalData?: boolean
    emptyTimeframe?: boolean
  }
  error?: string
}

export type AllProblemsData = Record<string, Record<string, Problem[]>>
