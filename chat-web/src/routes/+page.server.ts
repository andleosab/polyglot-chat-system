import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ locals, request }) => {
    // console.log('==> in root +page.server.ts, No user session found, redirecting to sign-in page');
    // throw redirect(303, '/chats');

};