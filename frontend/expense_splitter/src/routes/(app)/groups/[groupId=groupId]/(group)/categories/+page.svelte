<script lang="ts">
	import { createPromiseClient } from "@bufbuild/connect";
	import type { PageData } from "./$types";
	import { CategoryService } from "../../../../../../../../../gen/lib/ts/service/category/v1/service_connect";
	import type { Category } from "../../../../../../../../../gen/lib/ts/common/category/v1/category_pb";
	import { onDestroy, onMount } from "svelte";
	import { writable, type Writable } from "svelte/store";
	import { goto } from "$app/navigation";
	import { streamCategories } from "./utils";
	import LayoutGrid, {Cell as LayoutCell} from "@smui/layout-grid";
	import DataTable, { Body, Cell, Head, Row } from "@smui/data-table";
	import { t } from "$lib/localization";
	import IconButton from "@smui/icon-button";
	import LinearProgress from "@smui/linear-progress";
	import Textfield from "@smui/textfield";
	import HelperText from "@smui/textfield/helper-text";
	import Button, { Label } from "@smui/button";
	import { Separator } from "@smui/list";

	export let data: PageData;

	const categoryClient = createPromiseClient(CategoryService, data.grpcWebTransport);
	let categories: Writable<
		Map<string, {category?: Category, abortController: AbortController}> | undefined
	> = writable();

	const newCategory = {
		name: '',
	};

	const abortController = new AbortController();
	onDestroy(() => {
		abortController.abort();
	});

	onMount(
		() => streamCategories(categoryClient, data.groupId, abortController, categories)
	);

	async function createCategory() {
		try {
			const res = await categoryClient.createCategory({
                groupId: data.groupId,
                name: newCategory.name
			});
			console.log('Created category', res.id);

            newCategory.name = '';
		} catch (e) {
			console.error(`An error occurred trying to create category in group ${data.groupId}`, e);
		}
	}

	function deleteCategory(categoryId: string) {
		return async () => {
			try {
				await categoryClient.deleteCategory({id: categoryId});
				console.log('Deleted category');
			} catch (e) {
				console.error(`An error occurred trying to delete category ${categoryId} in group ${data.groupId}`, e);
			}
		}
	}

	function openCategory(categoryId: string) {
		return () => {
			goto(`./categories/${categoryId}`);
		}
	}
</script>

<LayoutGrid>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<h2>Categories</h2>
		
		<DataTable table$aria-label="Category list" style="width: 100%">
			<Head>
				<Row>
					<Cell>ID</Cell>
					<Cell>Name</Cell>
					<Cell>Action</Cell>
				</Row>
			</Head>
			<Body>
				{#if $categories}
					{#each [...$categories] as [cID, category]}
						{#if category.category}
							<Row on:click={openCategory(cID)}>
								<Cell>{cID}</Cell>
								<Cell>{category.category.name}</Cell>
								<Cell>
									<IconButton
										on:click$stopPropagation={deleteCategory(cID)}
										class="material-icons"
										aria-label="Delete category">delete</IconButton>
								</Cell>
							</Row>
						{:else}
							<Row>
								<LinearProgress
									indeterminate
									aria-label={$t('categories.loadingCategoryWithId', { categoryId: cID })} />
							</Row>
						{/if}
					{/each}
				{/if}
			</Body>
		
			<LinearProgress
				indeterminate
				closed={!!$categories}
				aria-label="Categories are being loaded..."
				slot="progress"
			/>
		</DataTable>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<Separator />
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<h4>New category</h4>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 6, tablet: 4, phone: 4 }}>
		<Textfield variant="outlined" bind:value={newCategory.name} label="Category name" style="width: 100%" helperLine$style="width: 100%">
			<HelperText slot="helper">The name of the category that is to be created</HelperText>
		</Textfield>
	</LayoutCell>
	<LayoutCell spanDevices={{ desktop: 6, tablet: 4, phone: 4 }} style="display: flex; justify-content: flex-end">
		<Button on:click={createCategory} touch variant="outlined">
			<Label>Create category</Label>
		</Button>
	</LayoutCell>
</LayoutGrid>