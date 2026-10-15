import { describe, expect, it } from 'vitest';
import { alertRoom } from './api';
import {
	ALERT_MESSAGES,
	connectMessage,
	cooldownMessage,
	OPEN_ROOM_MESSAGES,
	ruleMessage
} from './messages';

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

describe('excesso de requisições na sala (spec endurecimento, RN-15)', () => {
	const call = (res: Response) => ({
		fetchFn: (async () => res) as typeof fetch,
		apiBaseUrl: 'http://api.test',
		token: 't'
	});
	const json429 = (body: unknown) =>
		new Response(JSON.stringify(body), {
			status: 429,
			headers: { 'content-type': 'application/json' }
		});

	it('CA-05.2: o 429 do limite vira rate_limited, com a mensagem da regra', async () => {
		const result = await alertRoom(call(json429({ error: 'rate_limited' })), 'l1');
		expect(result).toEqual({ ok: false, kind: 'rate_limited' });
		expect(ruleMessage(ALERT_MESSAGES, 'rate_limited')).toBe(
			'Muitas tentativas. Aguarde um pouco e tente de novo.'
		);
		expect(ruleMessage(OPEN_ROOM_MESSAGES, 'rate_limited')).toBe(
			'Muitas tentativas. Aguarde um pouco e tente de novo.'
		);
		expect(connectMessage('muitas')?.text).toBe(
			'Muitas tentativas. Aguarde um pouco e tente de novo.'
		);
	});

	it('RN-27 da discord-bot: o intervalo do aviso continua sendo cooldown', async () => {
		const result = await alertRoom(
			call(json429({ error: 'cooldown', retryAfterSeconds: 300 })),
			'l1'
		);
		expect(result).toEqual({ ok: false, kind: 'cooldown', retryAfterSeconds: 300 });
	});
});
