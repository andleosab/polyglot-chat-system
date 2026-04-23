// Svelte5 Rune based state management for user session data
// replace this import when CurentUser type is moved to $lib/api/types/user.ts
import type { CurrentUser } from '$lib/store/user';

class UserSession {
  currentUser = $state<CurrentUser | null>(null);
}

export const userSession = new UserSession();