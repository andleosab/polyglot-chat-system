import type { PageServerLoad } from './$types';
import { fail, error, redirect } from '@sveltejs/kit';
import { API } from '$lib/server/config';
import { LIMIT } from '$lib/store/messages';

async function fetchConversationMessages(id: number, userid: string, fetch: typeof globalThis.fetch): Promise<any[]> {

    // initial load - fetch the first page of messages (before=MAX_INT64) with a limit of 20
    const url = new URL(`${API.messages}/conversations/${id}/messages/cursor?limit=${LIMIT}&userid=${userid}`);

    console.log('fetching from:', url.toString());

    const response = await fetch(url.toString(), {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
    });

    console.log('==> ' + response.status);
    console.log('Time:', new Date().toLocaleTimeString());
    console.log('URL: ', url.toString());

    if (!response.ok) {
        const errBody = await response.json().catch(() => null);
        throw {
            status: response.status,
            body: errBody
        };      
    }  

    const data = await response.json();
    return data;
}

export const load: PageServerLoad = async ({ params, locals, fetch }) => {

    const id = Number(params.id);
    const userid = locals.currentUser!.userId;

    // 2. Handle invalid numbers with SvelteKit's error helper
    if (!Number.isInteger(id) || id <= 0) {
        throw error(404, {
            message: 'No conversation found with the provided ID',
        });
    }

    console.log('Conversation ID:', id);


	try {

		let data = await fetchConversationMessages(id, userid, fetch);

		// check if data is empty. it can happen when a last user deleted from a current page.
		// in this case we should fetch the first page
		if (data.length === 0) {
			console.warn('No messages found.');
		}

		console.log('Messages found:', data);

		return {
            conversationId: id,
			messages: data || []
		};

	} catch (err: any) {
		console.error('Error fetching messages:', err);


        // handle expected auth failure properly
        // REVISIT: show error message on the page instead of redirecting 
        // to sign-in page if the user is authenticated but token is expired or invalid for some reason.
        if (err?.status === 401) {
            // console.log(err.body?.error);
            throw redirect(303, '/sign-in');
        }

        // optional: handle forbidden gracefully instead of 500
        // shwo error message on the page instead of showing empty list if 
        // the user is authenticated but doesn't have access to conversations for some reason.
        if (err?.status === 403) {
            //  console.log(err.body?.error);
            return {
                conversations: []
            };
        }

        throw error(500, 'Failed to load messages');
	}    

}