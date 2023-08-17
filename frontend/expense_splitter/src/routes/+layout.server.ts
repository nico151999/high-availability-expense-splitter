import type { LayoutServerLoad } from './$types';

export const load = (() => {
    return {
        schema: process.env.API_SECURE === 'true' ? 'https' : 'http',
        address: (process.env.API_HOSTNAME ?? (() => {throw new Error('hostname not defined')})()) as string,
        port: +((process.env.API_PORT ?? (() => {throw new Error('port not defined')})()) as string)
    };
}) satisfies LayoutServerLoad;