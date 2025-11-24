export interface Problem {
  id: number
  url: string
  title: string
  difficulty: string
  acceptance: number
  frequency: number
}

export interface APIResponse {
  success: boolean
  data?: {
    company: string
    timeframe: string
    problems: Problem[]
    count: number
  }
  error?: string
}

export type AllProblemsData = Record<string, Record<string, Problem[]>>
