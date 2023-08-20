import type { RouteParams, LayoutLoad } from "./$types";

export const load = (({ params }: {params: RouteParams}) => {
    return {
        categoryId: params.categoryId
    }
}) satisfies LayoutLoad;