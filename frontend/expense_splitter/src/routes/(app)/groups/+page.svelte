<script lang="ts">
	import { Code, ConnectError, createPromiseClient } from "@bufbuild/connect";
	import type { PageData } from "./$types";
	import { GroupService } from "../../../../../../gen/lib/ts/service/group/v1/service_connectweb";
	import type { Group } from "../../../../../../gen/lib/ts/common/group/v1/group_pb";
	import { onDestroy, onMount } from "svelte";
	import { writable } from "svelte/store";
	import { goto } from "$app/navigation";
	import { streamGroup } from "./utils";

	export let data: PageData;

	const groupClient = createPromiseClient(GroupService, data.grpcWebTransport);
	let groups: Map<bigint, {group?: Group, abortController: AbortController}> | undefined;

	const newGroup = writable({
		name: '',
		currency: ''
	});

	const abortController = new AbortController();
	const currenciesAbortController = new AbortController();
	onDestroy(() => {
		abortController.abort();
		currenciesAbortController.abort();
		if (groups) {
			for (const [_, group] of groups) {
				group.abortController.abort();
			}
		}
	});

	onMount(() => {
		streamGroups();
	});

	async function createGroup() {
		try {
			const createRequest = async () => {
				groupClient.createGroup
				const group = await groupClient.createGroup({
					name: $newGroup.name
				});
				console.log('Created group', group);
			};

			const creationRequests = [];
			creationRequests.push(createRequest());
			await Promise.all(creationRequests);

			newGroup.set({name: '', currency: ''});
		} catch (e) {
			console.error('An error occurred trying to create group', e);
		}
	}

	async function streamGroups() {
		try {
			for await (const gIDsRes of groupClient.streamGroupIds({}, {signal: abortController.signal})) {
				if (gIDsRes.update.case === 'stillAlive') {
					continue;
				}
				const groupIDs = gIDsRes.update.value!.groupIds;
				if (groups === undefined) {
					groups = new Map();
				}
				for (const gID of groups.keys()) {
					if (groupIDs.includes(gID)) {
						// remove element from items that are to be processed because it already exists
						groupIDs.splice(groupIDs.indexOf(gID), 1);
					} else {
						// remove groups that are not present any more
						groups!.get(gID)!.abortController.abort();
						groups!.delete(gID);
					}
				}
				for (const gID of groupIDs) {
					const abortController = new AbortController();
					groups.set(gID, {
						abortController: abortController
					});
					streamGroup(groupClient, gID, abortController, (group) => {
						groups = groups?.set(gID, {
							abortController: abortController,
							group: group
						});
					});
				}
				groups = groups;
			}
		} catch (e) {
            if (e instanceof ConnectError && e.code === Code.Canceled) {
				console.log('Intentionally aborting groups stream');
				return;
            } else {
				console.error('An error occurred trying to stream group IDs. Trying anew in 5 seconds.', e);
			}
		}
		console.log('Ended groups stream');
		setTimeout(() => streamGroups(), 5000);
	}

	function deleteGroup(groupId: bigint) {
		return async () => {
			try {
				const groupRes = await groupClient.deleteGroup({groupId: groupId});
				console.log('Deleted group', groupRes.group);
			} catch (e) {
				console.error(`An error occurred trying to delete group ${groupId}`, e);
			}
		}
	}

	function openGroup(groupId: bigint) {
		return () => {
			goto(`/groups/${groupId}`);
		}
	}
</script>

<h2>Your groups</h2>

<table>
	<thead>
		<th>ID</th>
		<th>Name</th>
		<th>Currency</th>
		<th>Action</th>
	</thead>
	<tbody>
		{#if groups}
			{#each [...groups] as [gID, group]}
				{#if group.group}
					<tr on:click={openGroup(gID)}>
						<td>{gID}</td>
						<td>{group.group.name}</td>
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
			<td><button on:click={createGroup}>Create group</button></td>
		</tr>
	</tbody>
</table>