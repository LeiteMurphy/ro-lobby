import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';

const read = (path: string) => readFileSync(new URL(path, import.meta.url), 'utf8');

describe('configuração do web', () => {
	it('RN-14: usa o adaptador de servidor (Node), não o estático', () => {
		const viteConfig = read('../vite.config.ts');
		expect(viteConfig).toMatch(/from '@sveltejs\/adapter-node'/);
		expect(viteConfig).not.toMatch(/adapter-static/);
	});

	it('RN-16 / CA-06.8: o TypeScript roda em modo estrito', () => {
		const tsconfig = read('../tsconfig.json').replace(/^\s*\/\/.*$/gm, '');
		expect(JSON.parse(tsconfig).compilerOptions.strict).toBe(true);
	});

	it('RN-04: a versão do Node em engines bate com o .nvmrc', () => {
		const nvmrc = read('../../.nvmrc').trim();
		const engines: string = JSON.parse(read('../package.json')).engines.node;
		expect(engines).toContain(`>=${nvmrc}`);
		expect(engines).toContain(`<${Number(nvmrc.split('.')[0]) + 1}`);
	});
});
