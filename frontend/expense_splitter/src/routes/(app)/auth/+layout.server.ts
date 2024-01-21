import { getKratosUrl } from '$lib/kratos/server';
import type { LayoutServerLoad } from './$types';

export const load = (async () => {
    return {
        kratosUrl: getKratosUrl(),
    };
}) satisfies LayoutServerLoad;