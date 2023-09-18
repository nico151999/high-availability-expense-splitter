import { Code, ConnectError, type PromiseClient } from "@bufbuild/connect";
import { get, writable, type Writable } from "svelte/store";
import type { Group } from "../../../../../../gen/lib/ts/common/group/v1/group_pb";
import type { GroupService } from "../../../../../../gen/lib/ts/service/group/v1/service_connect";
import type { CurrencyService } from "../../../../../../gen/lib/ts/service/currency/v1/service_connect";
import type { Currency } from "../../../../../../gen/lib/ts/common/currency/v1/currency_pb";

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