<script lang="ts">
	import { createPromiseClient } from "@bufbuild/connect";
	import { onDestroy, onMount } from "svelte";
	import { writable, type Unsubscriber, type Writable } from "svelte/store";
	import type { Expense } from "../../../../../../../../../../../gen/lib/ts/common/expense/v1/expense_pb";
	import { ExpenseService } from "../../../../../../../../../../../gen/lib/ts/service/expense/v1/service_connect";
	import { streamExpense } from "../../utils";
	import type { PageData } from "./$types";
	import { Timestamp } from "@bufbuild/protobuf";
	import { DateInput } from 'date-picker-svelte';
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
	let timestampValid: boolean | undefined;
	const editedExpense = {
		name: '',
		byId: '',
		currencyId: '',
		timestamp: new Date()
	};

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
		if (!timestampValid) {
			console.error('Cannot update expense if timestamp input is invalid');
			return;
		}
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
		editedExpense.name = $expense.name ?? '';
		editedExpense.byId = $expense.byId;
		editedExpense.currencyId = $expense.currencyId;
		editedExpense.timestamp = $expense.timestamp?.toDate() ?? new Date();
        editMode = true;
    }

    function stopEdit() {
        editMode = false;
    }
</script>

<h2>Your expense with ID {data.expenseId}</h2>
<table>
	<thead>
		<th>Name</th>
		<th>By</th>
		<th>Currency</th>
		<th>Timestamp</th>
		<th>Action</th>
	</thead>
	<tbody>
		{#if $expense}
            {#if editMode}
                <tr>
                    <td><input type="text" placeholder="Expense name" bind:value={editedExpense.name}/></td>
					<td>
						{#if $people}
							<select bind:value={editedExpense.byId}>
								{#each [...$people] as [pID, person]}
									<option value={pID}>{person.person?.name}</option>
								{/each}
							</select>
						{:else}
							<span>Loading people...</span>
						{/if}
					</td>
					<td>
						{#if $currencies}
							<select bind:value={editedExpense.currencyId}>
								{#each [...$currencies] as [cID, currency]}
									<option value={cID}>{currency.acronym} - {currency.name}</option>
								{/each}
							</select>
						{:else}
							<span>Loading currencies...</span>
						{/if}
					</td>
					<td><DateInput min={new Date(1640995200000)} max={new Date()} bind:value={editedExpense.timestamp} bind:valid={timestampValid}/></td>
                    <td>
                        <button on:click={updateExpense}>Update expense</button>
                        <button on:click={stopEdit}>Cancel</button>
                    </td>
                </tr>
            {:else}
                <tr>
                    <td>{$expense.name}</td>
					<td>
						{#if $people}
							<span>{$people.get($expense.byId)?.person?.name}</span>
						{:else}
							<span>Loading people...</span>
						{/if}
					</td>
					<td>
						{#if $currencies}
							{@const currency = $currencies.get($expense.currencyId)}
							<span>{currency?.acronym} - {currency?.name}</span>
						{:else}
							<span>Loading currencies...</span>
						{/if}
					</td>
					<td>{$expense.timestamp?.toDate().toLocaleString()}</td>
                    <td><button on:click={startEdit}>Update expense</button></td>
                </tr>
            {/if}
		{:else}
			<tr>Loading expense...</tr>
		{/if}
	</tbody>
</table>
{#if $expense}
	<Expensestakes expense={$expense} transport={data.grpcWebTransport} people={people} bind:stakeSum={stakeSum}></Expensestakes>
	{#if $currencies && $group && exchangeRate}
		{@const exCurrency = $currencies.get($expense.currencyId)}
		{@const grCurrency = $currencies.get($group.currencyId)}
		<span>{stakeSum} {exCurrency?.acronym} - {stakeSumInCurrency(exchangeRate, stakeSum).toFixed(2)} {grCurrency?.acronym}</span>
	{:else}
		<span>Loading data...</span>
	{/if}
	<Categories expense={$expense} transport={data.grpcWebTransport}></Categories>
{/if}