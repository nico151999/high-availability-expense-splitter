import { env } from "$env/dynamic/private";

import { getKratosApi as getKratosApiClient } from "../api";

export function getKratosUrl() {
    const kratosSecure = env.KRATOS_SECURE === 'true' ? 'https' : 'http';
    const kratosAddress = (env.KRATOS_HOSTNAME ?? (() => { throw new Error('Kratos hostname not defined') })()) as string;
    const kratosPort = +((env.KRATOS_PORT ?? (() => { throw new Error('Kratos port not defined') })()) as string);
    return `${kratosSecure}://${kratosAddress}:${kratosPort}`;
}

export function getKratosApi() {
    return getKratosApiClient(getKratosUrl());
}