// TODO: remove before commit
process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';
import type { Handle } from '@sveltejs/kit';
import { getKratosApi } from '$lib/kratos/server';

export const handle: Handle = async ({ event, resolve }) => {
    if (event.url.pathname !== '/' && !event.url.pathname.startsWith('/auth/')) {
        console.log('Received request to non root path. Checking if user is authenticated...');

        try {
            const session = (await getKratosApi().toSession({
                cookie: event.request.headers.get('cookie') ?? '',
            })).data;
            console.log('User has an active session', session.identity.traits);
        } catch (e) {
            console.log('User is unauthenticated');
            return Response.redirect(`${event.url.origin}/auth/login`);
        }
    }

    return await resolve(event);
};