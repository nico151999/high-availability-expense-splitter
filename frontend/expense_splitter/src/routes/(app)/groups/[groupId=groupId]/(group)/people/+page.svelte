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
	import LayoutGrid, {Cell as LayoutCell} from "@smui/layout-grid";
	import DataTable, { Body, Cell, Head, Row } from "@smui/data-table";
	import { t } from "$lib/localization";
	import IconButton from "@smui/icon-button";
	import LinearProgress from "@smui/linear-progress";
	import { Separator } from "@smui/list";
	import Textfield from "@smui/textfield";
	import HelperText from "@smui/textfield/helper-text";
	import Button, { Label } from "@smui/button";
	import type { Currency } from "../../../../../../../../../gen/lib/ts/common/currency/v1/currency_pb";

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
	let groupCurrency: Currency|undefined;

	$: if ($group) {
		fetchGroupCurrency($group).then(g => groupCurrency = g);
	}

	async function fetchGroupCurrency(group: Group): Promise<Currency|undefined> {
		try {
			return (await currencyClient.getCurrency({id: group.currencyId})).currency;
		} catch (e) {
			console.error('An error occurred trying to fetch group currency. Trying anew in 5 seconds.', e);
			await new Promise(resolve => setTimeout(resolve, 5000));
			return (await fetchGroupCurrency(group));
		}
	}

	const newPerson = {
		name: '',
	};

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
                name: newPerson.name
			});
			console.log('Created person', res.id);

            newPerson.name = '';
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

<LayoutGrid>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<h2>People</h2>
		
		<DataTable table$aria-label="Person list" style="width: 100%">
			<Head>
				<Row>
					<Cell>Name</Cell>
					<Cell>Account</Cell>
					<Cell>Action</Cell>
				</Row>
			</Head>
			<Body>
				{#if $people}
					{#each [...$people] as [pID, person]}
						{#if person.person}
							<Row on:click={openPerson(pID)}>
								<Cell>{person.person.name}</Cell>
								<Cell>
									{#if $accountsPerPerson && groupCurrency}
										{@const accountPerPerson = $accountsPerPerson.get(pID)}
										{#if accountPerPerson !== undefined}
											{accountPerPerson.toFixed(2)} {groupCurrency.acronym}
										{/if}
									{/if}
									<LinearProgress
										indeterminate
										closed={$accountsPerPerson?.get(pID) !== undefined}
										aria-label="Account is being loaded..." />
								</Cell>
								<Cell>
									<IconButton
										on:click$stopPropagation={deletePerson(pID)}
										class="material-icons"
										aria-label="Delete person">delete</IconButton>
								</Cell>
							</Row>
						{/if}
					{/each}
				{/if}
			</Body>
		
			<LinearProgress
				indeterminate
				closed={!!$people}
				aria-label="People are being loaded..."
				slot="progress"
			/>
		</DataTable>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<Separator />
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<h4>New person</h4>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 6, tablet: 4, phone: 4 }}>
		<Textfield variant="outlined" bind:value={newPerson.name} label="Person name" style="width: 100%" helperLine$style="width: 100%">
			<HelperText slot="helper">The name of the person that is to be created</HelperText>
		</Textfield>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 6, tablet: 6, phone: 4 }} style="display: flex; justify-content: flex-end">
		<Button on:click={createPerson} touch variant="outlined">
			<Label>Create person</Label>
		</Button>
	</LayoutCell>
</LayoutGrid>