import App from './App'
import { PrivacyPage } from './Privacy'

const isPrivacyRoute = () => window.location.pathname === '/privacy'

export const Root = () => {
  return isPrivacyRoute() ? <PrivacyPage /> : <App />
}
