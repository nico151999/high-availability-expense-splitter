<script lang="ts">
	import { createPromiseClient } from "@bufbuild/connect";
	import type { PageData } from "./$types";
	import { GroupService } from "../../../../../../gen/lib/ts/service/group/v1/service_connect";
	import { CurrencyService } from "../../../../../../gen/lib/ts/service/currency/v1/service_connect";
	import type { Group } from "../../../../../../gen/lib/ts/common/group/v1/group_pb";
	import type { Currency } from "../../../../../../gen/lib/ts/common/currency/v1/currency_pb";
	import { onDestroy, onMount } from "svelte";
	import { writable, type Writable } from "svelte/store";
	import { goto } from "$app/navigation";
	import { streamCurrencies, streamGroups } from "./utils";
	import { t } from '$lib/localization';

	export let data: PageData;

	const groupClient = createPromiseClient(GroupService, data.grpcWebTransport);
	let groups: Writable<
		Map<string, {group?: Group, abortController: AbortController}> | undefined
	> = writable();
	const abortController = new AbortController();

	const currencyClient = createPromiseClient(CurrencyService, data.grpcWebTransport);
	let currencies: Writable<Map<string, Currency> | undefined> = writable();
	const currencyAbortController = new AbortController();

	const newGroup = {
		name: '',
		currencyId: ''
	};

	onDestroy(() => {
		abortController.abort();
		currencyAbortController.abort();
	});

	onMount(() => {
		streamGroups(groupClient, abortController, groups);
		streamCurrencies(currencyClient, currencyAbortController, currencies);
	});

	async function createGroup() {
		try {
			const res = await groupClient.createGroup({
				name: newGroup.name,
				currencyId: newGroup.currencyId
			});
			console.log('Created group', res.id);

			newGroup.name = '';
			newGroup.currencyId = '';
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

<h1>{$t('groups.yourGroups')}</h1>

<table>
	<thead>
		<th>ID</th>
		<th>Name</th>
		<th>Default Currency</th>
		<th>{$t('groups.action')}</th>
	</thead>
	<tbody>
		{#if $groups}
			{#each [...$groups] as [gID, group]}
				{#if group.group}
					<tr on:click={openGroup(gID)}>
						<td>{gID}</td>
						<td>{group.group.name}</td>
						<td>
							{#if $currencies}
								{@const currency = $currencies.get(group.group.currencyId)}
								<span>{currency?.acronym} - {currency?.name}</span>
							{:else}
								<span>{$t('groups.loadingCurrencies')}</span>
							{/if}
						</td>
						<td><button on:click|stopPropagation={deleteGroup(gID)}>{$t('groups.delete')}</button></td>
					</tr>
				{:else}
					<tr>{$t('groups.loadingGroupWithId', {groupId: gID})}</tr>
				{/if}
			{/each}
		{:else}
			<tr>Loading groups...</tr>
		{/if}
		<tr>
			<td></td>
			<td><input type="text" placeholder="Group name" bind:value={newGroup.name}/></td>
			<td>
				{#if $currencies}
					<select bind:value={newGroup.currencyId}>
						{#each [...$currencies] as [cID, currency]}
							<option value={cID}>{currency.acronym} - {currency.name}</option>
						{/each}
					</select>
				{:else}
					<span>Loading currencies...</span>
				{/if}
			</td>
			<td><button on:click={createGroup}>Create group</button></td>
		</tr>
	</tbody>
</table>