import { loadEnv } from 'vite';
import { defineConfig } from 'vitest/config';
import adapter from '@sveltejs/adapter-node';
import { sveltekit } from '@sveltejs/kit/vite';

// D-07: um único .env na raiz do repositório serve ao Compose, ao backend e ao web.
const envDir = '..';

export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, envDir, '');

	return {
		plugins: [
			sveltekit({
				compilerOptions: {
					// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
					runes: ({ filename }) =>
						filename.split(/[/\\]/).includes('node_modules') ? undefined : true
				},
				// RN-14: adaptador de servidor (Node), porque o preview no Discord exige SSR (ADR-01).
				adapter: adapter(),
				env: { dir: envDir }
			})
		],
		server: {
			port: Number(env.WEB_PORT) || 5173,
			strictPort: true
		},
		test: {
			expect: { requireAssertions: true },
			// RN-21: relatório de cobertura sem percentual mínimo.
			coverage: {
				provider: 'v8',
				reporter: ['text-summary', 'json-summary', 'html'],
				reportsDirectory: 'coverage',
				include: ['src/**/*.{ts,svelte}'],
				exclude: ['src/**/*.spec.ts', 'src/**/*.gen.ts', 'src/**/*.d.ts']
			},
			projects: [
				{
					extends: './vite.config.ts',
					test: {
						name: 'server',
						environment: 'node',
						include: ['src/**/*.{test,spec}.{js,ts}'],
						exclude: ['src/**/*.svelte.{test,spec}.{js,ts}']
					}
				}
			]
		}
	};
});
