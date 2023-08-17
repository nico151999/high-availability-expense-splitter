import { Code, ConnectError, type PromiseClient } from "@bufbuild/connect";
import type { Group } from "../../../../../../gen/lib/ts/common/group/v1/group_pb";
import type { GroupService } from "../../../../../../gen/lib/ts/service/group/v1/service_connect";

export async function streamGroup(
	groupClient: PromiseClient<typeof GroupService>,
	groupID: string,
	abortController: AbortController,
    onGroupUpdate: (group: Group) => void
) {
	try {
		for await (const gRes of groupClient.streamGroup({groupId: groupID}, {signal: abortController.signal})) {
			if (gRes.update.case === 'stillAlive') {
				continue;
			}
			onGroupUpdate(gRes.update.value!);
		}
    } catch (e) {
        if (e instanceof ConnectError && e.code === Code.Canceled) {
            console.log(`Intentionally aborted group ${groupID} stream`);
            return;
        } else {
            console.error('An error occurred trying to stream group. Trying anew in 5 seconds.', e);
        }
    }
    console.log(`Ended group ${groupID} stream`);
    setTimeout(() => streamGroup(groupClient, groupID, abortController, onGroupUpdate), 5000);
}