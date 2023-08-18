import { Code, ConnectError, type PromiseClient } from "@bufbuild/connect";
import type { Group } from "../../../../../../gen/lib/ts/common/group/v1/group_pb";
import type { GroupService } from "../../../../../../gen/lib/ts/service/group/v1/service_connect";

// Streams a group specified by the passed group ID until it is aborted.
// In case of a finalised stream (with or without error) a retry is performed after a delay.
// If the stream is intentionally stopped the function returns true. If the stream is
// stopped due to a no longer existing group the function returns false.
export async function streamGroup(
	groupClient: PromiseClient<typeof GroupService>,
	groupID: string,
	abortController: AbortController,
    onGroupUpdate: (group: Group) => void
): Promise<boolean> {
	try {
		for await (const gRes of groupClient.streamGroup({groupId: groupID}, {signal: abortController.signal})) {
			if (gRes.update.case === 'stillAlive') {
				continue;
			}
			onGroupUpdate(gRes.update.value!);
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
    return await streamGroup(groupClient, groupID, abortController, onGroupUpdate);
}