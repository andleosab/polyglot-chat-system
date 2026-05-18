import { error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { API } from '$lib/server/config';

const forward: RequestHandler = async ({ params, request, fetch }) => {
  const upstream = `${API.presence}/${params.path}`;

  const init: RequestInit = { method: request.method, headers: request.headers };
  if (request.method !== 'GET' && request.method !== 'HEAD') {
    init.body = await request.text();
  }

  const res = await fetch(upstream, init);
  if (!res.ok && res.status >= 500) {
    error(res.status, `Presence service error: ${res.statusText}`);
  }

  return new Response(res.body, {
    status: res.status,
    headers: { 'Content-Type': res.headers.get('Content-Type') ?? 'application/json' },
  });
};

export const GET = forward;
export const POST = forward;
