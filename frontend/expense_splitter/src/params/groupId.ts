import type { ParamMatcher } from '@sveltejs/kit';

export const match = ((param) => {
    return /^group-[A-Za-z0-9]{15}$/.test(param);
}) satisfies ParamMatcher;