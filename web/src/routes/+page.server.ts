import { env } from '$env/dynamic/private';
import { checkApi } from '$lib/api/health';
import type { PageServerLoad } from './$types';

const DEFAULT_API_BASE_URL = 'http://localhost:8080';

// RN-15 / CA-05.4: a consulta roda no servidor, então o HTML já chega com o estado.
export const load: PageServerLoad = async ({ fetch }) => {
	return { apiStatus: await checkApi(fetch, env.API_BASE_URL || DEFAULT_API_BASE_URL) };
};
