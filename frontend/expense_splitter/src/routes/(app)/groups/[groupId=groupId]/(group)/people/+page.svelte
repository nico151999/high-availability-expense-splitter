<script lang="ts">
	import { createPromiseClient } from "@bufbuild/connect";
	import type { PageData } from "./$types";
	import { PersonService } from "../../../../../../../../../gen/lib/ts/service/person/v1/service_connect";
	import type { Person } from "../../../../../../../../../gen/lib/ts/common/person/v1/person_pb";
	import { onDestroy, onMount } from "svelte";
	import { writable, type Writable } from "svelte/store";
	import { goto } from "$app/navigation";
	import { streamAccountsPerPerson, streamPeople } from "./utils";
	import { ExpenseStakeService } from "../../../../../../../../../gen/lib/ts/service/expensestake/v1/service_connect";
	import { CurrencyService } from "../../../../../../../../../gen/lib/ts/service/currency/v1/service_connect";
	import { streamExpenseStakesInGroup, streamGroup } from "../../../utils";
	import { ExpenseService } from "../../../../../../../../../gen/lib/ts/service/expense/v1/service_connect";
	import type { Expense } from "../../../../../../../../../gen/lib/ts/common/expense/v1/expense_pb";
	import { streamExpenses } from "../expenses/utils";
	import type { ExpenseStake } from "../../../../../../../../../gen/lib/ts/common/expensestake/v1/expensestake_pb";
	import { GroupService } from "../../../../../../../../../gen/lib/ts/service/group/v1/service_connect";
	import type { Group } from "../../../../../../../../../gen/lib/ts/common/group/v1/group_pb";

	export let data: PageData;

	const personClient = createPromiseClient(PersonService, data.grpcWebTransport);
	let people: Writable<
		Map<string, {person?: Person, abortController: AbortController}> | undefined
	> = writable();
	const abortController = new AbortController();

	const groupClient = createPromiseClient(GroupService, data.grpcWebTransport);
	let group: Writable<Group|undefined> = writable();
	const groupAbortController = new AbortController();

	const expensestakeClient = createPromiseClient(ExpenseStakeService, data.grpcWebTransport);
	let expenseStakesInGroup: Writable<
		Map<string, {expensestake?: ExpenseStake, abortController: AbortController}> | undefined
	> = writable();
	const expenseStakesInGroupAbortController = new AbortController();

	const expenseClient = createPromiseClient(ExpenseService, data.grpcWebTransport);
	let expensesInGroup: Writable<
		Map<string, {expense?: Expense, abortController: AbortController}> | undefined
	> = writable();
	const expenseAbortController = new AbortController();

	// the account values per person in the group's currency
	let accountsPerPerson: Writable<
		Map<string, number> | undefined
	> = writable();
	const accountsPerPersonAbortController = new AbortController();

	const currencyClient = createPromiseClient(CurrencyService, data.grpcWebTransport);


	const newPerson = writable({
		name: '',
	});

	onDestroy(() => {
		abortController.abort();
		groupAbortController.abort();
		expenseAbortController.abort();
		expenseStakesInGroupAbortController.abort();
		accountsPerPersonAbortController.abort();
	});

	onMount(async () => {
		streamPeople(personClient, data.groupId, abortController, people);
		streamExpenses(expenseClient, data.groupId, expenseAbortController, expensesInGroup);
		streamExpenseStakesInGroup(expensestakeClient, data.groupId, expenseStakesInGroupAbortController, expenseStakesInGroup)
        streamAccountsPerPerson(accountsPerPersonAbortController, currencyClient, group, people, expensesInGroup, expenseStakesInGroup, accountsPerPerson);
		const res = await streamGroup(groupClient, data.groupId, groupAbortController, group);
        if (!res) {
            console.error('Group no longer exists');
        }
	});

	async function createPerson() {
		try {
			const res = await personClient.createPerson({
                groupId: data.groupId,
                name: $newPerson.name
			});
			console.log('Created person', res.id);

            newPerson.set({name: ''});
		} catch (e) {
			console.error(`An error occurred trying to create person in group ${data.groupId}`, e);
		}
	}

	function deletePerson(personId: string) {
		return async () => {
			try {
				await personClient.deletePerson({id: personId});
				console.log('Deleted person');
			} catch (e) {
				console.error(`An error occurred trying to delete person ${personId} in group ${data.groupId}`, e);
			}
		}
	}

	function openPerson(personId: string) {
		return () => {
			goto(`./people/${personId}`);
		}
	}
</script>

<h2>Your people in group {data.groupId}</h2>

<table>
	<thead>
		<th>ID</th>
		<th>Name</th>
		<th>Account</th>
		<th>Action</th>
	</thead>
	<tbody>
		{#if $people}
			{#each [...$people] as [pID, person]}
				{#if person.person}
					<tr on:click={openPerson(pID)}>
						<td>{pID}</td>
						<td>{person.person?.name}</td>
						<td>{$accountsPerPerson?.get(pID)?.toFixed(2) ?? 'Loading account...'}</td>
						<td><button on:click|stopPropagation={deletePerson(pID)}>Delete</button></td>
					</tr>
				{:else}
					<tr>Loading person with ID {pID}...</tr>
				{/if}
			{/each}
		{:else}
			<tr>Loading people...</tr>
		{/if}
		<tr>
			<td></td>
			<td><input type="text" placeholder="Person name" bind:value={$newPerson.name}/></td>
			<td><button on:click={createPerson}>Create person</button></td>
		</tr>
	</tbody>
</table>