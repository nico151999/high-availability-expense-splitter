<script lang="ts">
	import { createPromiseClient } from "@bufbuild/connect";
	import type { PageData } from "./$types";
	import { ExpenseService } from "../../../../../../../../../gen/lib/ts/service/expense/v1/service_connect";
	import type { Expense } from "../../../../../../../../../gen/lib/ts/common/expense/v1/expense_pb";
	import { onDestroy, onMount } from "svelte";
	import { writable, type Writable } from "svelte/store";
	import { goto } from "$app/navigation";
	import { streamExpenses } from "./utils";
	import { Timestamp } from "@bufbuild/protobuf";
	import { DateInput } from 'date-picker-svelte';
	import type { Person } from "../../../../../../../../../gen/lib/ts/common/person/v1/person_pb";
	import { PersonService } from "../../../../../../../../../gen/lib/ts/service/person/v1/service_connect";
	import { streamPeople } from "../people/utils";

	export let data: PageData;

	const expenseClient = createPromiseClient(ExpenseService, data.grpcWebTransport);
	let expenses: Writable<
		Map<string, {expense?: Expense, abortController: AbortController}> | undefined
	> = writable();
	const abortController = new AbortController();

	const personClient = createPromiseClient(PersonService, data.grpcWebTransport);
	let people: Writable<
		Map<string, {person?: Person, abortController: AbortController}> | undefined
	> = writable();
	const peopleAbortController = new AbortController();

	let timestampValid: boolean | undefined;
	const newExpense = {
		name: '',
		byId: '',
		currencyId: '',
		timestamp: new Date()
	};

	onDestroy(() => {
		abortController.abort();
		if ($expenses) {
			for (const [_, expense] of $expenses) {
				expense.abortController.abort();
			}
		}
		peopleAbortController.abort();
		if ($people) {
			for (const [_, person] of $people) {
				person.abortController.abort();
			}
		}
	});

	onMount(() => {
		streamExpenses(expenseClient, data.groupId, abortController, expenses);
		streamPeople(personClient, data.groupId, abortController, people);
	});

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

	function openExpense(expenseId: string) {
		return () => {
			goto(`./expenses/${expenseId}`);
		}
	}
</script>

<h2>Your expenses in group {data.groupId}</h2>

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
					<tr on:click={openExpense(pID)}>
						<td>{pID}</td>
						<td>{expense.expense.name}</td>
						<td>{$people.get(expense.expense.byId)?.person?.name}</td>
						<td>{expense.expense.currencyId}</td>
						<td>{expense.expense.timestamp?.toDate().toLocaleString()}</td>
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
					{#if $people}
						<select bind:value={newExpense.byId}>
							{#each [...$people] as [pID, person]}
								<option value={pID}>{person.person?.name}</option>
							{/each}
						</select>
					{:else}
						<span>Loading people...</span>
					{/if}
				</td>
				<td><input type="text" placeholder="Currency" bind:value={newExpense.currencyId}/></td> <!-- TODO: dropdown -->
				<td><DateInput min={new Date(1640995200000)} max={new Date()} bind:value={newExpense.timestamp} bind:valid={timestampValid}/></td>
				<td><button on:click={createExpense}>Create expense</button></td>
			</tr>
		{:else}
			<tr>Loading expenses...</tr>
		{/if}
	</tbody>
</table>