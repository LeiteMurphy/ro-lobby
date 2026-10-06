import type { LayoutServerLoad } from './$types';

// O usuário da sessão vem do hook (spec login-discord, D-03).
export const load: LayoutServerLoad = ({ locals }) => ({ user: locals.user });
