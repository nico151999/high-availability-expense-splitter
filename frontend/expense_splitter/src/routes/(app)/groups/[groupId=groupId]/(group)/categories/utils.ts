import { Code, ConnectError, type PromiseClient } from "@bufbuild/connect";
import { get, writable, type Writable } from "svelte/store";
import type { Category } from "../../../../../../../../../gen/lib/ts/common/category/v1/category_pb";
import type { CategoryService } from "../../../../../../../../../gen/lib/ts/service/category/v1/service_connect";

export async function streamCategory(
	categoryClient: PromiseClient<typeof CategoryService>,
	categoryID: string,
	abortController: AbortController,
    category: Writable<Category | undefined>
): Promise<boolean> {
	try {
		for await (const pRes of categoryClient.streamCategory({id: categoryID}, {signal: abortController.signal})) {
			if (pRes.update.case === 'stillAlive') {
				continue;
			}
            category.set(pRes.update.value);
		}
    } catch (e) {
        if (e instanceof ConnectError) {
            if (e.code === Code.Canceled) {
                console.log(`Intentionally aborted category ${categoryID} stream`);
                return true;
            } else if (e.code === Code.DataLoss) {
                console.log(`Category with ID ${categoryID} no longer exists`);
                return false;
            }
        }
        console.error('An error occurred trying to stream category.', e);
    }
    console.log(`Ended category ${categoryID} stream. Starting new one in 5 seconds.`);
    await new Promise(resolve => setTimeout(resolve, 5000));
    return await streamCategory(categoryClient, categoryID, abortController, category);
}

export async function streamCategories(
    categoryClient: PromiseClient<typeof CategoryService>,
    groupId: string,
    abortController: AbortController,
    categoriesStore: Writable<Map<string, {category?: Category, abortController: AbortController}> | undefined>
) {
    try {
        for await (const cIDsRes of categoryClient.streamCategoryIdsInGroup({
            groupId: groupId
        }, {signal: abortController.signal})) {
            if (cIDsRes.update.case === 'stillAlive') {
                continue;
            }
            const categoryIDs = cIDsRes.update.value!.ids;
            let categories = get(categoriesStore);
            if (categories === undefined) {
                categories = new Map();
            }
            for (const cID of categories.keys()) {
                if (categoryIDs.includes(cID)) {
                    // remove element from items that are to be processed because it already exists
                    categoryIDs.splice(categoryIDs.indexOf(cID), 1);
                } else {
                    // remove categories that are not present any more
                    categories!.get(cID)!.abortController.abort();
                    categories!.delete(cID);
                }
            }
            for (const pID of categoryIDs) {
                const abortController = new AbortController();
                categories.set(pID, {
                    abortController: abortController
                });
                const category: Writable<Category | undefined> = writable();
                category.subscribe((e) => {
                    categoriesStore.set(categories?.set(pID, {
                        abortController: abortController,
                        category: e
                    }));
                });
                streamCategory(categoryClient, pID, abortController, category);
            }
            categoriesStore.set(categories);
        }
    } catch (e) {
        if (e instanceof ConnectError && e.code === Code.Canceled) {
            console.log('Intentionally aborted categories stream');
            return;
        }
        console.error('An error occurred trying to stream categories', e);
    } finally {
        const categories = get(categoriesStore);
        if (categories) {
            for (let [_, category] of categories) {
                category.abortController.abort();
            }
        }
        categoriesStore.set(undefined);
    }
    console.log(`Ended categories stream. Starting new one in 5 seconds.`);
    await new Promise(resolve => setTimeout(resolve, 5000));
    await streamCategories(categoryClient, groupId, abortController, categoriesStore);
}