import { addTranslations, setLocale, setRoute } from '$lib/localization';
import type { LayoutLoad } from './$types';
import { createGrpcWebTransport } from '@bufbuild/connect-web';

export const load = (async ({ data }) => {
    const { i18n, translations, apiSchema, apiAddress, apiPort } = data;
    const { locale, route } = i18n;

    addTranslations(translations);

    await setRoute(route);
    await setLocale(locale);

    return {
        grpcWebTransport: createGrpcWebTransport({baseUrl: `${apiSchema}://${apiAddress}:${apiPort}`}),
        ...i18n
    }
}) satisfies LayoutLoad;