<script lang="ts">
	import { createPromiseClient } from "@bufbuild/connect";
	import type { PageData } from "./$types";
	import { PersonService } from "../../../../../../../../../gen/lib/ts/service/person/v1/service_connect";
	import type { Person } from "../../../../../../../../../gen/lib/ts/common/person/v1/person_pb";
	import { onDestroy, onMount } from "svelte";
	import { writable, type Writable } from "svelte/store";
	import { goto } from "$app/navigation";
	import { streamPeople } from "./utils";

	export let data: PageData;

	const personClient = createPromiseClient(PersonService, data.grpcWebTransport);
	let people: Writable<
		Map<string, {person?: Person, abortController: AbortController}> | undefined
	> = writable();

	const newPerson = writable({
		name: '',
	});

	const abortController = new AbortController();
	onDestroy(() => {
		abortController.abort();
		if ($people) {
			for (const [_, person] of $people) {
				person.abortController.abort();
			}
		}
	});

	onMount(
		() => streamPeople(personClient, data.groupId, abortController, people)
	);

	async function createPerson() {
		try {
			const res = await personClient.createPerson({
                groupId: data.groupId,
                name: $newPerson.name
			});
			console.log('Created person', res.personId);

            newPerson.set({name: ''});
		} catch (e) {
			console.error(`An error occurred trying to create person in group ${data.groupId}`, e);
		}
	}

	function deletePerson(personId: string) {
		return async () => {
			try {
				await personClient.deletePerson({personId: personId});
				console.log('Deleted person');
			} catch (e) {
				console.error(`An error occurred trying to delete person ${personId} in group ${data.groupId}`, e);
			}
		}
	}

	function openPerson(personId: string) {
		return () => {
			goto(`./people/${personId}`);
		}
	}
</script>

<h2>Your people in group {data.groupId}</h2>

<table>
	<thead>
		<th>ID</th>
		<th>Name</th>
		<th>Action</th>
	</thead>
	<tbody>
		{#if $people}
			{#each [...$people] as [pID, person]}
				{#if person.person}
					<tr on:click={openPerson(pID)}>
						<td>{pID}</td>
						<td>{person.person?.name}</td>
						<td><button on:click|stopPropagation={deletePerson(pID)}>Delete</button></td>
					</tr>
				{:else}
					<tr>Loading person with ID {pID}...</tr>
				{/if}
			{/each}
		{:else}
			<tr>Loading people...</tr>
		{/if}
		<tr>
			<td></td>
			<td><input type="text" placeholder="Person name" bind:value={$newPerson.name}/></td>
			<td><button on:click={createPerson}>Create person</button></td>
		</tr>
	</tbody>
</table>