import { Code, ConnectError, type PromiseClient } from "@bufbuild/connect";
import { get, writable, type Unsubscriber, type Writable } from "svelte/store";
import type { Person } from "../../../../../../../../../gen/lib/ts/common/person/v1/person_pb";
import type { PersonService } from "../../../../../../../../../gen/lib/ts/service/person/v1/service_connect";
import type { ExpenseStake } from "../../../../../../../../../gen/lib/ts/common/expensestake/v1/expensestake_pb";
import type { Expense } from "../../../../../../../../../gen/lib/ts/common/expense/v1/expense_pb";
import { marshalExpenseStakeValue, stakeSumInCurrency, summariseStakes } from "../../../utils";
import type { CurrencyService } from "../../../../../../../../../gen/lib/ts/service/currency/v1/service_connect";
import type { Group } from "../../../../../../../../../gen/lib/ts/common/group/v1/group_pb";

export async function streamPerson(
	personClient: PromiseClient<typeof PersonService>,
	personID: string,
	abortController: AbortController,
    person: Writable<Person | undefined>
): Promise<boolean> {
	try {
		for await (const pRes of personClient.streamPerson({id: personID}, {signal: abortController.signal})) {
			if (pRes.update.case === 'stillAlive') {
				continue;
			}
            person.set(pRes.update.value);
		}
    } catch (e) {
        if (e instanceof ConnectError) {
            if (e.code === Code.Canceled) {
                console.log(`Intentionally aborted person ${personID} stream`);
                return true;
            } else if (e.code === Code.DataLoss) {
                console.log(`Person with ID ${personID} no longer exists`);
                return false;
            }
        }
        console.error('An error occurred trying to stream person.', e);
    }
    console.log(`Ended person ${personID} stream. Starting new one in 5 seconds.`);
    await new Promise(resolve => setTimeout(resolve, 5000));
    return await streamPerson(personClient, personID, abortController, person);
}

export async function streamPeople(
    personClient: PromiseClient<typeof PersonService>,
    groupId: string,
    abortController: AbortController,
    peopleStore: Writable<Map<string, {person?: Person, abortController: AbortController}> | undefined>
) {
    try {
        for await (const cIDsRes of personClient.streamPersonIdsInGroup({
            groupId: groupId
        }, {signal: abortController.signal})) {
            if (cIDsRes.update.case === 'stillAlive') {
                continue;
            }
            const personIDs = cIDsRes.update.value!.ids;
            let people = get(peopleStore);
            if (people === undefined) {
                people = new Map();
            }
            for (const cID of people.keys()) {
                if (personIDs.includes(cID)) {
                    // remove element from items that are to be processed because it already exists
                    personIDs.splice(personIDs.indexOf(cID), 1);
                } else {
                    // remove people that are not present any more
                    people!.get(cID)!.abortController.abort();
                    people!.delete(cID);
                }
            }
            for (const pID of personIDs) {
                const abortController = new AbortController();
                people.set(pID, {
                    abortController: abortController
                });
                const person: Writable<Person | undefined> = writable();
                person.subscribe((e) => {
                    peopleStore.set(people?.set(pID, {
                        abortController: abortController,
                        person: e
                    }));
                });
                streamPerson(personClient, pID, abortController, person);
            }
            peopleStore.set(people);
        }
    } catch (e) {
        if (e instanceof ConnectError && e.code === Code.Canceled) {
            console.log('Intentionally aborted people stream');
            return;
        }
        console.error('An error occurred trying to stream people', e);
    } finally {
        const people = get(peopleStore);
        if (people) {
            for (let [_, person] of people) {
                person.abortController.abort();
            }
        }
        peopleStore.set(undefined);
    }
    console.log(`Ended people stream. Starting new one in 5 seconds.`);
    await new Promise(resolve => setTimeout(resolve, 5000));
    await streamPeople(personClient, groupId, abortController, peopleStore);
}

export function streamAccountsPerPerson(
    abortController: AbortController,
    currencyClient: PromiseClient<typeof CurrencyService>,
    groupStore: Writable<Group|undefined>,
    peopleStore: Writable<Map<string, {person?: Person, abortController: AbortController}> | undefined>,
    expensesStore: Writable<
        Map<string, {expense?: Expense, abortController: AbortController}> | undefined
    >,
    expensestakesStore: Writable<
        Map<string, {expensestake?: ExpenseStake, abortController: AbortController}> | undefined
    >,
    accountsPerPerson: Writable<
        Map<string, number> | undefined
    >
) {
    let unsubscribeGroup: Unsubscriber|undefined;
    let unsubscribePeople: Unsubscriber|undefined;
    let unsubscribeExpenses: Unsubscriber|undefined;
    let unsubscribeExpensestakes: Unsubscriber|undefined;
    abortController.signal.addEventListener("abort", () => {
        if (unsubscribeGroup) { unsubscribeGroup() }
        if (unsubscribePeople) { unsubscribePeople() }
        if (unsubscribeExpenses) { unsubscribeExpenses() }
        if (unsubscribeExpensestakes) { unsubscribeExpensestakes() }
    });
    const recalculate = async () => {
        const newAccounts: Map<string, number> = new Map();
        const group = get(groupStore);
        const people = get(peopleStore);
        const expenses = get(expensesStore);
        const expenseStakes = get(expensestakesStore);
        if (!group || !people || !expenses || !expenseStakes) {
            accountsPerPerson.set(newAccounts);
            return;
        }
        const exchangeRatePerExpense: Map<string, number> = new Map();
        try {
            for (const personID of people.keys()) {
                let totalStakeSum = 0;
                for (const [expenseID, expense] of expenses) {
                    const stakes: ExpenseStake[] = [];
                    const e = expense.expense;
                    if (!e) {
                        accountsPerPerson.set(newAccounts);
                        return;
                    }
                    if (e.byId === personID) {
                        for (const [_, expenseStake] of expenseStakes) {
                            const es = expenseStake.expensestake;
                            if (!es) {
                                accountsPerPerson.set(newAccounts);
                                return;
                            }
                            if (es.expenseId === expenseID) {
                                stakes.push(es);
                            }
                        }
                    }
                    const stakeSum = summariseStakes(stakes);
                    let exchangeRate = await getExchangeRateForExpense(currencyClient, exchangeRatePerExpense, e, group.currencyId);
                    totalStakeSum += stakeSumInCurrency(exchangeRate, stakeSum);
                }
    
                for (const [_, expensestake] of expenseStakes) {
                    const es = expensestake.expensestake;
                    if (!es) {
                        accountsPerPerson.set(newAccounts);
                        return;
                    }
                    if (es.forId === personID) {
                        const e = expenses.get(es.expenseId)?.expense;
                        if (!e) {
                            accountsPerPerson.set(newAccounts);
                            return;
                        }
                        const stake = marshalExpenseStakeValue(es);
                        let exchangeRate = await getExchangeRateForExpense(currencyClient, exchangeRatePerExpense, e, group.currencyId);
                        totalStakeSum -= stakeSumInCurrency(exchangeRate, stake);
                    }
                }
                accountsPerPerson.set(
                    newAccounts.set(personID, totalStakeSum)
                );
            }
        } catch (e) {
            console.error(e);
        }
    };
    unsubscribeGroup = groupStore.subscribe(recalculate);
    unsubscribePeople = peopleStore.subscribe(recalculate);
    unsubscribeExpenses = expensesStore.subscribe(recalculate);
    unsubscribeExpensestakes = expensestakesStore.subscribe(recalculate);
}

async function getExchangeRateForExpense(
    currencyClient: PromiseClient<typeof CurrencyService>,
    cache: Map<string, number>,
    expense: Expense,
    groupCurrencyId: string
): Promise<number> {
    if (expense.currencyId === groupCurrencyId) {
        return 1;
    }
    if (cache.has(expense.id)) {
        return cache.get(expense.id)!;
    }
    const res = await currencyClient.getExchangeRate({
        sourceCurrencyId: expense.currencyId,
        destinationCurrencyId: groupCurrencyId,
        timestamp: expense.timestamp
    });
    cache.set(expense.id, res.rate);
    return res.rate;
}