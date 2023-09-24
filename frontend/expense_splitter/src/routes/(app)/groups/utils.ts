import { Code, ConnectError, type PromiseClient } from "@bufbuild/connect";
import { get, writable, type Writable } from "svelte/store";
import type { Group } from "../../../../../../gen/lib/ts/common/group/v1/group_pb";
import type { GroupService } from "../../../../../../gen/lib/ts/service/group/v1/service_connect";
import type { CurrencyService } from "../../../../../../gen/lib/ts/service/currency/v1/service_connect";
import type { Currency } from "../../../../../../gen/lib/ts/common/currency/v1/currency_pb";
import type { ExpenseStakeService } from "../../../../../../gen/lib/ts/service/expensestake/v1/service_connect";
import type { ExpenseStake } from "../../../../../../gen/lib/ts/common/expensestake/v1/expensestake_pb";

export async function streamGroup(
	groupClient: PromiseClient<typeof GroupService>,
	groupID: string,
	abortController: AbortController,
    group: Writable<Group | undefined>
): Promise<boolean> {
	try {
		for await (const pRes of groupClient.streamGroup({id: groupID}, {signal: abortController.signal})) {
			if (pRes.update.case === 'stillAlive') {
				continue;
			}
            group.set(pRes.update.value);
		}
    } catch (e) {
        if (e instanceof ConnectError) {
            if (e.code === Code.Canceled) {
                console.log(`Intentionally aborted group ${groupID} stream`);
                return true;
            } else if (e.code === Code.DataLoss) {
                console.log(`Group with ID ${groupID} no longer exists`);
                return false;
            }
        }
        console.error('An error occurred trying to stream group.', e);
    }
    console.log(`Ended group ${groupID} stream. Starting new one in 5 seconds.`);
    await new Promise(resolve => setTimeout(resolve, 5000));
    return await streamGroup(groupClient, groupID, abortController, group);
}

export async function streamGroups(
    groupClient: PromiseClient<typeof GroupService>,
    abortController: AbortController,
    groupsStore: Writable<Map<string, {group?: Group, abortController: AbortController}> | undefined>
) {
    try {
        for await (const cIDsRes of groupClient.streamGroupIds({}, {signal: abortController.signal})) {
            if (cIDsRes.update.case === 'stillAlive') {
                continue;
            }
            const groupIDs = cIDsRes.update.value!.ids;
            let groups = get(groupsStore);
            if (groups === undefined) {
                groups = new Map();
            }
            for (const cID of groups.keys()) {
                if (groupIDs.includes(cID)) {
                    // remove element from items that are to be processed because it already exists
                    groupIDs.splice(groupIDs.indexOf(cID), 1);
                } else {
                    // remove groups that are not present any more
                    groups!.get(cID)!.abortController.abort();
                    groups!.delete(cID);
                }
            }
            for (const pID of groupIDs) {
                const abortController = new AbortController();
                groups.set(pID, {
                    abortController: abortController
                });
                const group: Writable<Group | undefined> = writable();
                group.subscribe((e) => {
                    groupsStore.set(groups?.set(pID, {
                        abortController: abortController,
                        group: e
                    }));
                });
                streamGroup(groupClient, pID, abortController, group);
            }
            groupsStore.set(groups);
        }
    } catch (e) {
        if (e instanceof ConnectError && e.code === Code.Canceled) {
            console.log('Intentionally aborted groups stream');
            return;
        }
        console.error('An error occurred trying to stream groups', e);
    } finally {
        const groups = get(groupsStore);
        if (groups) {
            for (let [_, group] of groups) {
                group.abortController.abort();
            }
        }
        groupsStore.set(undefined);
    }
    console.log(`Ended groups stream. Starting new one in 5 seconds.`);
    await new Promise(resolve => setTimeout(resolve, 5000));
    await streamGroups(groupClient, abortController, groupsStore);
}

export async function streamCurrencies(
    currencyClient: PromiseClient<typeof CurrencyService>,
    abortController: AbortController,
    currenciesStore: Writable<Map<string, Currency> | undefined>
) {
    try {
        for await (const cIDsRes of currencyClient.streamCurrencies({}, {signal: abortController.signal})) {
            if (cIDsRes.update.case === 'stillAlive') {
                continue;
            }
            const currencies = new Map<string, Currency>();
            for (let currency of cIDsRes.update.value!.currencies) {
                currencies.set(currency.id, currency);
            }
            currenciesStore.set(currencies);
        }
    } catch (e) {
        if (e instanceof ConnectError && e.code === Code.Canceled) {
            console.log('Intentionally aborted currencies stream');
            return;
        }
        console.error('An error occurred trying to stream currencies', e);
    } finally {
        currenciesStore.set(undefined);
    }
    console.log(`Ended currencies stream. Starting new one in 5 seconds.`);
    await new Promise(resolve => setTimeout(resolve, 5000));
    await streamCurrencies(currencyClient, abortController, currenciesStore);
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

export async function streamExpenseStakesInGroup(
    expensestakeClient: PromiseClient<typeof ExpenseStakeService>,
    groupId: string,
    abortController: AbortController,
    expensestakesStore: Writable<Map<string, {expensestake?: ExpenseStake, abortController: AbortController}> | undefined>
) {
    try {
        for await (const cIDsRes of expensestakeClient.streamExpenseStakeIdsInGroup({
            groupId: groupId
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
    await streamExpenseStakesInGroup(expensestakeClient, groupId, abortController, expensestakesStore);
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