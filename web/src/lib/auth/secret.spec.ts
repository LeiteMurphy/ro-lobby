import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

// CA-01.5 / RN-03 (login-discord): o Client Secret existe só no ambiente da API. O código
// do web nunca lê essa variável; a CI ainda confere o build com um secret marcador.
describe('Client Secret fora do web', () => {
	const src = fileURLToPath(new URL('../..', import.meta.url));
	const walk = (dir: string): string[] =>
		readdirSync(dir).flatMap((name) => {
			const path = join(dir, name);
			return statSync(path).isDirectory() ? walk(path) : [path];
		});

	it('CA-01.5 / RN-03: nenhum arquivo do web lê DISCORD_CLIENT_SECRET', () => {
		const offenders = walk(src)
			.filter((f) => !f.endsWith('.spec.ts'))
			.filter((f) => readFileSync(f, 'utf8').includes('DISCORD_CLIENT_SECRET'));
		expect(offenders).toEqual([]);
	});
});
