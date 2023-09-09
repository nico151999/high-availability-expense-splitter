import { Code, ConnectError, type PromiseClient } from "@bufbuild/connect";
import { get, writable, type Writable } from "svelte/store";
import type { ExpenseStakeService } from "../../../../../../../../../../../gen/lib/ts/service/expensestake/v1/service_connect";
import type { ExpenseStake } from "../../../../../../../../../../../gen/lib/ts/common/expensestake/v1/expensestake_pb";

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
        expensestakes?.clear();
    }
    console.log(`Ended expensestakes stream. Starting new one in 5 seconds.`);
    await new Promise(resolve => setTimeout(resolve, 5000));
    await streamExpenseStakes(expensestakeClient, expenseId, abortController, expensestakesStore);
}