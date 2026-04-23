import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';

import { auth } from "$lib/auth"; 
import { svelteKitHandler } from "better-auth/svelte-kit";
import { building } from '$app/environment';
import type { CurrentUser } from '$lib/store/user';
import { issueServiceToken, issueWsToken } from '$lib/server/jwt';

import { USER_API_BASE, MESSAGE_API_BASE } from '$env/static/private';

type ServiceConfig = {
  issueToken: (user: CurrentUser) => string;
};

// map service base URL → token strategy
const SERVICE_MAP: Record<string, ServiceConfig> = {
  [MESSAGE_API_BASE]: {
    issueToken: (user) => issueServiceToken('chat-message-service', user)
  },
  [USER_API_BASE]: {
    issueToken: (user) => issueServiceToken('chat-user-service', user)
  }
};

export const handle: Handle = async ({ event, resolve }) => {

  // console.log('==> (Better Auth) in hooks.server.ts, processing request for:', event.url.toString());

  // -------------------------
  // 1. SESSION (Better Auth)
  // -------------------------  
  // Fetch current session from Better Auth
  const session = await auth.api.getSession({
    headers: event.request.headers,
  });

  event.locals.session = session?.session;
  event.locals.user = session?.user;

  if (session?.user) {
    event.locals.currentUser = {
      userId: session.user.useruuid!,
      email: session.user.email,
      username: session.user.username!
    };
  }

  // -------------------------
  // 2. FETCH WRAPPER (JWT injection)
  // -------------------------
  const originalFetch = event.fetch;

  const tokenCache = new Map<string, string>();

  event.fetch = async (input, init = {}) => {
    const url = typeof input === 'string' ? input : input.toString();

    console.log(`Checking if fetch URL ${url} matches any protected service...`);

    const match = Object.entries(SERVICE_MAP)
      .find(([base]) => url.startsWith(base));

    if (match && event.locals.currentUser) {

      // console.log(`==> *** Intercepted fetch request to: ${url}`);

      const [base, config] = match;

      // per-request cache per service base
      let token = tokenCache.get(base);

      if (!token) {
        token = config.issueToken(event.locals.currentUser);
        tokenCache.set(base, token);
      }

      const headers = new Headers(init.headers ?? {});

      if (init.body && !headers.has('Content-Type')) {
        headers.set('Content-Type', 'application/json');
      }

      console.log(`===> Injecting JWT for service ${base} into request to ${url}`);
      headers.set('Authorization', `Bearer ${token}`);

      init = {
        ...init,
        headers
      };
    }

    return originalFetch(input, init);
  };  

  // -------------------------
  // 3. protect app routes
  // -------------------------
  if (event.route.id?.startsWith('/(app)/') && !session?.user) {
    throw redirect(303, '/sign-in');
  }

  // -------------------------
  // 4. CONTINUE PIPELINE
  // -------------------------  
  return svelteKitHandler({ event, resolve, auth, building });

};