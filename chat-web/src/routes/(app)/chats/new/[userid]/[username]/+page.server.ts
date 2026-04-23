import { API } from '$lib/server/config';
import type { PageServerLoad } from './$types';
import { goto } from '$app/navigation';
import { redirect, error } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ locals, params, fetch }) => {

  const sourceUserUuid = locals.currentUser?.userId;
  if (!sourceUserUuid) {
    throw error(401, 'Unauthorized');
  }
  const targetUserUuid = params.userid;
  const targetUsername = decodeURIComponent(params.username);

  const url = new URL(`${API.messages}/conversations/private?user1=${sourceUserUuid}&user2=${targetUserUuid}`);
  const response = await fetch(url.toString(), {
    method: 'GET',
    headers: { 'Content-Type': 'application/json' },
  });

  if (response.ok) { // Conversation already exists, navigate to it
    const conv = await response.json();
    throw redirect(303, `/chats/${conv.conversation_id}/${encodeURIComponent(targetUsername ?? '')}`);
  }

  return { // No conversation exists, return target user info for new chat */
    targetUserUuid: targetUserUuid,
    targetUsername: targetUsername,
  };
};
