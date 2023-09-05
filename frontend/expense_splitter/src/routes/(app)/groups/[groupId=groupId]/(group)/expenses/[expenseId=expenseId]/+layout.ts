import type { RouteParams, LayoutLoad } from "./$types";

export const load = (({ params }: {params: RouteParams}) => {
    return {
        expenseId: params.expenseId
    }
}) satisfies LayoutLoad;