import { fail, redirect, error } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { API } from '$lib/server/config';
import type { CreateUserResponse } from '$lib/api/types/user';
import type { ConversationResponse, Participant } from '$lib/api/types/conversation';
import { currentUser } from '$lib/store/user';

export const load: PageServerLoad = async ({ params, locals, fetch }) => {
  const id = Number(params.id);
  if (!Number.isInteger(id) || id <= 0) throw error(404, 'Group not found');

  const [usersRes, participantsRes, conversationRes] = await Promise.all([
    fetch(`${API.users}/api/users`),
    fetch(`${API.messages}/groups/${id}/participants/`),
    fetch(`${API.messages}/conversations/${id}`),
  ]);

  if (!usersRes.ok) {
    const errBody = await usersRes.json().catch(() => null);
    throw error(usersRes.status, errBody?.error || 'Failed to load users');
  }
  if (!participantsRes.ok) {
    const errBody = await participantsRes.json().catch(() => null);
    throw error(participantsRes.status, errBody?.error || 'Failed to load participants');
  }
  if (!conversationRes.ok) {
    const errBody = await conversationRes.json().catch(() => null);
    throw error(conversationRes.status, errBody?.error || 'Failed to load conversation');
        
  }
  if (!conversationRes.ok) {
    const errBody = await conversationRes.json().catch(() => null);
    throw error(conversationRes.status, errBody?.error || 'Failed to load conversation');
  }

  const users: CreateUserResponse[] = await usersRes.json();
  const participantsApi: Participant[] = await participantsRes.json();
  const conversation: ConversationResponse = await conversationRes.json();

  // work around mismatch between BE is_admin and UI isAdmin
  const participants: Participant[] = participantsApi.map((p: any) => ({
    user_uuid: p.user_uuid,
    username: p.username,
    isAdmin: p.is_admin
  }));

  // console.log('Participants:', participants);

  console.log('Participants with admin flag:', participants.map(p => ({
    ...p,
    isAdmin: p.isAdmin ?? false
  })));

  return {
    currentUser: locals.currentUser,
    createdBy: conversation.created_by,
    groupId: id,
    groupName: decodeURIComponent(params.name),
    users: users.filter(u => u.userId !== locals.currentUser?.userId),
    participants,
  };
};

export const actions: Actions = {
  update: async ({ request, params, locals, fetch }) => {
    if (!locals.currentUser) throw error(401, 'Unauthorized');
    const id = Number(params.id);
    const formData = await request.formData();
    const name = formData.get('name') as string;
    const selectedJson = formData.get('participants') as string;
    const originalJson = formData.get('originalParticipants') as string;

    if (!name?.trim()) {
      return fail(400, { error: 'Group name is required' });
    }

    let selected: Participant[] = [];
    let original: Participant[] = [];

    try {
      selected = selectedJson ? JSON.parse(selectedJson) : [];
      original = originalJson ? JSON.parse(originalJson) : [];
    } catch {
      return fail(400, { error: 'Invalid participants data' });
    }

    // Diff — who was removed, who was added
    const removed = original.filter(
      o => !selected.some(s => s.user_uuid === o.user_uuid)
    );
    const added = selected.filter(
      s => !original.some(o => o.user_uuid === s.user_uuid)
    );

    try {
      // 1. Rename group
      console.log(`Renaming group ${id} to "${name}"...`);

      const oldName = decodeURIComponent(params.name);
      console.log(`Old name: "${oldName}", New name: "${name}"`); 

      if (oldName !== name) {
        console.log(`Group name has changed. Sending rename request...`);
        const renameRes = await fetch(`${API.messages}/conversations/${id}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name }),
        });
        if (!renameRes.ok) {
          const errBody = await renameRes.json().catch(() => null);
          throw error(renameRes.status, errBody?.error || 'Failed to rename group');
        }
      } else {
        console.log(`Group name has not changed. Skipping rename request.`);
      }

      // 2. Remove participants
      await Promise.all(
        removed.map(p =>
          fetch(`${API.messages}/groups/${id}/participants/${p.user_uuid}`, {
            method: 'DELETE',
          })
        )
      );

      // 3. Add new participants
      await Promise.all(
        added.map(p =>
          fetch(`${API.messages}/groups/${id}/participants`, {
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
      console.error('Failed to update group:', err);
      if (err instanceof Response) {
        const errJson = await err.json().catch(() => null);
        return fail(err.status, { error: `Failed to update group: ${err.status} ${errJson?.error ?? ''}` });
      }
      return fail(500, { error: 'Failed to update group' });
    }

    redirect(303, '/groups');

  }
};