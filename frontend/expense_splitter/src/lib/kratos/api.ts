import { Configuration, FrontendApi } from "@ory/kratos-client";
import axios from 'axios';

export function getKratosApi(kratosUrl: string) {
    return new FrontendApi(
        new Configuration({
            basePath: kratosUrl, // TODO: make configurable with env variable that is passed to PageData
            baseOptions: {
                // Setting this is very important as axios will send the CSRF cookie otherwise
                // which causes problems with Ory Kratos' security detection.
                withCredentials: true,

                // Timeout after 10 seconds.
                timeout: 10000,
            }
        }),
        '',
        // Ensure that we are using the axios client with retry.
        axios
    );
}