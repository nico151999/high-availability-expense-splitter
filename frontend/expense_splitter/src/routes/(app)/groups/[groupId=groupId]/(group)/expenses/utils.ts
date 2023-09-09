import { Code, ConnectError, type PromiseClient } from "@bufbuild/connect";
import { get, writable, type Writable } from "svelte/store";
import type { Expense } from "../../../../../../../../../gen/lib/ts/common/expense/v1/expense_pb";
import type { ExpenseService } from "../../../../../../../../../gen/lib/ts/service/expense/v1/service_connect";

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
        expenses?.clear();
    }
    console.log(`Ended expenses stream. Starting new one in 5 seconds.`);
    await new Promise(resolve => setTimeout(resolve, 5000));
    await streamExpenses(expenseClient, groupId, abortController, expensesStore);
}