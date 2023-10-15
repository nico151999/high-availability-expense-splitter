<script lang="ts">
	import { createPromiseClient } from '@bufbuild/connect';
	import type { PageData } from './$types';
	import { GroupService } from '../../../../../../gen/lib/ts/service/group/v1/service_connect';
	import { CurrencyService } from '../../../../../../gen/lib/ts/service/currency/v1/service_connect';
	import type { Group } from '../../../../../../gen/lib/ts/common/group/v1/group_pb';
	import type { Currency } from '../../../../../../gen/lib/ts/common/currency/v1/currency_pb';
	import { onDestroy, onMount } from 'svelte';
	import { writable, type Writable } from 'svelte/store';
	import { goto } from '$app/navigation';
	import { streamCurrencies, streamGroups } from './utils';
	import Button, { Label } from '@smui/button';
	import DataTable, { Head, Body, Row, Cell } from '@smui/data-table';
	import LinearProgress from '@smui/linear-progress';
	import LayoutGrid, { Cell as LayoutCell } from '@smui/layout-grid';
	import Textfield from '@smui/textfield';
	import HelperText from '@smui/textfield/helper-text';
	import Select, { Option } from '@smui/select';
	import { t } from '$lib/localization';
	import IconButton from '@smui/icon-button';

	export let data: PageData;

	const groupClient = createPromiseClient(GroupService, data.grpcWebTransport);
	let groups: Writable<
		Map<string, { group?: Group; abortController: AbortController }> | undefined
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
				await groupClient.deleteGroup({ id: groupId });
				console.log(`Deleted group ${groupId}`);
			} catch (e) {
				console.error(`An error occurred trying to delete group ${groupId}`, e);
			}
		};
	}

	function openGroup(groupId: string) {
		return () => {
			goto(`./groups/${groupId}`);
		};
	}
</script>

<LayoutGrid>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<h1>{$t('groups.yourGroups')}</h1>
		
		<DataTable table$aria-label="User list" style="width: 100%">
			<Head>
				<Row>
					<Cell>ID</Cell>
					<Cell>Name</Cell>
					<Cell>Default Currency</Cell>
					<Cell>{$t('groups.action')}</Cell>
				</Row>
			</Head>
			<Body>
				{#if $groups}
					{#each [...$groups] as [gID, group]}
						{#if group.group}
							<Row on:click={openGroup(gID)}>
								<Cell>{gID}</Cell>
								<Cell>{group.group.name}</Cell>
								<Cell>
									{#if $currencies}
										{@const currency = $currencies.get(group.group.currencyId)}
										<span>{currency?.acronym} - {currency?.name}</span>
									{:else}
										<span>{$t('groups.loadingCurrencies')}</span>
									{/if}
								</Cell>
								<Cell>
									<IconButton
										on:click$stopPropagation={deleteGroup(gID)}
										class="material-icons"
										aria-label="Delete group">delete</IconButton>
								</Cell>
							</Row>
						{:else}
							<Row>{$t('groups.loadingGroupWithId', { groupId: gID })}</Row>
						{/if}
					{/each}
				{/if}
			</Body>
		
			<LinearProgress
				indeterminate
				closed={!!$groups}
				aria-label="Groups are being loaded..."
				slot="progress"
			/>
		</DataTable>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<h2>New group</h2>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 6, tablet: 4, phone: 4 }}>
		<Textfield variant="outlined" bind:value={newGroup.name} label="Group name" style="width: 100%" helperLine$style="width: 100%">
			<HelperText slot="helper">The name of the group that is to be created</HelperText>
		</Textfield>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 6, tablet: 4, phone: 4 }}>
		{#if $currencies}
			<Select variant="outlined" bind:value={newGroup.currencyId} label="Currency" style="width: 100%">
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
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }} style="display: flex; justify-content: flex-end">
		<Button on:click={createGroup} touch variant="outlined">
			<Label>Create group</Label>
		</Button>
	</LayoutCell>
</LayoutGrid>