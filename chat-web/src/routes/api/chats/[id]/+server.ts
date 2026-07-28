import { json, error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
// Read at runtime via config.ts, not $env/static/private: a static import is inlined
// at build time, which would silently ignore the MESSAGE_API_BASE set in the Helm
// values and bake in whatever the build environment happened to have.
import { API } from '$lib/server/config';

export const GET: RequestHandler = async ({ params, url }) => {

  const before = url.searchParams.get('before');
  const limit  = url.searchParams.get('limit');

  const qs = new URLSearchParams();
  if (before) qs.set('mid', before);
  if (limit)  qs.set('limit', limit);

  console.log('==> proxying to:', `${API.messages}/conversations/${params.id}/messages/cursor?${qs}`);

  const res = await fetch(
    `${API.messages}/conversations/${params.id}/messages/cursor?${qs}`
  );

  if (!res.ok) {
    error(res.status, `Backend error: ${res.statusText}`);
  }

  const data = await res.json();
  return json(data);
};
