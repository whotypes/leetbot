import App from './App'
import { PrivacyPage } from './Privacy'
import { TosPage } from './Tos'

export const Root = () => {
  const { pathname } = window.location
  if (pathname === '/privacy') return <PrivacyPage />
  if (pathname === '/tos') return <TosPage />
  return <App />
}
