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
	import LayoutGrid, {Cell as LayoutCell} from "@smui/layout-grid";
	import DataTable, { Body, Cell, Head, Row } from "@smui/data-table";
	import IconButton from "@smui/icon-button";
	import { t } from "$lib/localization";
	import LinearProgress from "@smui/linear-progress";
	import Textfield from "@smui/textfield";
	import HelperText from "@smui/textfield/helper-text";
	import Select, {Option} from "@smui/select";
	import Button, { Label } from "@smui/button";

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
		fractionalValue: 0 as number|undefined
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
			newExpenseStake.fractionalValue = 0;
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

<LayoutGrid>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<h2>Expense stakes</h2>
		
		<DataTable table$aria-label="Expense stake list" style="width: 100%">
			<Head>
				<Row>
					<Cell>ID</Cell>
					<Cell>For</Cell>
					<Cell>Value</Cell>
					<Cell>Action</Cell>
				</Row>
			</Head>
			<Body>
				{#if $expensestakes && $people}
					{#each [...$expensestakes] as [eID, expensestake]}
						{#if expensestake.expensestake}
							<Row>
								<Cell>{eID}</Cell>
								<Cell>{$people.get(expensestake.expensestake.forId)?.person?.name}</Cell>
								<Cell>{marshalExpenseStakeValue(expensestake.expensestake)}</Cell>
								<Cell>
									<IconButton
										on:click$stopPropagation={deleteExpenseStake(eID)}
										class="material-icons"
										aria-label="Delete expense stake">delete</IconButton>
								</Cell>
							</Row>
						{/if}
					{/each}
				{/if}
			</Body>
		
			<LinearProgress
				indeterminate
				closed={!!$expensestakes}
				aria-label="Expense stakes are being loaded..."
				slot="progress"
			/>
		</DataTable>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<h4>New expense stake</h4>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		{#if $people}
			<Select variant="outlined" bind:value={newExpenseStake.forId} label="For" style="width: 100%">
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
			closed={!!$expensestakes}
			aria-label="Expense stakes are being loaded..."/>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 6, tablet: 4, phone: 2 }}>
		<Textfield variant="outlined" bind:value={newExpenseStake.mainValue} type="number" input$min="0" input$step="1" label="Main value" style="width: 100%" helperLine$style="width: 100%">
			<HelperText slot="helper">The main value without the fractional part</HelperText>
		</Textfield>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 6, tablet: 4, phone: 2 }}>
		<Textfield variant="outlined" bind:value={newExpenseStake.fractionalValue} disabled="{fractionalDisabled}" input$min={0} input$max={99} input$step={1} type="number" label="Fractional value" style="width: 100%" helperLine$style="width: 100%">
			<HelperText slot="helper">The fractional part of the value</HelperText>
		</Textfield>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }} style="display: flex; justify-content: flex-end">
		<Button on:click={createExpenseStake} touch variant="outlined">
			<Label>Create expense stake</Label>
		</Button>
	</LayoutCell>
</LayoutGrid>