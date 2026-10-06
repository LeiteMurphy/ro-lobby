import { defineConfig, devices } from '@playwright/test';

// RN-22 (fundação): o teste ponta a ponta sobe o backend e o web. O PostgreSQL precisa
// estar no ar e migrado antes (docker compose up -d e go run ./cmd/migrate up).
//
// Spec login-discord (RNF-04, D-05): o login roda contra um Discord falso local, sem rede
// externa. A API e o web recebem as URLs do falso, e nenhum servidor é reaproveitado,
// para não falar com o Discord de verdade por engano.
const apiPort = process.env.API_PORT || '8080';
const apiBaseUrl = `http://localhost:${apiPort}`;
const webPort = 4173;
const fakeDiscord = 'http://127.0.0.1:8090';
const discordApp = { clientId: '000000000000000000', clientSecret: 'segredo-do-discord-falso-e2e' };

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
			command: `go run ./cmd/fakediscord -addr 127.0.0.1:8090 -client-id ${discordApp.clientId} -client-secret ${discordApp.clientSecret}`,
			cwd: '../backend',
			// Sem parâmetros, a página de autorização responde 400, o que já conta como no ar.
			url: `${fakeDiscord}/oauth2/authorize`,
			reuseExistingServer: false,
			timeout: 120_000
		},
		{
			command: 'go run ./cmd/api',
			cwd: '../backend',
			env: {
				DISCORD_CLIENT_ID: discordApp.clientId,
				DISCORD_CLIENT_SECRET: discordApp.clientSecret,
				DISCORD_API_BASE_URL: `${fakeDiscord}/api`
			},
			// Só fica pronto quando o /healthz responde 200, ou seja, com o banco no ar.
			url: `${apiBaseUrl}/healthz`,
			reuseExistingServer: false,
			timeout: 120_000
		},
		{
			command: `npm run build && npm run preview -- --port ${webPort} --strictPort`,
			port: webPort,
			env: {
				API_BASE_URL: apiBaseUrl,
				DISCORD_CLIENT_ID: discordApp.clientId,
				DISCORD_AUTHORIZE_URL: `${fakeDiscord}/oauth2/authorize`
			},
			reuseExistingServer: false,
			timeout: 120_000
		}
	]
});
