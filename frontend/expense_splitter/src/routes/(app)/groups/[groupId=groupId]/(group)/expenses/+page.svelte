<script lang="ts">
	import { createPromiseClient } from "@bufbuild/connect";
	import type { PageData } from "./$types";
	import { ExpenseService } from "../../../../../../../../../gen/lib/ts/service/expense/v1/service_connect";
	import type { Expense } from "../../../../../../../../../gen/lib/ts/common/expense/v1/expense_pb";
	import { onDestroy, onMount } from "svelte";
	import { get, writable, type Unsubscriber, type Writable } from "svelte/store";
	import { goto } from "$app/navigation";
	import { stakeSumInCurrency, streamExpenseStakes, streamExpenses, summariseStakes } from "./utils";
	import { Timestamp } from "@bufbuild/protobuf";
	import { DateInput } from 'date-picker-svelte';
	import type { Person } from "../../../../../../../../../gen/lib/ts/common/person/v1/person_pb";
	import { PersonService } from "../../../../../../../../../gen/lib/ts/service/person/v1/service_connect";
	import { streamPeople } from "../people/utils";
	import { CurrencyService } from "../../../../../../../../../gen/lib/ts/service/currency/v1/service_connect";
	import type { Currency } from "../../../../../../../../../gen/lib/ts/common/currency/v1/currency_pb";
	import { streamCurrencies, streamGroup } from "../../../utils";
	import type { ExpenseStake } from "../../../../../../../../../gen/lib/ts/common/expensestake/v1/expensestake_pb";
	import { ExpenseStakeService } from "../../../../../../../../../gen/lib/ts/service/expensestake/v1/service_connect";
	import { GroupService } from "../../../../../../../../../gen/lib/ts/service/group/v1/service_connect";
	import type { Group } from "../../../../../../../../../gen/lib/ts/common/group/v1/group_pb";

	export let data: PageData;

	const expensestakeClient = createPromiseClient(ExpenseStakeService, data.grpcWebTransport);
	let expensestakesPerExpense: Map<
		string,
		{
			expensestakes: Writable<
				Map<string, {expensestake?: ExpenseStake, abortController: AbortController}> | undefined
			>,
			abortController: AbortController,
			unsubscriber: Unsubscriber
		}
	> = new Map();

	const expenseClient = createPromiseClient(ExpenseService, data.grpcWebTransport);
	let expenses: Writable<
		Map<string, {expense?: Expense, abortController: AbortController}> | undefined
	> = writable();
	const abortController = new AbortController();
	let unsubscribeExpenses: Unsubscriber|undefined;

	const groupClient = createPromiseClient(GroupService, data.grpcWebTransport);
	let group: Writable<Group | undefined> = writable();
	const groupAbortController = new AbortController();
	let unsubscribeGroup: Unsubscriber|undefined;

	const personClient = createPromiseClient(PersonService, data.grpcWebTransport);
	let people: Writable<
		Map<string, {person?: Person, abortController: AbortController}> | undefined
	> = writable();
	const peopleAbortController = new AbortController();

	const currencyClient = createPromiseClient(CurrencyService, data.grpcWebTransport);
	let currencies: Writable<Map<string, Currency> | undefined> = writable();
	const currencyAbortController = new AbortController();

	let exchangeRatesPerExpense: Map<string, number> = new Map();

	let expensesSummary: number|undefined;

	let timestampValid: boolean | undefined;
	const newExpense = {
		name: '',
		byId: '',
		currencyId: '',
		timestamp: new Date()
	};

	onDestroy(() => {
		abortController.abort();
		peopleAbortController.abort();
		currencyAbortController.abort();
		groupAbortController.abort();
		if (unsubscribeExpenses) { unsubscribeExpenses() };
		if (unsubscribeGroup) { unsubscribeGroup() };
		for (let [_, expensestakes] of expensestakesPerExpense) {
			expensestakes.abortController.abort();
			expensestakes.unsubscriber();
		}
	});

	onMount(() => {
		streamExpenses(expenseClient, data.groupId, abortController, expenses);
		streamPeople(personClient, data.groupId, abortController, people);
		streamCurrencies(currencyClient, currencyAbortController, currencies);
		streamGroup(groupClient, data.groupId, groupAbortController, group);
		unsubscribeExpenses = expenses.subscribe(onExpensesChanged);
		unsubscribeGroup = group.subscribe(onGroupChanged);
	});

	async function updateExchangeRates() {
		if (!$group || !$expenses) {
			exchangeRatesPerExpense.clear();
			expensesSummary = summariseExpenses();
			return;
		}
		// remove all exchange rates of expenses that are no longer used
		for (let expenseId of exchangeRatesPerExpense.keys()) {
			if (!$expenses.has(expenseId)) {
				exchangeRatesPerExpense.delete(expenseId);
				expensesSummary = summariseExpenses();
			}
		}
		// add exchange rates for new currencies and update existing ones
		for (let [expenseId, expense] of $expenses) {
			if (expense.expense) {
				try {
					const res = await currencyClient.getExchangeRate({
						sourceCurrencyId: expense.expense.currencyId,
						destinationCurrencyId: $group.currencyId,
						timestamp: expense.expense.timestamp
					})
					exchangeRatesPerExpense.set(expenseId, res.rate);
					expensesSummary = summariseExpenses();
				} catch (e) {
					console.error(e);
				}
			}
		}
	}

	async function onGroupChanged() {
		await updateExchangeRates();
	}

	async function onExpensesChanged() {
		// stop and remove all expensestakes subscriptions if no expenses are set
		if (!$expenses) {
			for (let [expenseId, expensestakes] of expensestakesPerExpense) {
				expensestakes.abortController.abort();
				expensestakes.unsubscriber();
				expensestakesPerExpense.delete(expenseId);
			}
			return;
		}
		// stop and remove all expensestakes subscriptions of expenses that are no longer in use
		for (let [expenseId, expensestakes] of expensestakesPerExpense) {
			if (!$expenses.has(expenseId)) {
				expensestakes.abortController.abort();
				expensestakes.unsubscriber();
				expensestakesPerExpense.delete(expenseId);
			}
		}
		// start expensestakes subscriptions for all new expenses
		for (let expenseId of $expenses.keys()) {
			if (!expensestakesPerExpense.has(expenseId)) {
				let expenseStakes: Writable<
					Map<string, {expensestake?: ExpenseStake, abortController: AbortController}> | undefined
				> = writable();
				const unsubscribe = expenseStakes.subscribe(() => {
					expensesSummary = summariseExpenses();
				});
				const abortController = new AbortController();
				streamExpenseStakes(expensestakeClient, expenseId, abortController, expenseStakes);
				expensestakesPerExpense.set(expenseId, {
					expensestakes: expenseStakes,
					abortController: abortController,
					unsubscriber: unsubscribe
				});
			}
		}

		await updateExchangeRates();

		expensesSummary = summariseExpenses();
	}

	async function createExpense() {
		if (!timestampValid) {
			console.error('Cannot create expense if timestamp input is invalid');
			return;
		}
		try {
			const res = await expenseClient.createExpense({
                groupId: data.groupId,
                name: newExpense.name,
				byId: newExpense.byId,
				currencyId: newExpense.currencyId,
				timestamp: Timestamp.fromDate(newExpense.timestamp),
			});
			console.log('Created expense', res.id);

			newExpense.name = '';
			newExpense.byId = '';
			newExpense.currencyId = '';
			newExpense.timestamp = new Date();
		} catch (e) {
			console.error(`An error occurred trying to create expense in group ${data.groupId}`, e);
		}
	}

	function deleteExpense(expenseId: string) {
		return async () => {
			try {
				await expenseClient.deleteExpense({id: expenseId});
				console.log('Deleted expense');
			} catch (e) {
				console.error(`An error occurred trying to delete expense ${expenseId} in group ${data.groupId}`, e);
			}
		}
	}

	function summariseExpenses(): number|undefined {
		let totalStakeSum = 0;
		for (let [expenseId, es] of expensestakesPerExpense) {
			const expensestakes = get(es.expensestakes);
			if (!expensestakes) {
				return;
			}
			const stakes: ExpenseStake[] = [];
			for (let [_, stake] of expensestakes) {
				if (stake.expensestake) {
					stakes.push(stake.expensestake);
				} else {
					return;
				}
			}
			const stakeSum = summariseStakes(stakes);
			const exchangeRate = exchangeRatesPerExpense.get(expenseId);
			if (!exchangeRate) {
				return;
			}
			totalStakeSum += stakeSumInCurrency(exchangeRate, stakeSum);
		}
		return totalStakeSum;
	}

	function openExpense(expenseId: string) {
		return () => {
			goto(`./expenses/${expenseId}`);
		}
	}
</script>


{#if $group}
	<h2>Your expenses in group {$group?.name}</h2>
	{#if $currencies && expensesSummary}
		<span>Total value: {expensesSummary.toFixed(2)} {$currencies.get($group.currencyId)?.acronym}</span>
	{:else}
		<span>Loading total value...</span>
	{/if}
{:else}
	<span>Loading group...</span>
{/if}

<table>
	<thead>
		<th>ID</th>
		<th>Name</th>
		<th>By</th>
		<th>Currency</th>
		<th>Timestamp</th>
		<th>Action</th>
	</thead>
	<tbody>
		{#if $expenses && $people}
			{#each [...$expenses] as [pID, expense]}
				{#if expense.expense}
					{@const e = expense.expense}
					<tr on:click={openExpense(pID)}>
						<td>{pID}</td>
						<td>{e.name}</td>
						<td>{$people.get(e.byId)?.person?.name}</td>
						<td>
							{#if $currencies}
								{@const currency = $currencies.get(e.currencyId)}
								<span>{currency?.acronym} - {currency?.name}</span>
							{:else}
								<span>Loading currencies...</span>
							{/if}
						</td>
						<td>{e.timestamp?.toDate().toLocaleString()}</td>
						<td><button on:click|stopPropagation={deleteExpense(pID)}>Delete</button></td>
					</tr>
				{:else}
					<tr>Loading expense with ID {pID}...</tr>
				{/if}
			{/each}
			<tr>
				<td></td>
				<td><input type="text" placeholder="Expense name" bind:value={newExpense.name}/></td>
				<td>
					<select bind:value={newExpense.byId}>
						{#each [...$people] as [pID, person]}
							<option value={pID}>{person.person?.name}</option>
						{/each}
					</select>
				</td>
				<td>
					{#if $currencies}
						<select bind:value={newExpense.currencyId}>
							{#each [...$currencies] as [cID, currency]}
								<option value={cID}>{currency.acronym} - {currency.name}</option>
							{/each}
						</select>
					{:else}
						<span>Loading currencies...</span>
					{/if}
				</td>
				<td><DateInput min={new Date(1640995200000)} max={new Date()} bind:value={newExpense.timestamp} bind:valid={timestampValid}/></td>
				<td><button on:click={createExpense}>Create expense</button></td>
			</tr>
		{:else}
			<tr>Loading expenses...</tr>
		{/if}
	</tbody>
</table>