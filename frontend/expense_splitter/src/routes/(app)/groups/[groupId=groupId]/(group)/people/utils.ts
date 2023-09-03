import { Code, ConnectError, type PromiseClient } from "@bufbuild/connect";
import { get, writable, type Writable } from "svelte/store";
import type { Person } from "../../../../../../../../../gen/lib/ts/common/person/v1/person_pb";
import type { PersonService } from "../../../../../../../../../gen/lib/ts/service/person/v1/service_connect";

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
    }
    console.log(`Ended people stream. Starting new one in 5 seconds.`);
    await new Promise(resolve => setTimeout(resolve, 5000));
    await streamPeople(personClient, groupId, abortController, peopleStore);
}