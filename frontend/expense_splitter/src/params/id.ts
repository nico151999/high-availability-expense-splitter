import type { ParamMatcher } from '@sveltejs/kit';

export const match = ((param) => {
    return /^[a-z]+-[A-Za-z0-9]{15}$/.test(param);
}) satisfies ParamMatcher;