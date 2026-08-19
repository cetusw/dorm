import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'

import '@mantine/core/styles.css'
import '@mantine/dates/styles.css'

import { App } from './app/App'
import { AppProviders } from './app/providers'
import { registerServiceWorker } from './shared/lib/service-worker/registerServiceWorker'
import './shared/styles/globals.css'

void registerServiceWorker()

document.getElementById('app-preloader')?.remove()

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <AppProviders>
            <App />
        </AppProviders>
    </StrictMode>,
)
