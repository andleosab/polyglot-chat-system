import type { PageServerLoad } from './$types';
import { fail, error, redirect } from '@sveltejs/kit';
// import { USER_API_BASE } from '$env/static/private';
import { API } from '$lib/server/config';

async function fetchUserConversations(uuid: string | undefined, fetch: typeof globalThis.fetch) {

    const url = new URL(`${API.messages}/conversations?id=${uuid}`);

    console.log('Constructed URL:', url.toString());

    // // Add query parameters
    // Object.entries(queryParams).forEach(([key, value]) => {
    //     url.searchParams.append(key, String(value));
    // });

    console.log('fetching from:', url.toString());

    const response = await fetch(url.toString(), {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
    });

    console.log('==> ' + response.status);
    console.log('==> Time:', new Date().toLocaleTimeString());
    console.log('==> URL: ', url.toString());

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

export const load: PageServerLoad = async ({ locals, fetch }) => {

    const uuid = locals.currentUser?.userId;

    console.log('User uuid:', uuid);

    console.log('Loading chats page with base url:', API.messages);

	try {

		let data = await fetchUserConversations(uuid, fetch);

		// check if data is empty. it can happen when a last user deleted from a current page.
		// in this case we should fetch the first page
		if (data.length === 0) {
			console.warn('No conversations found for the given user.');
		}

		console.log('Conversations found:', data);

		return {
			conversations: data || []
		};

	} catch (err: any) {

		console.error('*** Error fetching conversations:', err);

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

        throw error(500, 'Failed to load conversations');
	}    

}