import { describe, expect, it } from 'vitest';
import { SERVERS, serverName } from './servers';

describe('catálogo de servidores', () => {
	it('CA-01.1 / RN-01: Freya e Nidhogg, nesta ordem, com os códigos da API', () => {
		expect(SERVERS).toEqual([
			{ id: 'freya', name: 'Freya' },
			{ id: 'nidhogg', name: 'Nidhogg' }
		]);
	});

	it('RN-05: nome do servidor para a tela', () => {
		expect(serverName('nidhogg')).toBe('Nidhogg');
		expect(serverName('thor')).toBe('thor');
	});
});
