<script lang="ts">
	import { createPromiseClient } from "@bufbuild/connect";
	import type { PageData } from "./$types";
	import { ExpenseService } from "../../../../../../../../../gen/lib/ts/service/expense/v1/service_connect";
	import type { Expense } from "../../../../../../../../../gen/lib/ts/common/expense/v1/expense_pb";
	import { onDestroy, onMount } from "svelte";
	import { get, writable, type Unsubscriber, type Writable } from "svelte/store";
	import { goto } from "$app/navigation";
	import { streamExpenseStakesInExpense, streamExpenses } from "./utils";
	import { Timestamp } from "@bufbuild/protobuf";
	import { DateInput, DatePicker } from 'date-picker-svelte';
	import type { Person } from "../../../../../../../../../gen/lib/ts/common/person/v1/person_pb";
	import { PersonService } from "../../../../../../../../../gen/lib/ts/service/person/v1/service_connect";
	import { streamPeople } from "../people/utils";
	import { CurrencyService } from "../../../../../../../../../gen/lib/ts/service/currency/v1/service_connect";
	import type { Currency } from "../../../../../../../../../gen/lib/ts/common/currency/v1/currency_pb";
	import { stakeSumInCurrency, streamCurrencies, streamGroup, summariseStakes } from "../../../utils";
	import type { ExpenseStake } from "../../../../../../../../../gen/lib/ts/common/expensestake/v1/expensestake_pb";
	import { ExpenseStakeService } from "../../../../../../../../../gen/lib/ts/service/expensestake/v1/service_connect";
	import { GroupService } from "../../../../../../../../../gen/lib/ts/service/group/v1/service_connect";
	import type { Group } from "../../../../../../../../../gen/lib/ts/common/group/v1/group_pb";
	import LayoutGrid, {Cell as LayoutCell} from "@smui/layout-grid";
	import { t } from "$lib/localization";
	import DataTable, { Body, Cell, Head, Row } from "@smui/data-table";
	import LinearProgress from "@smui/linear-progress";
	import IconButton from "@smui/icon-button";
	import { Separator } from "@smui/list";
	import Textfield from "@smui/textfield";
	import HelperText from "@smui/textfield/helper-text";
	import Select, {Option} from "@smui/select";
	import Button, { Label } from "@smui/button";

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

		// initially set currency to group currency in form but do not update on changes
		let unsubscribeInitialGroup: Unsubscriber|undefined;
		unsubscribeInitialGroup = group.subscribe((g) => {
			if (unsubscribeInitialGroup) {
				// unsubscribe so that future changes have no effect
				unsubscribeInitialGroup();
				newExpense.currencyId = g?.currencyId ?? '';
			}
		});

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
				streamExpenseStakesInExpense(expensestakeClient, expenseId, abortController, expenseStakes);
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
			newExpense.currencyId = $group?.currencyId ?? '';
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

<LayoutGrid>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<h2>Expenses</h2>
		{#if $group && $currencies && expensesSummary !== undefined}
			<span>Total value: {expensesSummary.toFixed(2)} {$currencies.get($group.currencyId)?.acronym}</span>
		{:else}
			<span>Loading total value...</span>
		{/if}
		
		<DataTable table$aria-label="Group list" style="width: 100%">
			<Head>
				<Row>
					<Cell>Name</Cell>
					<Cell>By</Cell>
					<Cell>Currency</Cell>
					<Cell>Timestamp</Cell>
					<Cell>Action</Cell>
				</Row>
			</Head>
			<Body>
				{#if $expenses && $people}
					{#each [...$expenses] as [eID, expense]}
						{#if expense.expense}
							{@const e = expense.expense}
							<Row on:click={openExpense(eID)}>
								<Cell>{e.name}</Cell>
								<Cell>{$people.get(e.byId)?.person?.name}</Cell>
								<Cell>
									{#if $currencies}
										{@const currency = $currencies.get(e.currencyId)}
										<span>{currency?.acronym} - {currency?.name}</span>
									{/if}
									<LinearProgress
										indeterminate
										closed={!!$currencies}
										aria-label="Currencies are being loaded..."/>
								</Cell>
								<Cell>{e.timestamp?.toDate().toLocaleString()}</Cell>
								<Cell>
									<IconButton
										on:click$stopPropagation={deleteExpense(eID)}
										class="material-icons"
										aria-label="Delete expense">delete</IconButton>
								</Cell>
							</Row>
						{/if}
					{/each}
				{/if}
			</Body>
		
			<LinearProgress
				indeterminate
				closed={!!$expenses && !!people}
				aria-label="Expenses and people are being loaded..."
				slot="progress"
			/>
		</DataTable>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<Separator />
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<h4>New expense</h4>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 6, tablet: 4, phone: 4 }}>
		<Textfield variant="outlined" bind:value={newExpense.name} label="Expense name" style="width: 100%" helperLine$style="width: 100%">
			<HelperText slot="helper">The name of the expense that is to be created</HelperText>
		</Textfield>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 6, tablet: 4, phone: 4 }}>
		{#if $people}
			<Select variant="outlined" bind:value={newExpense.byId} label="Person" style="width: 100%">
				{#each [...$people] as [pID, person]}
					<Option value={pID}>
						{#if person.person}
							{person.person.name}
						{:else}
							<LinearProgress
								indeterminate
								aria-label={$t('expenses.loadingPersonWithId', { personId: pID })} />
						{/if}
					</Option>
				{/each}
			</Select>
		{/if}
		<LinearProgress
			indeterminate
			closed={!!$currencies}
			aria-label="People are being loaded..."/>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		{#if $currencies && newExpense.currencyId}
			<Select variant="outlined" bind:value={newExpense.currencyId} label="Currency" style="width: 100%">
				{#each [...$currencies] as [cID, currency]}
					<Option value={cID}>{currency.acronym} - {currency.name}</Option>
				{/each}
			</Select>
		{/if}
		<LinearProgress
			indeterminate
			closed={!!$currencies && !!newExpense.currencyId}
			aria-label="Currencies are being loaded..."/>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }} style="display: flex; justify-content: center">
		<DatePicker min={new Date(1640995200000)} max={new Date()} bind:value={newExpense.timestamp} />
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }} style="display: flex; justify-content: flex-end">
		<Button on:click={createExpense} touch variant="outlined">
			<Label>Create expense</Label>
		</Button>
	</LayoutCell>
</LayoutGrid>