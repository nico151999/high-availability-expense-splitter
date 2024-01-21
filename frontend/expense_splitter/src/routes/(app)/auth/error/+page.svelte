<script lang="ts">
	import { onMount } from "svelte";
	import type { PageData } from "./$types";
	import type { FlowError } from "@ory/kratos-client";
    import { page } from '$app/stores';
	import { getKratosApi } from "$lib/kratos";

	export let data: PageData;
    let flowError: FlowError|undefined;

    onMount(async () => {
        const errorId = $page.url.searchParams.get('id');
        if (errorId) {
            flowError = (await getKratosApi(data.kratosUrl).getFlowError({id: errorId})).data;
        }
    });

    Object.entries(flowError?.error ?? {})
</script>
{#if flowError}
    <p>The flow error is as follows:</p>
    <table>
        <thead>
            <tr>
                <th>Key</th>
                <th>Value</th>
            </tr>
        </thead>
        <tbody>
            {#each Object.entries(flowError?.error ?? {}) as errEntry}
                <tr>
                    <td>{errEntry[0]}</td>
                    <td>{errEntry[1]}</td>
                </tr>
            {/each}
        </tbody>
    </table>
{:else}
    <p>There is no flow error detail (yet)</p>
{/if}