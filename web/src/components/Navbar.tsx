import type { ThemePreference } from '@/hooks/useTheme';
import type { CompanyInfo } from '@/types';
import { Search, X } from 'lucide-react';
import { CompanySelector } from './CompanySelector';
import { DifficultyFilter, type Difficulty } from './DifficultyFilter';
import { SiteBrand } from './SiteBrand';
import { ThemeToggle } from './ThemeToggle';
import { TimeframeSelector } from './TimeframeSelector';
import { Input } from './ui/input';

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
  return (
    <nav className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-backdrop-filter:bg-background/60">
      <div className="container mx-auto px-4">
        <div className="flex h-16 items-center gap-4">
          <div className="flex items-center gap-3 min-w-0">
            <SiteBrand
              discordInviteUrl={discordInviteUrl}
              dataLastUpdated={dataLastUpdated}
              showDataLastUpdated={showDataLastUpdated}
            />
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
