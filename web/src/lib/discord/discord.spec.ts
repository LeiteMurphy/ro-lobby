import { describe, expect, it } from 'vitest';
import { connectMessage, cooldownMessage } from './messages';

describe('mensagens da sala no Discord (spec discord-bot)', () => {
	it('CA-01.4 / CA-01.5: cada resultado da conexão tem a mensagem', () => {
		expect(connectMessage('conta')).toEqual({
			text: 'Use a mesma conta do Discord do seu login.',
			ok: false
		});
		expect(connectMessage('cancelado')).toEqual({ text: 'Conexão cancelada.', ok: false });
		expect(connectMessage('conectado')?.ok).toBe(true);
		expect(connectMessage('outra-coisa')).toBeNull();
		expect(connectMessage(null)).toBeNull();
	});

	it('CA-08.2 / RN-27: "Aguarde 6 min para avisar de novo"', () => {
		expect(cooldownMessage(360)).toBe('Aguarde 6 min para avisar de novo');
		expect(cooldownMessage(301)).toBe('Aguarde 6 min para avisar de novo');
		expect(cooldownMessage(5)).toBe('Aguarde 1 min para avisar de novo');
	});
});
