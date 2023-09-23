import { Code, ConnectError, type PromiseClient } from "@bufbuild/connect";
import { get, writable, type Writable } from "svelte/store";
import type { Expense } from "../../../../../../../../../gen/lib/ts/common/expense/v1/expense_pb";
import type { ExpenseService } from "../../../../../../../../../gen/lib/ts/service/expense/v1/service_connect";
import type { ExpenseStakeService } from "../../../../../../../../../gen/lib/ts/service/expensestake/v1/service_connect";
import type { ExpenseStake } from "../../../../../../../../../gen/lib/ts/common/expensestake/v1/expensestake_pb";
import type { ExpenseCategoryRelationService } from "../../../../../../../../../gen/lib/ts/service/expensecategoryrelation/v1/service_connect";
import type { CurrencyService } from "../../../../../../../../../gen/lib/ts/service/currency/v1/service_connect";
import type { Category } from "../../../../../../../../../gen/lib/ts/common/category/v1/category_pb";
import { streamCategory } from "../categories/utils";
import type { CategoryService } from "../../../../../../../../../gen/lib/ts/service/category/v1/service_connect";

export async function streamExpense(
	expenseClient: PromiseClient<typeof ExpenseService>,
	expenseID: string,
	abortController: AbortController,
    expense: Writable<Expense | undefined>
): Promise<boolean> {
	try {
		for await (const pRes of expenseClient.streamExpense({id: expenseID}, {signal: abortController.signal})) {
			if (pRes.update.case === 'stillAlive') {
				continue;
			}
            expense.set(pRes.update.value);
		}
    } catch (e) {
        if (e instanceof ConnectError) {
            if (e.code === Code.Canceled) {
                console.log(`Intentionally aborted expense ${expenseID} stream`);
                return true;
            } else if (e.code === Code.DataLoss) {
                console.log(`Expense with ID ${expenseID} no longer exists`);
                return false;
            }
        }
        console.error('An error occurred trying to stream expense.', e);
    }
    console.log(`Ended expense ${expenseID} stream. Starting new one in 5 seconds.`);
    await new Promise(resolve => setTimeout(resolve, 5000));
    return await streamExpense(expenseClient, expenseID, abortController, expense);
}

export async function streamExpenses(
    expenseClient: PromiseClient<typeof ExpenseService>,
    groupId: string,
    abortController: AbortController,
    expensesStore: Writable<Map<string, {expense?: Expense, abortController: AbortController}> | undefined>
) {
    try {
        for await (const cIDsRes of expenseClient.streamExpenseIdsInGroup({
            groupId: groupId
        }, {signal: abortController.signal})) {
            if (cIDsRes.update.case === 'stillAlive') {
                continue;
            }
            const expenseIDs = cIDsRes.update.value!.ids;
            let expenses = get(expensesStore);
            if (expenses === undefined) {
                expenses = new Map();
            }
            for (const cID of expenses.keys()) {
                if (expenseIDs.includes(cID)) {
                    // remove element from items that are to be processed because it already exists
                    expenseIDs.splice(expenseIDs.indexOf(cID), 1);
                } else {
                    // remove expenses that are not present any more
                    expenses!.get(cID)!.abortController.abort();
                    expenses!.delete(cID);
                }
            }
            for (const pID of expenseIDs) {
                const abortController = new AbortController();
                expenses.set(pID, {
                    abortController: abortController
                });
                const expense: Writable<Expense | undefined> = writable();
                expense.subscribe((e) => {
                    expensesStore.set(expenses?.set(pID, {
                        abortController: abortController,
                        expense: e
                    }));
                });
                streamExpense(expenseClient, pID, abortController, expense);
            }
            expensesStore.set(expenses);
        }
    } catch (e) {
        if (e instanceof ConnectError && e.code === Code.Canceled) {
            console.log('Intentionally aborted expenses stream');
            return;
        }
        console.error('An error occurred trying to stream expenses', e);
    } finally {
        const expenses = get(expensesStore);
        if (expenses) {
            for (let [_, expense] of expenses) {
                expense.abortController.abort();
            }
        }
        expensesStore.set(undefined);
    }
    console.log(`Ended expenses stream. Starting new one in 5 seconds.`);
    await new Promise(resolve => setTimeout(resolve, 5000));
    await streamExpenses(expenseClient, groupId, abortController, expensesStore);
}

export async function streamCategoriesForExpense(
    relationClient: PromiseClient<typeof ExpenseCategoryRelationService>,
    expenseId: string,
    abortController: AbortController,
    categoryIdsStore: Writable<string[] | undefined>
) {
    try {
        for await (const cIDsRes of relationClient.streamCategoryIdsForExpense({
            expenseId: expenseId
        }, {signal: abortController.signal})) {
            if (cIDsRes.update.case === 'stillAlive') {
                continue;
            }
            const categoryIDs = cIDsRes.update.value!.ids;
            categoryIdsStore.set(categoryIDs);
        }
    } catch (e) {
        if (e instanceof ConnectError && e.code === Code.Canceled) {
            console.log('Intentionally aborted categories for expense stream');
            return;
        }
        console.error('An error occurred trying to stream categories for expense', e);
    } finally {
        categoryIdsStore.set(undefined);
    }
    console.log(`Ended categories for expense stream. Starting new one in 5 seconds.`);
    await new Promise(resolve => setTimeout(resolve, 5000));
    await streamCategoriesForExpense(relationClient, expenseId, abortController, categoryIdsStore);
}

export async function streamExpenseStake(
	expensestakeClient: PromiseClient<typeof ExpenseStakeService>,
	expensestakeID: string,
	abortController: AbortController,
    expensestake: Writable<ExpenseStake | undefined>
): Promise<boolean> {
	try {
		for await (const pRes of expensestakeClient.streamExpenseStake({id: expensestakeID}, {signal: abortController.signal})) {
			if (pRes.update.case === 'stillAlive') {
				continue;
			}
            expensestake.set(pRes.update.value);
		}
    } catch (e) {
        if (e instanceof ConnectError) {
            if (e.code === Code.Canceled) {
                console.log(`Intentionally aborted expensestake ${expensestakeID} stream`);
                return true;
            } else if (e.code === Code.DataLoss) {
                console.log(`ExpenseStake with ID ${expensestakeID} no longer exists`);
                return false;
            }
        }
        console.error('An error occurred trying to stream expensestake.', e);
    }
    console.log(`Ended expensestake ${expensestakeID} stream. Starting new one in 5 seconds.`);
    await new Promise(resolve => setTimeout(resolve, 5000));
    return await streamExpenseStake(expensestakeClient, expensestakeID, abortController, expensestake);
}

export async function streamExpenseStakes(
    expensestakeClient: PromiseClient<typeof ExpenseStakeService>,
    expenseId: string,
    abortController: AbortController,
    expensestakesStore: Writable<Map<string, {expensestake?: ExpenseStake, abortController: AbortController}> | undefined>
) {
    try {
        for await (const cIDsRes of expensestakeClient.streamExpenseStakeIdsInExpense({
            expenseId: expenseId
        }, {signal: abortController.signal})) {
            if (cIDsRes.update.case === 'stillAlive') {
                continue;
            }
            const expensestakeIDs = cIDsRes.update.value!.ids;
            let expensestakes = get(expensestakesStore);
            if (expensestakes === undefined) {
                expensestakes = new Map();
            }
            for (const cID of expensestakes.keys()) {
                if (expensestakeIDs.includes(cID)) {
                    // remove element from items that are to be processed newly because it already exists
                    expensestakeIDs.splice(expensestakeIDs.indexOf(cID), 1);
                } else {
                    // remove expensestakes that are not present any more
                    expensestakes!.get(cID)!.abortController.abort();
                    expensestakes!.delete(cID);
                }
            }
            for (const pID of expensestakeIDs) {
                const abortController = new AbortController();
                expensestakes.set(pID, {
                    abortController: abortController
                });
                const expensestake: Writable<ExpenseStake | undefined> = writable();
                expensestake.subscribe((e) => {
                    expensestakesStore.set(expensestakes?.set(pID, {
                        abortController: abortController,
                        expensestake: e
                    }));
                });
                streamExpenseStake(expensestakeClient, pID, abortController, expensestake);
            }
            expensestakesStore.set(expensestakes);
        }
    } catch (e) {
        if (e instanceof ConnectError && e.code === Code.Canceled) {
            console.log('Intentionally aborted expensestakes stream');
            return;
        }
        console.error('An error occurred trying to stream expensestakes', e);
    } finally {
        const expensestakes = get(expensestakesStore);
        if (expensestakes) {
            for (let [_, expensestake] of expensestakes) {
                expensestake.abortController.abort();
            }
        }
        expensestakesStore.set(undefined);
    }
    console.log(`Ended expensestakes stream. Starting new one in 5 seconds.`);
    await new Promise(resolve => setTimeout(resolve, 5000));
    await streamExpenseStakes(expensestakeClient, expenseId, abortController, expensestakesStore);
}

export async function streamExchangeRate(
	currencyClient: PromiseClient<typeof CurrencyService>,
	srcCurrencyID: string,
	destCurrencyID: string,
	abortController: AbortController,
    exchangeRate: Writable<number | undefined>
): Promise<boolean> {
	try {
		for await (const erRes of currencyClient.streamExchangeRate({
            sourceCurrencyId: srcCurrencyID,
            destinationCurrencyId: destCurrencyID
        }, {signal: abortController.signal})) {
			if (erRes.update.case === 'stillAlive') {
				continue;
			}
            exchangeRate.set(erRes.update.value);
		}
    } catch (e) {
        if (e instanceof ConnectError) {
            if (e.code === Code.Canceled) {
                console.log(`Intentionally aborted exchange rate stream with source currency ${srcCurrencyID} and destination currency ${destCurrencyID}`);
                return true;
            } else if (e.code === Code.DataLoss) {
                console.log(`A currency of exchange rate with source currency ${srcCurrencyID} and destination currency ${destCurrencyID} no longer exists`);
                return false;
            }
        }
        console.error('An error occurred trying to stream exchange rate.', e);
    }
    console.log(`Ended exchange rate stream. Starting new one in 5 seconds.`);
    await new Promise(resolve => setTimeout(resolve, 5000));
    return await streamExchangeRate(currencyClient, srcCurrencyID, destCurrencyID, abortController, exchangeRate);
}

export function summariseStakes(expensestakes: ExpenseStake[]): string {
    const mainValues: number[] = [];
    const fractionalValues: number[] = [];
    for (let stake of expensestakes) {
        mainValues.push(stake.mainValue);
        if (stake.fractionalValue) {
            fractionalValues.push(stake.fractionalValue);
        }
    }
    let mainSummary = mainValues.reduce((partialSum, a) => partialSum + a, 0);
    const fractionalSummary = fractionalValues.reduce((partialSum, a) => partialSum + a, 0)
    mainSummary += Math.floor(fractionalSummary / 100);
    const fractionalRemainder = fractionalSummary % 100;
    return marshalMainAndFractionalValues(mainSummary, fractionalRemainder);
}

export function stakeSumInCurrency(
    exchangeRate: number,
    stakeSum: string
): number {
    const sum = parseFloat(stakeSum);
    return sum * exchangeRate;
}

export function marshalExpenseStakeValue(expensestake: ExpenseStake): string {
    return marshalMainAndFractionalValues(expensestake.mainValue, expensestake.fractionalValue)
}

function marshalMainAndFractionalValues(main: number, fractional?: number): string {
    let fractionalValue: string;
    if (fractional) {
        fractionalValue = fractional.toString();
        if (fractionalValue.length === 1) {
            fractionalValue = `0${fractionalValue}`;
        }
    } else {
        fractionalValue = '00';
    }
    return `${main}.${fractionalValue}`;
}