<script lang="ts">
	import { createPromiseClient } from "@bufbuild/connect";
	import { onDestroy, onMount } from "svelte";
	import { writable, type Unsubscriber, type Writable } from "svelte/store";
	import type { Expense } from "../../../../../../../../../../../gen/lib/ts/common/expense/v1/expense_pb";
	import { ExpenseService } from "../../../../../../../../../../../gen/lib/ts/service/expense/v1/service_connect";
	import { streamExpense } from "../../utils";
	import type { PageData } from "./$types";
	import { Timestamp } from "@bufbuild/protobuf";
	import { DateInput, DatePicker } from 'date-picker-svelte';
	import type { Person } from "../../../../../../../../../../../gen/lib/ts/common/person/v1/person_pb";
	import { PersonService } from "../../../../../../../../../../../gen/lib/ts/service/person/v1/service_connect";
	import { streamPeople } from "../../../people/utils";
	import Expensestakes from "./expensestakes.svelte";
	import { CurrencyService } from "../../../../../../../../../../../gen/lib/ts/service/currency/v1/service_connect";
	import type { Currency } from "../../../../../../../../../../../gen/lib/ts/common/currency/v1/currency_pb";
	import { stakeSumInCurrency, streamCurrencies, streamGroup } from "../../../../../utils";
	import { GroupService } from "../../../../../../../../../../../gen/lib/ts/service/group/v1/service_connect";
	import type { Group } from "../../../../../../../../../../../gen/lib/ts/common/group/v1/group_pb";
	import Categories from "./categories.svelte";
	import LayoutGrid, {Cell as LayoutCell} from "@smui/layout-grid";
	import Textfield from "@smui/textfield";
	import Select, {Option} from "@smui/select";
	import LinearProgress from "@smui/linear-progress";
	import Button, { Label } from "@smui/button";
	import { Separator } from "@smui/list";
	import { t } from "$lib/localization";

	export let data: PageData;

	const expenseClient = createPromiseClient(ExpenseService, data.grpcWebTransport);
	let expense: Writable<Expense | undefined> = writable();
	const abortController = new AbortController();
	let unsubscribeExpense: Unsubscriber|undefined;

	const personClient = createPromiseClient(PersonService, data.grpcWebTransport);
	let people: Writable<
		Map<string, {person?: Person, abortController: AbortController}> | undefined
	> = writable();
	const peopleAbortController = new AbortController();

	const currencyClient = createPromiseClient(CurrencyService, data.grpcWebTransport);
	let currencies: Writable<Map<string, Currency> | undefined> = writable();
	const currenciesAbortController = new AbortController();

	const groupClient = createPromiseClient(GroupService, data.grpcWebTransport);
	let group: Writable<Group | undefined> = writable();
	const groupAbortController = new AbortController();
	const unsubscribeGroup = group.subscribe(updateExchangeRate);

    let editMode = false;
	const editedExpense = {
		name: '',
		byId: '',
		currencyId: '',
		timestamp: new Date()
	};
	$: if (!editMode) {
		editedExpense.name = $expense?.name ?? '';
		editedExpense.byId = $expense?.byId ?? '';
		editedExpense.currencyId = $expense?.currencyId ?? '';
		editedExpense.timestamp = $expense?.timestamp?.toDate() ?? new Date();
	}

	let stakeSum: string;
	let exchangeRate: number | undefined;

	onDestroy(() => {
		if (unsubscribeExpense) { unsubscribeExpense() };
		unsubscribeGroup();
		abortController.abort();
		peopleAbortController.abort();
		currenciesAbortController.abort();
		groupAbortController.abort();
	});

	onMount(async () => {
		streamPeople(personClient, data.groupId, abortController, people);
		streamCurrencies(currencyClient, currenciesAbortController, currencies);
		streamGroup(groupClient, data.groupId, groupAbortController, group);
		unsubscribeExpense = expense.subscribe(updateExchangeRate);
		const res = await streamExpense(expenseClient, data.expenseId, abortController, expense);
        if (!res) {
            console.error('Expense no longer exists');
        }
	});

	async function updateExchangeRate() {
		if ($expense && $group) {
			try {
				const res = await currencyClient.getExchangeRate({
					sourceCurrencyId: $expense.currencyId,
					destinationCurrencyId: $group.currencyId,
					timestamp: $expense.timestamp
				})
				exchangeRate = res.rate;
			} catch (e) {
				console.error(e);
			}
		}
	}

	async function updateExpense() {
		try {
			const res = await expenseClient.updateExpense({
				id: data.expenseId,
				updateFields: [
					{
						updateOption: {
							case: 'name',
							value: editedExpense.name
						}
					},
					{
						updateOption: {
							case: 'byId',
							value: editedExpense.byId
						}
					},
					{
						updateOption: {
							case: 'currencyId',
							value: editedExpense.currencyId
						}
					},
					{
						updateOption: {
							case: 'timestamp',
							value: Timestamp.fromDate(editedExpense.timestamp)
						}
					}
				]
			});
			console.log('Updated expense', res.expense);
            editMode = false;
		} catch (e) {
			console.error('An error occurred trying to update expense', e);
		}
	}

    function startEdit() {
        if (!$expense) {
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
		<h2>Expense</h2>
	</LayoutCell>
	{#if $expense}
		<LayoutCell spanDevices={{ desktop: 6, tablet: 4, phone: 4 }}>
			<Textfield variant="outlined" disabled={!editMode} bind:value={editedExpense.name} label="Expense name" style="width: 100%" />
		</LayoutCell>
		<LayoutCell spanDevices={{ desktop: 6, tablet: 4, phone: 4 }}>
			{#if $people}
				<Select variant="outlined" disabled={!editMode} bind:value={editedExpense.byId} label="Person" style="width: 100%">
					{#each [...$people] as [pID, person]}
						{#if person.person}
							<Option value={pID}>
								{person.person.name}
							</Option>
						{/if}
					{/each}
				</Select>
			{/if}
			<LinearProgress
				indeterminate
				closed={!!$currencies}
				aria-label="Currencies are being loaded..."/>
		</LayoutCell>
		<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
			{#if $currencies}
				<Select variant="outlined" disabled={!editMode} bind:value={editedExpense.currencyId} label="Currency" style="width: 100%">
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
			<DatePicker min={new Date(1640995200000)} max={new Date()} bind:value={editedExpense.timestamp}/>
		</LayoutCell>
		<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }} style="display: flex; justify-content: center">
			{#if editMode}
				<Button on:click={updateExpense} variant="outlined">
					<Label>Update expense</Label>
				</Button>
				<Button on:click={stopEdit} variant="outlined">
					<Label>Cancel</Label>
				</Button>
			{:else}
				<Button on:click={startEdit} variant="outlined">
					<Label>Edit expense</Label>
				</Button>
			{/if}
		</LayoutCell>
	{/if}
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<LinearProgress
			indeterminate
			closed={!!$expense}
			aria-label="Expense is being loaded..."/>
	</LayoutCell>
	{#if $expense}
		<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
			<Separator />
		</LayoutCell>
		<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
			<Expensestakes expense={$expense} transport={data.grpcWebTransport} people={people} bind:stakeSum={stakeSum}></Expensestakes>
			{#if $currencies && $group && exchangeRate}
				{@const exCurrency = $currencies.get($expense.currencyId)}
				{@const grCurrency = $currencies.get($group.currencyId)}
				<span>{stakeSum} {exCurrency?.acronym} - {stakeSumInCurrency(exchangeRate, stakeSum).toFixed(2)} {grCurrency?.acronym}</span>
			{:else}
				<span>Loading data...</span>
			{/if}
		</LayoutCell>
		<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
			<Separator />
		</LayoutCell>
		<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
			<Categories expense={$expense} transport={data.grpcWebTransport}></Categories>
		</LayoutCell>
	{/if}
</LayoutGrid>