<script lang="ts">
	import { writable, type Writable } from "svelte/store";
	import type { ExpenseStake } from "../../../../../../../../../../../gen/lib/ts/common/expensestake/v1/expensestake_pb";
	import { ExpenseStakeService } from "../../../../../../../../../../../gen/lib/ts/service/expensestake/v1/service_connect";
	import { createPromiseClient, type Transport } from "@bufbuild/connect";
	import type { Expense } from "../../../../../../../../../../../gen/lib/ts/common/expense/v1/expense_pb";
	import { onDestroy, onMount } from "svelte";
	import { streamExpenseStakes } from "./utils";
	import type { Person } from "../../../../../../../../../../../gen/lib/ts/common/person/v1/person_pb";

    export let expense: Expense;
    export let transport: Transport;
    export let people: Writable<
		Map<string, {person?: Person, abortController: AbortController}> | undefined
	>;

	const newExpenseStake = {
		forId: '',
		value: '',
	};

    const expensestakeClient = createPromiseClient(ExpenseStakeService, transport);
	let expensestakes: Writable<
		Map<string, {expensestake?: ExpenseStake, abortController: AbortController}> | undefined
	> = writable();
	const abortController = new AbortController();

    onDestroy(() => {
        abortController.abort();
    });

    onMount(async () => {
        streamExpenseStakes(expensestakeClient, expense.id, abortController, expensestakes);
    });

    async function createExpenseStake() {
		try {
            const [mainValue, fractionalValue] = unmarshalValue(newExpenseStake.value);
			const res = await expensestakeClient.createExpenseStake({
                expenseId: expense.id,
				forId: newExpenseStake.forId,
				mainValue: mainValue,
				fractionalValue: fractionalValue,
			});
			console.log('Created expense stake', res.id);

			newExpenseStake.forId = '';
			newExpenseStake.value = '';
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

    function marshalValue(expensestake: ExpenseStake): string {
		let fractionalValue: string;
		if (expensestake.fractionalValue) {
			fractionalValue = expensestake.fractionalValue.toString();
			if (fractionalValue.length === 1) {
				fractionalValue = `0${fractionalValue}`;
			}
		} else {
			fractionalValue = '00';
		}
        return `${expensestake.mainValue}.${fractionalValue}`;
    }

    function unmarshalValue(value: string): [number, number] {
		if (!/^\d+(\.\d{2})?$/.test(value)) {
			throw 'Value input is wrongly formatted';
		}
        const valueParts = value.split(".");
        if (valueParts.length > 2) {
            throw 'Value input has to many parts';
        }
        const mainValue = parseInt(valueParts[0]);
        if (!mainValue) {
            throw 'Main value input is invalid';
        }
        let fractionalValue: number;
        if (valueParts.length == 2) {
            fractionalValue = parseInt(valueParts[1]);
            if (!fractionalValue) {
                throw 'Fractional value input is invalid';
            }
        } else {
            fractionalValue = 0;
        }
        return [mainValue, fractionalValue];
    }

	function summariseStakes(expensestakes: Map<string, {
		expensestake?: ExpenseStake | undefined;
		abortController: AbortController;
	}>): string {
		const mainValues: number[] = [];
		const fractionalValues: number[] = [];
		for (let [id, stake] of expensestakes) {
			const expensestake = stake.expensestake
			if (expensestake) {
				mainValues.push(expensestake.mainValue);
				if (expensestake.fractionalValue) {
					fractionalValues.push(expensestake.fractionalValue);
				}
			} else {
				return 'Loading expense stakes...';
			}
		}
		let mainSummary = mainValues.reduce((partialSum, a) => partialSum + a, 0);
		const fractionalSummary = fractionalValues.reduce((partialSum, a) => partialSum + a, 0)
		mainSummary += Math.floor(fractionalSummary / 100);
		const fractionalRemainder = fractionalSummary % 100;
		return `${mainSummary}.${fractionalRemainder}`;
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
						<td>{marshalValue(expensestake.expensestake)}</td>
						<td><button on:click|stopPropagation={deleteExpenseStake(eID)}>Delete</button></td>
					</tr>
				{:else}
					<tr>Loading expense with ID {eID}...</tr>
				{/if}
			{/each}
			<tr>
				<td></td>
				<td></td>
				<td>{summariseStakes($expensestakes)}</td>
				<td></td>
			</tr>
			<tr>
				<td></td>
				<td>
                    <select bind:value={newExpenseStake.forId}>
                        {#each [...$people] as [pID, person]}
                            <option value={pID}>{person.person?.name}</option>
                        {/each}
                    </select>
				</td>
				<td><input type="text" placeholder="Value" bind:value={newExpenseStake.value}/></td>
				<td><button on:click={createExpenseStake}>Create expense stake</button></td>
			</tr>
		{:else}
			<tr>Loading expenses...</tr>
		{/if}
	</tbody>
</table>