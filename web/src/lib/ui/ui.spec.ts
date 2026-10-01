import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { createRawSnippet } from 'svelte';
import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import Button from './Button.svelte';
import Check from './Check.svelte';
import Drawer from './Drawer.svelte';
import IconButton from './IconButton.svelte';
import Select from './Select.svelte';

const text = (value: string) => createRawSnippet(() => ({ render: () => `<span>${value}</span>` }));
const noop = () => {};

describe('Button', () => {
	it('RN-18 / CA-02.5: ação sem backend fica com aria-disabled e a dica "Disponível em breve"', () => {
		const { body } = render(Button, { props: { soon: true, children: text('Criar lobby') } });
		expect(body).toContain('aria-disabled="true"');
		expect(body).toContain('Disponível em breve');
		const describedBy = body.match(/aria-describedby="([^"]+)"/)?.[1];
		expect(describedBy).toBeTruthy();
		expect(body).toContain(`id="${describedBy}"`);
	});

	it('RNF-01: botão comum não fica desabilitado nem mostra a dica', () => {
		const { body } = render(Button, { props: { children: text('Limpar') } });
		expect(body).not.toContain('aria-disabled');
		expect(body).not.toContain('Disponível em breve');
		expect(body).toContain('Limpar');
	});
});

describe('IconButton', () => {
	it('RNF-01: tem rótulo acessível', () => {
		const { body } = render(IconButton, { props: { icon: 'x', label: 'Fechar' } });
		expect(body).toContain('aria-label="Fechar"');
	});

	it('RN-18: versão "em breve" fica com aria-disabled e a dica', () => {
		const { body } = render(IconButton, {
			props: { icon: 'plus', label: 'Criar lobby', soon: true }
		});
		expect(body).toContain('aria-disabled="true"');
		expect(body).toContain('Disponível em breve');
	});
});

describe('Check', () => {
	it('RNF-01: checkbox com rótulo e contagem', () => {
		const { body } = render(Check, {
			props: { type: 'checkbox', label: 'Tank', checked: true, meta: 3, onchange: noop }
		});
		expect(body).toMatch(/<label[^>]*>[\s\S]*type="checkbox"[\s\S]*Tank[\s\S]*3/);
		expect(body).toContain('checked');
	});

	it('RNF-01: radio com rótulo', () => {
		const { body } = render(Check, {
			props: {
				type: 'radio',
				name: 'faixa',
				value: 'any',
				label: 'Qualquer horário',
				checked: false,
				onchange: noop
			}
		});
		expect(body).toContain('type="radio"');
		expect(body).toContain('Qualquer horário');
	});
});

describe('Select', () => {
	it('RNF-01: tem rótulo acessível, placeholder e opções', () => {
		const { body } = render(Select, {
			props: {
				label: 'Instância',
				placeholder: 'Todas',
				value: '',
				options: [{ value: 'Torre sem fim', label: 'Torre sem fim' }],
				onchange: noop
			}
		});
		expect(body).toContain('aria-label="Instância"');
		expect(body).toContain('Todas');
		expect(body).toContain('Torre sem fim');
	});
});

describe('Drawer', () => {
	it('RNF-01: fechada, fica fora da árvore de acessibilidade e do foco', () => {
		const { body } = render(Drawer, {
			props: { open: false, title: 'Filtros', onclose: noop, children: text('x') }
		});
		expect(body).toContain('aria-hidden="true"');
		expect(body).toContain('inert');
	});

	it('RNF-01: aberta, é um diálogo com título', () => {
		const { body } = render(Drawer, {
			props: { open: true, title: 'Filtros', onclose: noop, children: text('x') }
		});
		expect(body).toContain('role="dialog"');
		expect(body).toContain('aria-label="Filtros"');
		expect(body).not.toContain('inert');
	});
});

describe('recursos locais', () => {
	const root = fileURLToPath(new URL('../../..', import.meta.url));
	const walk = (dir: string): string[] =>
		readdirSync(dir).flatMap((name) => {
			const path = join(dir, name);
			return statSync(path).isDirectory() ? walk(path) : [path];
		});

	it('RNF-02: o design system e os assets não apontam para nenhum domínio externo', () => {
		const files = [...walk(join(root, 'src/lib/ui')), ...walk(join(root, 'static/brand'))].filter(
			(f) => !f.endsWith('.spec.ts')
		);
		const external = files.flatMap((file) =>
			[...readFileSync(file, 'utf8').matchAll(/https?:\/\/[^\s"')]+/g)]
				.map((m) => m[0])
				.filter((url) => !/^https?:\/\/(www\.w3\.org|c2pa\.org)\//.test(url))
				.map((url) => `${file}: ${url}`)
		);
		expect(external).toEqual([]);
	});
});
