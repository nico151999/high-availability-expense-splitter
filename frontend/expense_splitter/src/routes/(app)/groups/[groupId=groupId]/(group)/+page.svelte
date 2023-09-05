<script lang="ts">
	import { goto } from "$app/navigation";
	import { createPromiseClient } from "@bufbuild/connect";
	import { onDestroy, onMount } from "svelte";
	import { writable } from "svelte/store";
	import type { Group } from "../../../../../../../../gen/lib/ts/common/group/v1/group_pb";
	import { GroupService } from "../../../../../../../../gen/lib/ts/service/group/v1/service_connect";
	import { streamGroup } from "../../utils";
	import type { PageData } from "./$types";

	export let data: PageData;

	const groupClient = createPromiseClient(GroupService, data.grpcWebTransport);
	let group = writable(undefined as Group | undefined);
	const abortController = new AbortController();
    let editMode = false;

	const editedGroup = writable({
		name: '',
		currencyId: ''
	});

	onDestroy(() => {
		abortController.abort();
	});

	onMount(async () => {
		const res = await streamGroup(groupClient, data.groupId, abortController, group);
        if (!res) {
            console.error('Group no longer exists');
        }
	});

	async function updateGroup() {
		try {
			const res = await groupClient.updateGroup({
				id: data.groupId,
				updateFields: [
					{
						updateOption: {
							case: 'name',
							value: $editedGroup.name
						}
					},
					{
						updateOption: {
							case: 'currencyId',
							value: $editedGroup.currencyId
						}
					}
				]
			});
			console.log('Updated group', res.group);
            editMode = false;
		} catch (e) {
			console.error('An error occurred trying to update group', e);
		}
	}

	function openExpenses() {
		goto(`./${data.groupId}/expenses`);
	}

	function openCategories() {
		goto(`./${data.groupId}/categories`);
	}

	function openPeople() {
		goto(`./${data.groupId}/people`);
	}

    function startEdit() {
        if (!$group) {
            return;
        }
        editedGroup.set({
            name: $group.name,
			currencyId: $group.currencyId
        })
        editMode = true;
    }

    function stopEdit() {
        editMode = false;
    }
</script>

<h2>Your group with ID {data.groupId}</h2>
<table>
	<thead>
		<th>Name</th>
		<th>Currency</th>
		<th>Action</th>
	</thead>
	<tbody>
		{#if $group}
            {#if editMode}
                <tr>
                    <td><input type="text" placeholder="Group name" bind:value={$editedGroup.name}/></td>
                    <td><input type="text" placeholder="Currency" bind:value={$editedGroup.currencyId}/></td>
                    <td>
                        <button on:click={updateGroup}>Update group</button>
                        <button on:click={stopEdit}>Cancel</button>
                    </td>
                </tr>
            {:else}
                <tr>
                    <td>{$group.name}</td>
                    <td>{$group.currencyId}</td>
                    <td><button on:click={startEdit}>Update group</button></td>
                </tr>
            {/if}
		{:else}
			<tr>Loading group...</tr>
		{/if}
	</tbody>
</table>
<div>
	<button on:click={openExpenses}>Open expenses</button>
	<button on:click={openCategories}>Open categories</button>
	<button on:click={openPeople}>Open people</button>
</div>