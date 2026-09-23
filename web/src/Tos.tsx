import { SiteBrand } from '@/components/SiteBrand'
import { ThemeToggle } from '@/components/ThemeToggle'
import { useTheme } from '@/hooks/useTheme'

const isProd = import.meta.env.PROD
const discordInviteUrl = isProd
  ? 'https://discord.com/oauth2/authorize?client_id=1431162839187460126&permissions=277025736768&integration_type=0&scope=applications.commands+bot'
  : 'https://discord.com/oauth2/authorize?client_id=1431596971767894036&permissions=277025736768&integration_type=0&scope=applications.commands+bot'

export const TosPage = () => {
  const { theme, toggleTheme } = useTheme()

  return (
    <div className="flex min-h-screen flex-col bg-background text-foreground">
      <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-backdrop-filter:bg-background/60">
        <div className="container mx-auto px-4">
          <div className="flex h-16 items-center gap-4">
            <div className="flex items-center gap-3 min-w-0">
              <SiteBrand discordInviteUrl={discordInviteUrl} />
            </div>
            <div className="flex-1" />
            <ThemeToggle theme={theme} onToggle={toggleTheme} />
          </div>
        </div>
      </header>

      <main className="container mx-auto flex-1 px-4 py-10">
        <h1 className="text-3xl font-bold tracking-tight">Terms of Service</h1>
        <p className="mt-4 max-w-2xl text-lg leading-relaxed text-muted-foreground">
          Short version: be cool, and use leetbot for what it's for.<br/><br/>
          This website and the Discord bot are free side projects, provided as-is.
          I do my best to keep the data accurate and the lights on, but I can't promise
          either will always be the case.<br/>
          <br/>
          Please don't spam, scrape, or hammer the site or bot, and don't try to break things
          on purpose. If you do, I may block you, no hard feelings.
          <br/>
          <br/>
          Btw, this project isn't affiliated with LeetCode.
        </p>

        <h2 className="mt-10 text-xl font-semibold tracking-tight">.....</h2>
        <p className="mt-3 max-w-2xl text-lg leading-relaxed text-muted-foreground">
          By using leetbot, you're agreeing to these terms and the{' '}
          <a
            href="/privacy"
            className="text-primary underline underline-offset-4 hover:no-underline"
          >
            privacy policy
          </a>
          . I'm not liable for anything that happens from using it, so use your own judgement. If you get caught alpha farming, that's on you.
          <br/>
          <br/>
          I might update these terms now and then. If you keep using leetbot after that, you're good with the changes.
        </p>
      </main>
    </div>
  )
}
