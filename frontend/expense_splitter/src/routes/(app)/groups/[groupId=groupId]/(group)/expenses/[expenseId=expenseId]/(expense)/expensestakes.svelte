<script lang="ts">
	import { writable, type Unsubscriber, type Writable } from "svelte/store";
	import type { ExpenseStake } from "../../../../../../../../../../../gen/lib/ts/common/expensestake/v1/expensestake_pb";
	import { ExpenseStakeService } from "../../../../../../../../../../../gen/lib/ts/service/expensestake/v1/service_connect";
	import { createPromiseClient, type Transport } from "@bufbuild/connect";
	import type { Expense } from "../../../../../../../../../../../gen/lib/ts/common/expense/v1/expense_pb";
	import { onDestroy, onMount } from "svelte";
	import type { Person } from "../../../../../../../../../../../gen/lib/ts/common/person/v1/person_pb";
	import { streamExpenseStakesInExpense } from "../../utils";
	import { marshalExpenseStakeValue, summariseStakes } from "../../../../../utils";

    export let expense: Expense;
    export let transport: Transport;
    export let people: Writable<
		Map<string, {person?: Person, abortController: AbortController}> | undefined
	>;
	export let stakeSum = '0';
	export let fractionalDisabled = false;

	const newExpenseStake = {
		forId: '',
		mainValue: 0,
		fractionalValue: undefined as number|undefined
	};

    const expensestakeClient = createPromiseClient(ExpenseStakeService, transport);
	let expensestakes: Writable<
		Map<string, {expensestake?: ExpenseStake, abortController: AbortController}> | undefined
	> = writable();
	let unsubscribeExpensestake: Unsubscriber|undefined;
	const abortController = new AbortController();

    onDestroy(() => {
        abortController.abort();
		if (unsubscribeExpensestake) {unsubscribeExpensestake()}
    });

    onMount(async () => {
        streamExpenseStakesInExpense(expensestakeClient, expense.id, abortController, expensestakes);
		unsubscribeExpensestake = expensestakes.subscribe((expensestakes) => {
			if (!expensestakes) {
				return;
			}
			const stakes: ExpenseStake[] = [];
			for (let stake of expensestakes.values()) {
				if (stake.expensestake) {
					stakes.push(stake.expensestake);
				}
			}
			stakeSum = summariseStakes(stakes);
		});
    });

    async function createExpenseStake() {
		try {
			const res = await expensestakeClient.createExpenseStake({
                expenseId: expense.id,
				forId: newExpenseStake.forId,
				mainValue: newExpenseStake.mainValue,
				fractionalValue: newExpenseStake.fractionalValue,
			});
			console.log('Created expense stake', res.id);

			newExpenseStake.forId = '';
			newExpenseStake.mainValue = 0;
			newExpenseStake.fractionalValue = undefined
		} catch (e) {
			console.error(`An error occurred trying to create expensestake in expense ${expense.id}`, e);
		}
	}

	function deleteExpenseStake(expensestakeId: string) {
		return async () => {
			try {
				await expensestakeClient.deleteExpenseStake({id: expensestakeId});
				console.log('Deleted expense stake');
			} catch (e) {
				console.error(`An error occurred trying to delete expense stake ${expensestakeId} in expense ${expense.id}`, e);
			}
		}
	}
</script>

<h2>Your expense stakes in expense {expense.id}</h2>

<table>
	<thead>
		<th>ID</th>
		<th>For</th>
		<th>Value</th>
		<th>Action</th>
	</thead>
	<tbody>
		{#if $expensestakes && $people}
			{#each [...$expensestakes] as [eID, expensestake]}
				{#if expensestake.expensestake}
					<tr>
						<td>{eID}</td>
						<td>{$people.get(expensestake.expensestake.forId)?.person?.name}</td>
						<td>{marshalExpenseStakeValue(expensestake.expensestake)}</td>
						<td><button on:click|stopPropagation={deleteExpenseStake(eID)}>Delete</button></td>
					</tr>
				{:else}
					<tr>Loading expense with ID {eID}...</tr>
				{/if}
			{/each}
			<tr>
				<td></td>
				<td>
                    <select bind:value={newExpenseStake.forId}>
                        {#each [...$people] as [pID, person]}
                            <option value={pID}>{person.person?.name}</option>
                        {/each}
                    </select>
				</td>
				<td>
					<input type="number" placeholder="Main value" min="0" step="1" bind:value={newExpenseStake.mainValue}/>
					<input disabled="{fractionalDisabled}" min="0" max="99" step="1" type="number" placeholder="Fractional value" bind:value={newExpenseStake.fractionalValue}/>
				</td>
				<td><button on:click={createExpenseStake}>Create expense stake</button></td>
			</tr>
		{:else}
			<tr>Loading expenses...</tr>
		{/if}
	</tbody>
</table>