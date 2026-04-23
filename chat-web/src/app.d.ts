// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces

import type { CurrentUser } from "$lib/store/user";

declare global {
	namespace App {
		// interface Error {}
		// interface Locals {}
		interface Locals {
			currentUser: CurrentUser | null;
			session?: Session;
			user?: User;			
		}		
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
