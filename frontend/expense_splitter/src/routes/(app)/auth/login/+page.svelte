<script lang="ts">
	import { onMount } from "svelte";
	import type { PageData } from "./$types";
	import { UiTextTypeEnum, type LoginFlow, type UiNodeInputAttributes, type UiNode } from "@ory/kratos-client";
    import { page } from '$app/stores';
	import { getKratosApi } from "$lib/kratos";
	import Button, { Label } from "@smui/button";
	import Textfield from "@smui/textfield";
	import LayoutGrid, { Cell as LayoutCell } from '@smui/layout-grid';
	import { goto } from "$app/navigation";
    import axios from 'axios';

	export let data: PageData;
    let loginFlow: LoginFlow|undefined;
    let errorMessage: string|undefined;

    onMount(async () => {
        const flowId = $page.url.searchParams.get('flow');
        const kratosApi = getKratosApi(data.kratosUrl);
        if (flowId) {
            loginFlow = (await kratosApi.getLoginFlow({id: flowId})).data;
        } else {
            try {
                loginFlow = (await kratosApi.createBrowserLoginFlow()).data;
            } catch (e) {
                if (axios.isAxiosError(e)) {
                    errorMessage = e.response?.data.error?.message;
                } else {
                    console.error(e);
                }
            }
        }
        
    });

    function getNodeInputAttributes(node: UiNode): UiNodeInputAttributes {
        if (node.attributes.node_type === 'input') {
            return node.attributes as UiNodeInputAttributes;
        } else {
            throw `Expected node type to be input but it was ${node.attributes.node_type}`;
        }
    }
</script>
<LayoutGrid>
    <LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
        {#if loginFlow}
            {@const ui = loginFlow.ui}
            <form method="{ui.method}" action="{ui.action}">
                {#each ui.nodes.slice(0, ui.nodes.length-1) as node}
                    {@const inputAttributes = getNodeInputAttributes(node)}
                    {#if inputAttributes.type === 'hidden'}
                        <input name="{inputAttributes.name}" type="{inputAttributes.type}" required="{inputAttributes.required}" value="{inputAttributes.value ?? ''}"/>
                    {:else}
                        <Textfield input$name="{inputAttributes.name}" input$autocomplete="{inputAttributes.autocomplete}" input$pattern="{inputAttributes.pattern}" type="{inputAttributes.type}" required="{inputAttributes.required}" variant="outlined" value="{inputAttributes.value ?? ''}" label="{node.meta.label?.text}" style="width: 100%" helperLine$style="width: 100%"/>
                    {/if}
                    {#each node.messages as message}
                        {#if message.type === UiTextTypeEnum.Error}
                            <p style="color: red">{message.text}</p>
                        {:else}
                            <p>{message.text}</p>
                        {/if}
                    {/each}
                {/each}
                <Button name="method" value="password" type="submit" touch variant="outlined">
                    <Label>Log in</Label>
                </Button>
                <Button on:click="{() => goto('/auth/registration')}" touch variant="outlined">
                    <Label>Register instead?</Label>
                </Button>
            </form>
            {#if ui.messages}
                {#each ui.messages as message}
                    {#if message.type === UiTextTypeEnum.Error}
                        <p style="color: red">{message.text}</p>
                    {:else}
                        <p>{message.text}</p>
                    {/if}
                {/each}
            {/if}
        {:else if errorMessage}
            <p style="color: red">{errorMessage}</p>
        {:else}
            <p>There is no login flow (yet)</p>
        {/if}
    </LayoutCell>
</LayoutGrid>