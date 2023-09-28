import i18n, {type Config, type Parser} from 'sveltekit-i18n';
import lang from './lang.json';

const config: Config<Parser.Params> = ({
    translations: {
        en: {lang},
        de: {lang},
        nb: {lang}
    },
    loaders: [
        {
            locale: 'en',
            key: 'root',
            routes: ['/'],
            loader: async () => (
                await import('./en/root.json')
            ).default,
        },
        {
            locale: 'en',
            key: 'groups',
            routes: ['/groups'],
            loader: async () => (
                await import('./en/groups.json')
            ).default,
        },
        {
            locale: 'en',
            key: 'group',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}$/],
            loader: async () => (
                await import('./en/group.json')
            ).default,
        },
        {
            locale: 'en',
            key: 'categories',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}\/categories$/],
            loader: async () => (
                await import('./en/category.json')
            ).default,
        },
        {
            locale: 'en',
            key: 'category',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}\/categories\/category-[A-Za-z0-9]{15}$/],
            loader: async () => (
                await import('./en/category.json')
            ).default,
        },
        {
            locale: 'en',
            key: 'expenses',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}\/expenses$/],
            loader: async () => (
                await import('./en/expenses.json')
            ).default,
        },
        {
            locale: 'en',
            key: 'expenses',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}\/expenses\/expense-[A-Za-z0-9]{15}$/],
            loader: async () => (
                await import('./en/expense.json')
            ).default,
        },
        {
            locale: 'en',
            key: 'people',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}\/people$/],
            loader: async () => (
                await import('./en/people.json')
            ).default,
        },
        {
            locale: 'en',
            key: 'person',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}\/people\/person-[A-Za-z0-9]{15}$/],
            loader: async () => (
                await import('./en/person.json')
            ).default,
        },
        {
            locale: 'de',
            key: 'root',
            routes: ['/'],
            loader: async () => (
                await import('./de/root.json')
            ).default,
        },
        {
            locale: 'de',
            key: 'groups',
            routes: ['/groups'],
            loader: async () => (
                await import('./de/groups.json')
            ).default,
        },
        {
            locale: 'de',
            key: 'group',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}$/],
            loader: async () => (
                await import('./de/group.json')
            ).default,
        },
        {
            locale: 'de',
            key: 'categories',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}\/categories$/],
            loader: async () => (
                await import('./de/category.json')
            ).default,
        },
        {
            locale: 'de',
            key: 'category',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}\/categories\/category-[A-Za-z0-9]{15}$/],
            loader: async () => (
                await import('./de/category.json')
            ).default,
        },
        {
            locale: 'de',
            key: 'expenses',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}\/expenses$/],
            loader: async () => (
                await import('./de/expenses.json')
            ).default,
        },
        {
            locale: 'de',
            key: 'expenses',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}\/expenses\/expense-[A-Za-z0-9]{15}$/],
            loader: async () => (
                await import('./de/expense.json')
            ).default,
        },
        {
            locale: 'de',
            key: 'people',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}\/people$/],
            loader: async () => (
                await import('./de/people.json')
            ).default,
        },
        {
            locale: 'de',
            key: 'person',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}\/people\/person-[A-Za-z0-9]{15}$/],
            loader: async () => (
                await import('./de/person.json')
            ).default,
        },
        {
            locale: 'nb',
            key: 'root',
            routes: ['/'],
            loader: async () => (
                await import('./nb/root.json')
            ).default,
        },
        {
            locale: 'nb',
            key: 'groups',
            routes: ['/groups'],
            loader: async () => (
                await import('./nb/groups.json')
            ).default,
        },
        {
            locale: 'nb',
            key: 'group',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}$/],
            loader: async () => (
                await import('./nb/group.json')
            ).default,
        },
        {
            locale: 'nb',
            key: 'categories',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}\/categories$/],
            loader: async () => (
                await import('./nb/category.json')
            ).default,
        },
        {
            locale: 'nb',
            key: 'category',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}\/categories\/category-[A-Za-z0-9]{15}$/],
            loader: async () => (
                await import('./nb/category.json')
            ).default,
        },
        {
            locale: 'nb',
            key: 'expenses',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}\/expenses$/],
            loader: async () => (
                await import('./nb/expenses.json')
            ).default,
        },
        {
            locale: 'nb',
            key: 'expenses',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}\/expenses\/expense-[A-Za-z0-9]{15}$/],
            loader: async () => (
                await import('./nb/expense.json')
            ).default,
        },
        {
            locale: 'nb',
            key: 'people',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}\/people$/],
            loader: async () => (
                await import('./nb/people.json')
            ).default,
        },
        {
            locale: 'nb',
            key: 'person',
            routes: [/^\/groups\/group-[A-Za-z0-9]{15}\/people\/person-[A-Za-z0-9]{15}$/],
            loader: async () => (
                await import('./nb/person.json')
            ).default,
        },
    ],
});

export const { t, loading, locales, locale, loadTranslations, addTranslations, translations, setLocale, setRoute } = new i18n(config);

loading.subscribe((l) => l && console.log('Loading translations for the main instance...'));