<script lang="ts">
	import { goto } from "$app/navigation";
	import { createPromiseClient } from "@bufbuild/connect";
	import { onDestroy, onMount } from "svelte";
	import { writable, type Writable } from "svelte/store";
	import type { Person } from "../../../../../../../../../../../gen/lib/ts/common/person/v1/person_pb";
	import { PersonService } from "../../../../../../../../../../../gen/lib/ts/service/person/v1/service_connect";
	import { streamPerson } from "../../utils";
	import type { PageData } from "./$types";

	export let data: PageData;

	const personClient = createPromiseClient(PersonService, data.grpcWebTransport);
	let person = writable(undefined as Person | undefined);
	const abortController = new AbortController();

    let editMode = false;

	const editedPerson = writable({
		name: ''
	});

	onDestroy(() => {
		abortController.abort();
	});

	onMount(async () => {
		const res = await streamPerson(personClient, data.personId, abortController, person);
        if (!res) {
            console.error('Person no longer exists');
        }
	});

	async function updatePerson() {
		try {
			const res = await personClient.updatePerson({
				id: data.personId,
				updateFields: [
					{
						updateOption: {
							case: 'name',
							value: $editedPerson.name
						}
					}
				]
			});
			console.log('Updated person', res.person);
            editMode = false;
		} catch (e) {
			console.error('An error occurred trying to update person', e);
		}
	}

    function startEdit() {
        if (!$person) {
            return;
        }
        editedPerson.set({
            name: $person.name
        })
        editMode = true;
    }

    function stopEdit() {
        editMode = false;
    }
</script>

<h2>Your person with ID {data.personId}</h2>
<table>
	<thead>
		<th>Name</th>
		<th>Action</th>
	</thead>
	<tbody>
		{#if $person}
            {#if editMode}
                <tr>
                    <td><input type="text" placeholder="Person name" bind:value={$editedPerson.name}/></td>
                    <td>
                        <button on:click={updatePerson}>Update person</button>
                        <button on:click={stopEdit}>Cancel</button>
                    </td>
                </tr>
            {:else}
                <tr>
                    <td>{$person.name}</td>
                    <td><button on:click={startEdit}>Update person</button></td>
                </tr>
            {/if}
		{:else}
			<tr>Loading person...</tr>
		{/if}
	</tbody>
</table>