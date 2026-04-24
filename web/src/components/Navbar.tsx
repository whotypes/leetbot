import type { ThemePreference } from '@/hooks/useTheme';
import type { CompanyInfo } from '@/types';
import { Info, Search, X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { CompanySelector } from './CompanySelector';
import { DifficultyFilter, type Difficulty } from './DifficultyFilter';
import { ThemeToggle } from './ThemeToggle';
import { TimeframeSelector } from './TimeframeSelector';
import { Button } from './ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from './ui/dialog';
import { Input } from './ui/input';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from './ui/tooltip';

/** `?about` opens this dialog — e.g. https://leetbot.org/?about */
const ABOUT_QUERY_PARAM = 'about'

/** LeetCode Discuss — community context on company “frequency” ordering and caveats */
const LEETCODE_FREQUENCY_FAQ_URL =
  'https://leetcode.com/discuss/general-discussion/1677842/is-leetcode-ordering-company-questions-by-frequency-incorrectly/'

function stripAboutFromUrl() {
  const url = new URL(window.location.href)
  if (!url.searchParams.has(ABOUT_QUERY_PARAM)) return
  url.searchParams.delete(ABOUT_QUERY_PARAM)
  const qs = url.searchParams.toString()
  const next = `${url.pathname}${qs ? `?${qs}` : ''}${url.hash}`
  window.history.replaceState(null, '', next)
}

interface NavbarProps {
  companies: CompanyInfo[]
  selectedCompany: string
  onCompanyChange: (company: string) => void
  onCompanyPreview: (company: string) => void
  timeframes: string[]
  selectedTimeframe: string
  onTimeframeChange: (timeframe: string) => void
  selectedDifficulties: Difficulty[]
  onDifficultyChange: (difficulties: Difficulty[]) => void
  searchQuery: string
  onSearchChange: (query: string) => void
  theme: ThemePreference
  onThemeToggle: () => void
  discordInviteUrl: string
  dataLastUpdated?: string
  showDataLastUpdated?: boolean
}

const DiscordIcon = ({ className }: { className?: string }) => (
  <svg
    className={className}
    viewBox="0 0 24 24"
    fill="currentColor"
    xmlns="http://www.w3.org/2000/svg"
  >
    <path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028 14.09 14.09 0 0 0 1.226-1.994.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.946 2.418-2.157 2.418z" />
  </svg>
)

export const Navbar = ({
  companies,
  selectedCompany,
  onCompanyChange,
  onCompanyPreview,
  timeframes,
  selectedTimeframe,
  onTimeframeChange,
  selectedDifficulties,
  onDifficultyChange,
  searchQuery,
  onSearchChange,
  theme,
  onThemeToggle,
  discordInviteUrl,
  dataLastUpdated,
  showDataLastUpdated,
}: NavbarProps) => {
  const [aboutOpen, setAboutOpen] = useState(false)

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    if (params.has(ABOUT_QUERY_PARAM)) {
      setAboutOpen(true)
    }
  }, [])

  const handleAboutOpenChange = (open: boolean) => {
    setAboutOpen(open)
    if (!open) {
      stripAboutFromUrl()
    }
  }

  const lastUpdatedLabel =
    showDataLastUpdated && dataLastUpdated
      ? (() => {
        try {
          return new Intl.DateTimeFormat(undefined, {
            dateStyle: 'medium',
            timeZone: 'UTC',
          }).format(new Date(dataLastUpdated))
        } catch {
          return dataLastUpdated
        }
      })()
      : null

  return (
    <nav className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-backdrop-filter:bg-background/60">
      <div className="container mx-auto px-4">
        <div className="flex h-16 items-center gap-4">
          <div className="flex items-center gap-3 min-w-0">
            <TooltipProvider delayDuration={100}>
              <Tooltip>
                <TooltipTrigger asChild>
                  <a
                    href={discordInviteUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-muted-foreground hover:text-fuchsia-400 transition-colors shrink-0"
                    aria-label="Add leetbot to your Discord server"
                    tabIndex={0}
                  >
                    <DiscordIcon className="h-5 w-5" />
                  </a>
                </TooltipTrigger>
                <TooltipContent>
                  <p>Add leetbot to your Discord servers</p>
                </TooltipContent>
              </Tooltip>

              <div className="flex flex-col gap-0.5 min-w-0">
                <div className="flex items-center gap-1 min-w-0">
                  <h1 className="text-xl font-bold text-foreground whitespace-nowrap">
                    leetbot.org
                  </h1>
                  <Dialog open={aboutOpen} onOpenChange={handleAboutOpenChange}>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <DialogTrigger asChild>
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7 shrink-0 text-muted-foreground hover:text-foreground"
                            aria-label="About leetbot"
                          >
                            <Info className="h-4 w-4" />
                          </Button>
                        </DialogTrigger>
                      </TooltipTrigger>
                      <TooltipContent side="bottom">
                        <p>About this site</p>
                      </TooltipContent>
                    </Tooltip>
                    <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
                      <DialogHeader>
                        <DialogTitle>About <span className="text-accent">leetbot.org</span></DialogTitle>
                        <DialogDescription asChild>
                          <div className="space-y-3 pt-1 text-left text-sm text-muted-foreground">
                            <p>
                              Welcome to leetbot! <br /> <br />

                              Thousands of engineers around the world use this <span className="text-accent">always free</span> tool to stay on top of the latest company questions as they are updated on LeetCode.

                              <br /> <br />
                              The main interface of leetbot (hence the name) is a Discord bot that emits problems to a channel, most notably in <a href="https://discord.com/invite/cscareers" target="_blank" rel="noopener noreferrer" className="font-medium text-primary underline underline-offset-4 hover:no-underline">the CSCD community</a>.
                              <br /> <br />

                             I run the scraper every couple of weeks to ensure the data is up to date.
                            </p>
                            <p>
                              All you have to do is pick a company, timeframe, and optional difficulty filters. The problems are
                              ordered like on LeetCode, including the frequency % column used for
                              ranking.
                            </p>
                            <p>
                              <span className="text-accent">About “frequency”</span>{' '}
                              - I get a lot of questions about this. The frequency column reflects LeetCode’s own signals and community tagging, and is not a
                              guarantee of what you will see in an interview. <br/> <br/> For background on how leetcode themselves determine frequency, see this{' '}
                              <a
                                href={LEETCODE_FREQUENCY_FAQ_URL}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="font-medium text-primary underline underline-offset-4 hover:no-underline"
                              >
                                LeetCode discussion thread on frequency ordering
                              </a>
                              .
                            </p>
                          </div>
                        </DialogDescription>
                      </DialogHeader>
                    </DialogContent>
                  </Dialog>
                </div>
                {lastUpdatedLabel && (
                  <p className="text-xs text-accent truncate" title={dataLastUpdated}>
                    Last updated: {lastUpdatedLabel}
                  </p>
                )}
              </div>
            </TooltipProvider>
          </div>

          <div className="hidden sm:flex flex-1 items-center gap-3">
            <div className="w-48">
              <CompanySelector
                companies={companies}
                selectedCompany={selectedCompany}
                onCompanyChange={onCompanyChange}
                onCompanyPreview={onCompanyPreview}
                compact
              />
            </div>

            <div className="w-40">
              <TimeframeSelector
                timeframes={timeframes}
                selectedTimeframe={selectedTimeframe}
                onTimeframeChange={onTimeframeChange}
                disabled={!selectedCompany}
                compact
              />
            </div>

            <div className="hidden lg:block">
              <DifficultyFilter
                selectedDifficulties={selectedDifficulties}
                onDifficultyChange={onDifficultyChange}
              />
            </div>

            <div className="hidden lg:block relative flex-1 max-w-xs">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                type="text"
                placeholder="Search problems..."
                value={searchQuery}
                onChange={(e) => onSearchChange(e.target.value)}
                className="h-9 pl-9 pr-9"
                aria-label="Search problems"
              />
              {searchQuery && (
                <button
                  type="button"
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  onClick={() => onSearchChange('')}
                  aria-label="Clear search"
                >
                  <X className="h-4 w-4" />
                </button>
              )}
            </div>
          </div>

          <div className="flex-1 sm:hidden" />

          <ThemeToggle theme={theme} onToggle={onThemeToggle} />
        </div>
      </div>
    </nav>
  )
}
