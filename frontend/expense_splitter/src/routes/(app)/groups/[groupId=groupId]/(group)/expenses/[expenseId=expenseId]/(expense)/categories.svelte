<script lang="ts">
	import { ExpenseCategoryRelationService } from "../../../../../../../../../../../gen/lib/ts/service/expensecategoryrelation/v1/service_connect";
	import type { Category } from "../../../../../../../../../../../gen/lib/ts/common/category/v1/category_pb";
	import { CategoryService } from "../../../../../../../../../../../gen/lib/ts/service/category/v1/service_connect";
	import { createPromiseClient, type Transport } from "@bufbuild/connect";
	import { writable, type Writable } from "svelte/store";
	import { streamCategoriesForExpense as streamCategoryIdsForExpense } from "../../utils";
	import { onDestroy, onMount } from "svelte";
	import type { Expense } from "../../../../../../../../../../../gen/lib/ts/common/expense/v1/expense_pb";
	import { streamCategories } from "../../../categories/utils";

    export let transport: Transport;
    export let expense: Expense;

	const expenseCategoryRelationClient = createPromiseClient(ExpenseCategoryRelationService, transport);
	let relatedCategoryIds: Writable<string[] | undefined> = writable();
    const relatedCategoriesAbortController = new AbortController();

	const categoryClient = createPromiseClient(CategoryService, transport);
	let categories: Writable<
		Map<string, {category?: Category, abortController: AbortController}> | undefined
	> = writable();
	const categoriesAbortController = new AbortController();

    const newCategoryRelation = {
        categoryId: '',
    };

    onDestroy(() => {
        relatedCategoriesAbortController.abort();
        categoriesAbortController.abort();
    });

    onMount(() => {
        streamCategories(categoryClient, expense.groupId, categoriesAbortController, categories);
        streamCategoryIdsForExpense(expenseCategoryRelationClient, expense.id, relatedCategoriesAbortController, relatedCategoryIds);
    });

    async function createCategoryRelation() {
        try {
            const res = await expenseCategoryRelationClient.createExpenseCategoryRelation({
                expenseId: expense.id,
                categoryId: newCategoryRelation.categoryId
            });
            console.log('Created category relation');

            newCategoryRelation.categoryId = '';
        } catch (e) {
            console.error(`An error occurred trying to create category relation between expense ${expense.id} and category ${newCategoryRelation.categoryId}`, e);
        }
    }

    function deleteCategory(categoryId: string) {
        return async () => {
            try {
                await categoryClient.deleteCategory({id: categoryId});
                console.log('Deleted category');
            } catch (e) {
                console.error(`An error occurred trying to delete category ${categoryId}`, e);
            }
        }
    }
</script>

<h2>Your categories stakes in expense {expense.id}</h2>

<table>
	<thead>
		<th>ID</th>
		<th>Name</th>
		<th>Action</th>
	</thead>
	<tbody>
		{#if $categories}
			{#each [...$categories] as [cID, category]}
                {#if $relatedCategoryIds?.includes(cID)}
                    {#if category.category}
                        <tr>
                            <td>{cID}</td>
                            <td>{category.category.name}</td>
                            <td><button on:click|stopPropagation={deleteCategory(cID)}>Delete</button></td>
                        </tr>
                    {:else}
                        <tr>Loading category with ID {cID}...</tr>
                    {/if}
                {/if}
			{/each}
			<tr>
				<td></td>
				<td>
                    <select bind:value={newCategoryRelation.categoryId}>
                        {#each [...$categories] as [cID, category]}
                            <option value={cID}>{category.category?.name}</option>
                        {/each}
                    </select>
				</td>
				<td><button on:click={createCategoryRelation}>Create category relation</button></td>
			</tr>
		{:else}
			<tr>Loading categories...</tr>
		{/if}
	</tbody>
</table>