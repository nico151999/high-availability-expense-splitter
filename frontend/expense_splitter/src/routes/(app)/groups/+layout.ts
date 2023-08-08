import { createGrpcWebTransport } from "@bufbuild/connect-web";
import type { LayoutLoad } from "./$types";

export const load = (async ({parent}) => {
    const p = await parent();
    return {
        grpcWebTransport: createGrpcWebTransport({baseUrl: `${p.schema}://${p.address}:${p.port}`})
    }
}) satisfies LayoutLoad;