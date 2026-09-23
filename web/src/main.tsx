import { createRoot } from 'react-dom/client'
import './index.css'
import { Providers } from './providers'
import { Root } from './Root'

createRoot(document.getElementById('root')!).render(
  <Providers>
    <Root />
  </Providers>,
)