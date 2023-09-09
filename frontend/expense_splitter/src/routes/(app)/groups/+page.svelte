<script lang="ts">
	import { createPromiseClient } from "@bufbuild/connect";
	import type { PageData } from "./$types";
	import { GroupService } from "../../../../../../gen/lib/ts/service/group/v1/service_connect";
	import type { Group } from "../../../../../../gen/lib/ts/common/group/v1/group_pb";
	import { onDestroy, onMount } from "svelte";
	import { writable, type Writable } from "svelte/store";
	import { goto } from "$app/navigation";
	import { streamGroups } from "./utils";

	export let data: PageData;

	const groupClient = createPromiseClient(GroupService, data.grpcWebTransport);
	let groups: Writable<
		Map<string, {group?: Group, abortController: AbortController}> | undefined
	> = writable();

	const newGroup = writable({
		name: '',
		currencyId: '' // TODO: add to UI once a currency service is available
	});

	const abortController = new AbortController();
	onDestroy(() => {
		abortController.abort();
	});

	onMount(
		() => streamGroups(groupClient, abortController, groups)
	);

	async function createGroup() {
		try {
			const res = await groupClient.createGroup({
				name: $newGroup.name,
				currencyId: $newGroup.currencyId
			});
			console.log('Created group', res.id);

			newGroup.set({name: '', currencyId: ''});
		} catch (e) {
			console.error('An error occurred trying to create group', e);
		}
	}

	function deleteGroup(groupId: string) {
		return async () => {
			try {
				await groupClient.deleteGroup({id: groupId});
				console.log(`Deleted group ${groupId}`);
			} catch (e) {
				console.error(`An error occurred trying to delete group ${groupId}`, e);
			}
		}
	}

	function openGroup(groupId: string) {
		return () => {
			goto(`./groups/${groupId}`);
		}
	}
</script>

<h2>Your groups</h2>

<table>
	<thead>
		<th>ID</th>
		<th>Name</th>
		<th>Default Currency</th>
		<th>Action</th>
	</thead>
	<tbody>
		{#if $groups}
			{#each [...$groups] as [gID, group]}
				{#if group.group}
					<tr on:click={openGroup(gID)}>
						<td>{gID}</td>
						<td>{group.group.name}</td>
						<td>{group.group.currencyId}</td>
						<td><button on:click|stopPropagation={deleteGroup(gID)}>Delete</button></td>
					</tr>
				{:else}
					<tr>Loading group with ID {gID}...</tr>
				{/if}
			{/each}
		{:else}
			<tr>Loading groups...</tr>
		{/if}
		<tr>
			<td></td>
			<td><input type="text" placeholder="Group name" bind:value={$newGroup.name}/></td>
			<td><input type="text" placeholder="Group name" bind:value={$newGroup.currencyId}/></td> <!-- TODO: make selector -->
			<td><button on:click={createGroup}>Create group</button></td>
		</tr>
	</tbody>
</table>