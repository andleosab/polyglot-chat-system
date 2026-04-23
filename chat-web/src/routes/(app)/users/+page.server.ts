import type { PageServerLoad } from './$types';
import { fail, error, redirect } from '@sveltejs/kit';
// import { USER_API_BASE } from '$env/static/private';
import { API } from '$lib/server/config';
import type { CreateUserResponse } from '$lib/api/types/user';


async function fetchUsers(path: string, fetch: typeof globalThis.fetch) {
    const url = new URL(`${API.users}${path}`);

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

    console.log('Loading user page with base url:', API.users);

	try {

		let data: CreateUserResponse[] = await fetchUsers('/api/users', fetch);

		// check if data is empty. it can happen when a last user deleted from a current page.
		// in this case we should fetch the first page
		if (data.length === 0) {
			console.warn('No users found for the given page and size. Fetching default page.');
		}

		console.log('Users found:', data);

		return {
			 users: data.filter(u => u.userId !== locals.currentUser?.userId)
		};

	} catch (err: any) {
		console.error('Error fetching users:', err);

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

        throw error(500, 'Failed to load users');        

	}    

}