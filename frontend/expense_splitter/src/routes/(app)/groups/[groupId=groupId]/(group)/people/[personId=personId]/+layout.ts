import type { RouteParams, LayoutLoad } from "./$types";

export const load = (({ params }: {params: RouteParams}) => {
    return {
        personId: params.personId
    }
}) satisfies LayoutLoad;