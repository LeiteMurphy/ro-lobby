// Sala no Discord em construção (spec discord-em-construcao, RN-01): a chave fica no
// ambiente do web. Com "true", a entrada no servidor e a sala do grupo ficam fechadas no
// site; a API e o bot não mudam.
import { env } from '$env/dynamic/private';

export function discordUnderConstruction(): boolean {
	return env.DISCORD_UNDER_CONSTRUCTION === 'true';
}
