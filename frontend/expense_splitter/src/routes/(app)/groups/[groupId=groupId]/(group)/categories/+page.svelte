<script lang="ts">
	import { createPromiseClient } from "@bufbuild/connect";
	import type { PageData } from "./$types";
	import { CategoryService } from "../../../../../../../../../gen/lib/ts/service/category/v1/service_connect";
	import type { Category } from "../../../../../../../../../gen/lib/ts/common/category/v1/category_pb";
	import { onDestroy, onMount } from "svelte";
	import { writable, type Writable } from "svelte/store";
	import { goto } from "$app/navigation";
	import { streamCategories } from "./utils";

	export let data: PageData;

	const categoryClient = createPromiseClient(CategoryService, data.grpcWebTransport);
	let categories: Writable<
		Map<string, {category?: Category, abortController: AbortController}> | undefined
	> = writable();

	const newCategory = writable({
		name: '',
	});

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
                name: $newCategory.name
			});
			console.log('Created category', res.id);

            newCategory.set({name: ''});
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

<h2>Your categories in group {data.groupId}</h2>

<table>
	<thead>
		<th>ID</th>
		<th>Name</th>
		<th>Action</th>
	</thead>
	<tbody>
		{#if $categories}
			{#each [...$categories] as [pID, category]}
				{#if category.category}
					<tr on:click={openCategory(pID)}>
						<td>{pID}</td>
						<td>{category.category?.name}</td>
						<td><button on:click|stopPropagation={deleteCategory(pID)}>Delete</button></td>
					</tr>
				{:else}
					<tr>Loading category with ID {pID}...</tr>
				{/if}
			{/each}
		{:else}
			<tr>Loading categories...</tr>
		{/if}
		<tr>
			<td></td>
			<td><input type="text" placeholder="Category name" bind:value={$newCategory.name}/></td>
			<td><button on:click={createCategory}>Create category</button></td>
		</tr>
	</tbody>
</table>