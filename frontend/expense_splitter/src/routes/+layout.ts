import { addTranslations, setLocale, setRoute } from '$lib/localization';
import type { Load } from '@sveltejs/kit';

export const load: Load = async ({ data }) => {
    const { i18n, translations } = data;
    const { locale, route } = i18n;

    addTranslations(translations);

    await setRoute(route);
    await setLocale(locale);

    return i18n;
};