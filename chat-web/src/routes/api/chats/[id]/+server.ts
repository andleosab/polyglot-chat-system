import { json, error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { MESSAGE_API_BASE } from '$env/static/private';

export const GET: RequestHandler = async ({ params, url }) => {

  const before = url.searchParams.get('before');
  const limit  = url.searchParams.get('limit');

  const qs = new URLSearchParams();
  if (before) qs.set('mid', before);
  if (limit)  qs.set('limit', limit);

  console.log('==> proxying to:', `${MESSAGE_API_BASE}/conversations/${params.id}/messages/cursor?${qs}`);

  const res = await fetch(
    `${MESSAGE_API_BASE}/conversations/${params.id}/messages/cursor?${qs}`
  );

  if (!res.ok) {
    error(res.status, `Backend error: ${res.statusText}`);
  }

  const data = await res.json();
  return json(data);
};
