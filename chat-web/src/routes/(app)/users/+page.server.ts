import type { PageServerLoad } from './$types';
import { error, redirect } from '@sveltejs/kit';
import { API } from '$lib/server/config';
import type { CreateUserResponse } from '$lib/api/types/user';
import type { PresenceInfo } from '$lib/api/types/presence';


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

async function fetchPresence(userId: string, fetch: typeof globalThis.fetch): Promise<PresenceInfo> {
    try {
        const res = await fetch(`${API.presence}/presence/${userId}`);
        if (!res.ok) return { online: false, last_seen: null };
        return await res.json();
    } catch {
        return { online: false, last_seen: null };
    }
}

export const load: PageServerLoad = async ({ params, locals, fetch }) => {

    console.log('Loading user page with base url:', API.users);

	try {

		let data: CreateUserResponse[] = await fetchUsers('/api/users', fetch);

		if (data.length === 0) {
			console.warn('No users found for the given page and size. Fetching default page.');
		}

		const otherUsers = data.filter(u => u.userId !== locals.currentUser?.userId);

		// Enrich with presence — failures are non-fatal (returns offline defaults)
		const presenceResults = await Promise.allSettled(
			otherUsers.map(u => fetchPresence(u.userId, fetch))
		);

		const presenceMap: Record<string, PresenceInfo> = {};
		otherUsers.forEach((u, i) => {
			const r = presenceResults[i];
			presenceMap[u.userId] = r.status === 'fulfilled' ? r.value : { online: false, last_seen: null };
		});

		return { users: otherUsers, presenceMap };

	} catch (err: any) {
		console.error('Error fetching users:', err);

        if (err?.status === 401) {
            throw redirect(303, '/sign-in');
        }

        if (err?.status === 403) {
            return { users: [], presenceMap: {} };
        }

        throw error(500, 'Failed to load users');

	}

}