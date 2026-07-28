import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'
import fs from 'node:fs'
import path from 'node:path'

export default defineConfig({
    base: '/app/',
    plugins: [react()],
    build: {
        outDir: '../web/app',
        emptyOutDir: true,
    },
    server: {
        host: '0.0.0.0',
        port: 5173,

        https: {
            key: fs.readFileSync(
                path.resolve(__dirname, '.cert/dev-key.pem'),
            ),
            cert: fs.readFileSync(
                path.resolve(__dirname, '.cert/dev-cert.pem'),
            ),
        },

        proxy: {
            '/api': {
                target: 'http://localhost:8080',
                changeOrigin: true,
                secure: false,
            },
        },
    },
})
