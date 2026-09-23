import { SiteBrand } from '@/components/SiteBrand'
import { ThemeToggle } from '@/components/ThemeToggle'
import { useTheme } from '@/hooks/useTheme'

const isProd = import.meta.env.PROD
const discordInviteUrl = isProd
  ? 'https://discord.com/oauth2/authorize?client_id=1431162839187460126&permissions=277025736768&integration_type=0&scope=applications.commands+bot'
  : 'https://discord.com/oauth2/authorize?client_id=1431596971767894036&permissions=277025736768&integration_type=0&scope=applications.commands+bot'

export const PrivacyPage = () => {
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
        <h1 className="text-3xl font-bold tracking-tight">Privacy Policy</h1>
        <p className="mt-4 max-w-2xl text-lg leading-relaxed text-muted-foreground">
          We don't store what you type into leetbot.org.<br/><br/>
          Data queries all stay in your browser for the moment you're on the page,
          and it's gone the second you leave.<br/> 
          <br/>
          Hence why we have no accounts, no profiles, or sessions.
          And since we have nothing, there's nothing to sell/give to anyone.
          
          <br/>
          <br/>
          We DO collect anonymous analytics on page views, historical patterns, and share of volume. These do not identify any individual user and are purely for improving the site! 
        </p>
        
        

        <h2 className="mt-10 text-xl font-semibold tracking-tight">
          The{' '}
          <a
            href={discordInviteUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="text-primary underline underline-offset-4 hover:no-underline"
          >
            Discord
          </a>{' '}
          bot
        </h2>
        <p className="mt-3 max-w-2xl text-lg leading-relaxed text-muted-foreground">
          Note, that when you use leetbot in Discord, the bot reads the content of the messages
          you send it so it can do its job. 
          <br/>
            <br/> Message content is processed on
          the spot and gone
          once the bot responds. The page turns for responses aren't stateless, so you can flip through them to your heart's content.
        </p>
      </main>
    </div>
  )
}
