<script lang="ts">
	import { createPromiseClient } from "@bufbuild/connect";
	import { onDestroy, onMount } from "svelte";
	import { writable, type Writable } from "svelte/store";
	import type { Person } from "../../../../../../../../../../../gen/lib/ts/common/person/v1/person_pb";
	import { PersonService } from "../../../../../../../../../../../gen/lib/ts/service/person/v1/service_connect";
	import { streamPerson } from "../../utils";
	import type { PageData } from "./$types";
	import { t } from '$lib/localization';
	import LayoutGrid, {Cell as LayoutCell} from "@smui/layout-grid";
	import Textfield from "@smui/textfield";
	import Button, { Label } from "@smui/button";
	import LinearProgress from "@smui/linear-progress";

	export let data: PageData;

	const personClient = createPromiseClient(PersonService, data.grpcWebTransport);
	let person: Writable<Person | undefined> = writable();
	const abortController = new AbortController();

    let editMode = false;

	const editedPerson = {
		name: ''
	};
	$: if (!editMode) {
        editedPerson.name = $person?.name ?? '';
	}

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
							value: editedPerson.name
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
        editMode = true;
    }

    function stopEdit() {
        editMode = false;
    }
</script>

<LayoutGrid>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<h2>Person</h2>
	</LayoutCell>
	{#if $person}
		<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
			<Textfield variant="outlined" disabled={!editMode} bind:value={editedPerson.name} label={$t('person.namePlaceholder')} style="width: 100%" />
		</LayoutCell>
		<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }} style="display: flex; justify-content: center">
			{#if editMode}
				<Button on:click={updatePerson} variant="outlined">
					<Label>Update person</Label>
				</Button>
				<Button on:click={stopEdit} variant="outlined">
					<Label>Cancel</Label>
				</Button>
			{:else}
				<Button on:click={startEdit} variant="outlined">
					<Label>Edit person</Label>
				</Button>
			{/if}
		</LayoutCell>
	{/if}
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<LinearProgress
			indeterminate
			closed={!!$person}
			aria-label="Person is being loaded..."/>
	</LayoutCell>
</LayoutGrid>