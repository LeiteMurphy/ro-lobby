import { defineConfig, devices } from '@playwright/test';

// RN-22: o teste ponta a ponta sobe o backend e o web. O PostgreSQL precisa estar no
// ar e migrado antes (docker compose up -d e go run ./cmd/migrate up).
const apiPort = process.env.API_PORT || '8080';
const apiBaseUrl = process.env.API_BASE_URL || `http://localhost:${apiPort}`;
const webPort = 4173;

export default defineConfig({
	testDir: 'test/e2e',
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 1 : 0,
	reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'list',
	use: {
		baseURL: `http://localhost:${webPort}`,
		trace: 'retain-on-failure'
	},
	projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
	webServer: [
		{
			command: 'go run ./cmd/api',
			cwd: '../backend',
			// Só fica pronto quando o /healthz responde 200, ou seja, com o banco no ar.
			url: `${apiBaseUrl}/healthz`,
			reuseExistingServer: !process.env.CI,
			timeout: 120_000
		},
		{
			command: `npm run build && npm run preview -- --port ${webPort} --strictPort`,
			port: webPort,
			env: { API_BASE_URL: apiBaseUrl },
			reuseExistingServer: !process.env.CI,
			timeout: 120_000
		}
	]
});
