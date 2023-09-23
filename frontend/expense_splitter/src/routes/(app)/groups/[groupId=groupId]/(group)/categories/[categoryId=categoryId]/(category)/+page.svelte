<script lang="ts">
	import { createPromiseClient } from "@bufbuild/connect";
	import { onDestroy, onMount } from "svelte";
	import { writable } from "svelte/store";
	import type { Category } from "../../../../../../../../../../../gen/lib/ts/common/category/v1/category_pb";
	import { CategoryService } from "../../../../../../../../../../../gen/lib/ts/service/category/v1/service_connect";
	import { streamCategory } from "../../utils";
	import type { PageData } from "./$types";

	export let data: PageData;

	const categoryClient = createPromiseClient(CategoryService, data.grpcWebTransport);
	let category = writable(undefined as Category | undefined);
	const abortController = new AbortController();
    let editMode = false;

	const editedCategory = writable({
		name: ''
	});

	onDestroy(() => {
		abortController.abort();
	});

	onMount(async () => {
		const res = await streamCategory(categoryClient, data.categoryId, abortController, category);
        if (!res) {
            console.error('Category no longer exists');
        }
	});

	async function updateCategory() {
		try {
			const res = await categoryClient.updateCategory({
				id: data.categoryId,
				updateFields: [
					{
						updateOption: {
							case: 'name',
							value: $editedCategory.name
						}
					}
				]
			});
			console.log('Updated category', res.category);
            editMode = false;
		} catch (e) {
			console.error('An error occurred trying to update category', e);
		}
	}

    function startEdit() {
        if (!$category) {
            return;
        }
        editedCategory.set({
            name: $category.name
        })
        editMode = true;
    }

    function stopEdit() {
        editMode = false;
    }
</script>

<h2>Your category with ID {data.categoryId}</h2>
<table>
	<thead>
		<th>Name</th>
		<th>Action</th>
	</thead>
	<tbody>
		{#if $category}
            {#if editMode}
                <tr>
                    <td><input type="text" placeholder="Category name" bind:value={$editedCategory.name}/></td>
                    <td>
                        <button on:click={updateCategory}>Update category</button>
                        <button on:click={stopEdit}>Cancel</button>
                    </td>
                </tr>
            {:else}
                <tr>
                    <td>{$category.name}</td>
                    <td><button on:click={startEdit}>Update category</button></td>
                </tr>
            {/if}
		{:else}
			<tr>Loading category...</tr>
		{/if}
	</tbody>
</table>