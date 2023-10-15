<script lang="ts">
	import { goto } from "$app/navigation";
	import { createPromiseClient } from "@bufbuild/connect";
	import { onDestroy, onMount } from "svelte";
	import { writable, type Writable } from "svelte/store";
	import type { Group } from "../../../../../../../../gen/lib/ts/common/group/v1/group_pb";
	import { GroupService } from "../../../../../../../../gen/lib/ts/service/group/v1/service_connect";
	import { streamCurrencies, streamGroup } from "../../utils";
	import type { PageData } from "./$types";
	import { CurrencyService } from "../../../../../../../../gen/lib/ts/service/currency/v1/service_connect";
	import type { Currency } from "../../../../../../../../gen/lib/ts/common/currency/v1/currency_pb";
	import LayoutGrid, {Cell as LayoutCell} from "@smui/layout-grid";
	import { t } from "$lib/localization";
	import LinearProgress from "@smui/linear-progress";
	import Button, { Label } from "@smui/button";
	import Textfield from "@smui/textfield";
	import Select, {Option} from "@smui/select";

	export let data: PageData;

	const groupClient = createPromiseClient(GroupService, data.grpcWebTransport);
	let group: Writable<Group | undefined> = writable();
	const abortController = new AbortController();

	const currencyClient = createPromiseClient(CurrencyService, data.grpcWebTransport);
	let currencies: Writable<Map<string, Currency> | undefined> = writable();
	const currencyAbortController = new AbortController();

    let editMode = false;

	const editedGroup = {
		name: '',
		currencyId: ''
	};
	$: if (!editMode) {
        editedGroup.name = $group?.name ?? '';
		editedGroup.currencyId = $group?.currencyId ?? '';
	}

	onDestroy(() => {
		abortController.abort();
		currencyAbortController.abort();
	});

	onMount(async () => {
		streamCurrencies(currencyClient, currencyAbortController, currencies);
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
							value: editedGroup.name
						}
					},
					{
						updateOption: {
							case: 'currencyId',
							value: editedGroup.currencyId
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
        editMode = true;
    }

    function stopEdit() {
        editMode = false;
    }
</script>

<LayoutGrid>
	{#if $group}
		<LayoutCell spanDevices={{ desktop: 6, tablet: 4, phone: 4 }}>
			<Textfield variant="outlined" disabled={!editMode} bind:value={editedGroup.name} label="Group name" style="width: 100%" />
		</LayoutCell>
		<LayoutCell spanDevices={{ desktop: 6, tablet: 4, phone: 4 }}>
			{#if $currencies}
				<Select variant="outlined" disabled={!editMode} bind:value={editedGroup.currencyId} label="Currency" style="width: 100%">
					{#each [...$currencies] as [cID, currency]}
						<Option value={cID}>{currency.acronym} - {currency.name}</Option>
					{/each}
				</Select>
			{/if}
			<LinearProgress
				indeterminate
				closed={!!$currencies}
				aria-label="Currencies are being loaded..."/>
		</LayoutCell>
		<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }} style="display: flex; justify-content: center">
			{#if editMode}
				<Button on:click={updateGroup} variant="outlined">
					<Label>Update group</Label>
				</Button>
				<Button on:click={stopEdit} variant="outlined">
					<Label>Cancel</Label>
				</Button>
			{:else}
				<Button on:click={startEdit} variant="outlined">
					<Label>Edit group</Label>
				</Button>
			{/if}
		</LayoutCell>
	{/if}
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<LinearProgress
			indeterminate
			closed={!!$group}
			aria-label="Group is being loaded..."/>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 4, tablet: 4, phone: 2 }} style="display: flex; justify-content: center">
		<Button on:click={openExpenses} variant="outlined">
			<Label>Open expenses</Label>
		</Button>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 4, tablet: 4, phone: 2 }} style="display: flex; justify-content: center">
		<Button on:click={openCategories} variant="outlined">
			<Label>Open categories</Label>
		</Button>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 4, tablet: 8, phone: 4 }} style="display: flex; justify-content: center">
		<Button on:click={openPeople} variant="outlined">
			<Label>Open people</Label>
		</Button>
	</LayoutCell>
</LayoutGrid>