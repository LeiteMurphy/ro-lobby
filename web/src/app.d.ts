// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
import type { SessionUser } from '$lib/auth/api';

declare global {
	namespace App {
		// interface Error {}
		interface Locals {
			/** Usuário da sessão, ou null para visitante (spec login-discord). */
			user: SessionUser | null;
		}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
