// src/lib/store/user.ts
import { writable } from 'svelte/store';
export const currentUserUuid = writable<string | null>(null);

export interface CurrentUser {
  userId: string;
  username: string;
  email: string;
}

export const currentUser = writable<CurrentUser | null>(null);
