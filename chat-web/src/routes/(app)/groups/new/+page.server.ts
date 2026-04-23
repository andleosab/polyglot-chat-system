import { fail, redirect, error } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { API } from '$lib/server/config';
import type { CreateUserResponse } from '$lib/api/types/user';
import type { Participant } from '$lib/api/types/conversation';

export const load: PageServerLoad = async ({ locals, fetch }) => {
  const res = await fetch(`${API.users}/api/users`);
  const users: CreateUserResponse[] = await res.json();

  // exclude current user — they're auto-added by backend
  return {
    currentUser: locals.currentUser,
    users: users.filter(u => u.userId !== locals.currentUser?.userId)
  };
};

export const actions: Actions = {
  create: async ({ request, locals, fetch }) => {

    const currentUser = locals.currentUser!;

    const formData = await request.formData();
    const name = formData.get('name') as string;
    const participantsJson = formData.get('participants') as string;

    if (!name?.trim()) {
      return fail(400, { error: 'Group name is required' });
    }

    // participants come as JSON string from the form hidden input, parse it back to array
    let participants: Participant[] = [];
    try {
      participants = participantsJson ? JSON.parse(participantsJson) : [];
    } catch {
      return fail(400, { error: 'Invalid participants data' });
    }

    try {
      // Use existing API now. Provide composite that includes group and participants endpoint later. 
      // 1. Create group
      const createRes = await fetch(`${API.messages}/conversations`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          type: 'group', 
          name,
          created_by: currentUser.userId
        }),
      });
      if (!createRes.ok) throw new Error('Failed to create group');      

      const group = await createRes.json();
      const groupId = group.conversation_id;

      // 2. Add current user as admin
      await fetch(`${API.messages}/groups/${groupId}/participants`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          user_uuid: currentUser.userId,
          username: currentUser.username,
          is_admin: true,
        }),
      });      
      
      // 3. Add remaining participants
      await Promise.all(
        participants.map(p =>
          fetch(`${API.messages}/groups/${groupId}/participants`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              user_uuid: p.user_uuid,
              username: p.username,
              is_admin: p.isAdmin ?? false,
            }),
          })
        )
      );

    } catch (err) {
      if (err instanceof Response) throw err; // let redirects through
      return fail(500, { error: 'Failed to create group' });
    }

    redirect(303, '/groups'); 

  }
};