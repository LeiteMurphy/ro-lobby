import { readFileSync } from 'node:fs';
import { defineConfig, devices } from '@playwright/test';

// RN-22 (fundação): o teste ponta a ponta sobe o backend e o web. O PostgreSQL precisa
// estar no ar (docker compose up -d); o banco do ponta a ponta é criado aqui (ver abaixo).
//
// Spec login-discord (RNF-04, D-05): o login roda contra um Discord falso local, sem rede
// externa. A API e o web recebem as URLs do falso, e nenhum servidor é reaproveitado,
// para não falar com o Discord de verdade por engano.
const apiPort = process.env.API_PORT || '8080';
const apiBaseUrl = `http://localhost:${apiPort}`;
const webPort = 4173;
const fakeDiscord = 'http://127.0.0.1:8090';
const discordApp = { clientId: '000000000000000000', clientSecret: 'segredo-do-discord-falso-e2e' };

// Spec lobbies (T-10): o ponta a ponta usa um banco só dele, ro_lobby_e2e, apagado e migrado
// a cada execução (go run ./cmd/migrate fresh, que só aceita bancos *_e2e). Assim os testes
// não dependem do que ficou no banco de desenvolvimento nem o sujam.
function databaseUrl(): string {
	if (process.env.DATABASE_URL) return process.env.DATABASE_URL;
	try {
		const line = readFileSync('../.env', 'utf8')
			.split(/\r?\n/)
			.find((l) => l.startsWith('DATABASE_URL='));
		if (line) return line.slice('DATABASE_URL='.length).trim();
	} catch {
		// sem .env: usa o padrão do README
	}
	return 'postgres://ro_lobby:ro_lobby_dev@localhost:5432/ro_lobby?sslmode=disable';
}
const e2eDatabase = new URL(databaseUrl());
e2eDatabase.pathname = '/ro_lobby_e2e';
export const E2E_DATABASE_URL = e2eDatabase.toString();

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
			command: 'go run ./cmd/migrate fresh && go run ./cmd/api',
			cwd: '../backend',
			env: {
				DATABASE_URL: E2E_DATABASE_URL,
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
