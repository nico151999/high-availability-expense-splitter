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
	import LayoutGrid, {Cell as LayoutCell} from "@smui/layout-grid";
	import DataTable, { Body, Cell, Head, Row } from "@smui/data-table";
	import IconButton from "@smui/icon-button";
	import LinearProgress from "@smui/linear-progress";
	import { Separator } from "@smui/list";
	import Select, {Option} from "@smui/select";
	import Button, { Label } from "@smui/button";

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

    function deleteCategoryRelation(categoryId: string) {
        return async () => {
            try {
                await expenseCategoryRelationClient.deleteExpenseCategoryRelation({categoryId: categoryId, expenseId: expense.id});
                console.log('Deleted category relation');
            } catch (e) {
                console.error(`An error occurred trying to delete category relation to ${categoryId}`, e);
            }
        }
    }
</script>

<LayoutGrid>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<h2>Categories</h2>
		
		<DataTable table$aria-label="Category relation list" style="width: 100%">
			<Head>
				<Row>
					<Cell>Name</Cell>
					<Cell>Action</Cell>
				</Row>
			</Head>
			<Body>
				{#if $categories && $relatedCategoryIds}
					{#each [...$relatedCategoryIds] as cID}
                        {@const category = $categories.get(cID)}
                        {#if category?.category}
                            <Row>
                                <Cell>{category.category.name}</Cell>
                                <Cell>
                                    <IconButton
                                        on:click$stopPropagation={deleteCategoryRelation(cID)}
                                        class="material-icons"
                                        aria-label="Delete category">delete</IconButton>
                                </Cell>
                            </Row>
                        {/if}
					{/each}
				{/if}
			</Body>
		
			<LinearProgress
				indeterminate
				closed={!!$categories && !!$relatedCategoryIds}
				aria-label="Categories are being loaded..."
				slot="progress"
			/>
		</DataTable>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<h4>New expense category relation</h4>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		{#if $categories}
			<Select variant="outlined" bind:value={newCategoryRelation.categoryId} label="Expense category relation" style="width: 100%">
				{#each [...$categories] as [cID, category]}
                    {#if category.category && $relatedCategoryIds && !$relatedCategoryIds.includes(cID)}
                        <Option value={cID}>
                            {category.category.name}
                        </Option>
                    {/if}
				{/each}
			</Select>
		{/if}
		<LinearProgress
			indeterminate
			closed={!!$categories}
			aria-label="Currencies are being loaded..."/>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }} style="display: flex; justify-content: flex-end">
		<Button on:click={createCategoryRelation} touch variant="outlined">
			<Label>Create category relation</Label>
		</Button>
	</LayoutCell>
</LayoutGrid>