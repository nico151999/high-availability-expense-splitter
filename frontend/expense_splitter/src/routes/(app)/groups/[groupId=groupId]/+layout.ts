import type { RouteParams, LayoutLoad } from "./$types";

export const load = (({ params }: {params: RouteParams}) => {
    return {
        groupId: params.groupId
    }
}) satisfies LayoutLoad;