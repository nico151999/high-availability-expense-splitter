import type { ParamMatcher } from '@sveltejs/kit';

export const match = ((param) => {
    return /^category-[A-Za-z0-9]{15}$/.test(param);
}) satisfies ParamMatcher;