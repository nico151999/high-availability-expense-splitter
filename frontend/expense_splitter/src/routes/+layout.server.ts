import type { LayoutServerLoad } from './$types';
import { API_SECURE, API_HOSTNAME, API_PORT } from '$env/static/private';
import { locales, loadTranslations, translations } from '$lib/localization';

export const load = (async ({ url, cookies, request }) => {
    const { pathname } = url;

    // Try to get the locale from cookie
    let locale = (cookies.get('lang') || '').toLowerCase();

    // Get user preferred locale
    if (!locale) {
        locale = `${`${request.headers.get('accept-language')}`.match(/^[a-zA-Z]+/)}`.toLowerCase();
    }

    // Get defined locales
    const supportedLocales = locales.get().map((l) => l.toLowerCase());

    // Use default locale if current locale is not supported
    if (!supportedLocales.includes(locale)) {
        locale = 'en';
    }

    await loadTranslations(locale, pathname);
    return {
        schema: API_SECURE === 'true' ? 'https' : 'http',
        address: (API_HOSTNAME ?? (() => { throw new Error('hostname not defined') })()) as string,
        port: +((API_PORT ?? (() => { throw new Error('port not defined') })()) as string),
        i18n: { locale, route: pathname },
        translations: translations.get(),
    };
}) satisfies LayoutServerLoad;