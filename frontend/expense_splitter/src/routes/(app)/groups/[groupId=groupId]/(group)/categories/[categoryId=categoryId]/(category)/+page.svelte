<script lang="ts">
	import { createPromiseClient } from "@bufbuild/connect";
	import { onDestroy, onMount } from "svelte";
	import { writable, type Writable } from "svelte/store";
	import type { Category } from "../../../../../../../../../../../gen/lib/ts/common/category/v1/category_pb";
	import { CategoryService } from "../../../../../../../../../../../gen/lib/ts/service/category/v1/service_connect";
	import { streamCategory } from "../../utils";
	import type { PageData } from "./$types";
	import LayoutGrid, {Cell as LayoutCell} from "@smui/layout-grid";
	import Textfield from "@smui/textfield";
	import Button, { Label } from "@smui/button";
	import LinearProgress from "@smui/linear-progress";

	export let data: PageData;

	const categoryClient = createPromiseClient(CategoryService, data.grpcWebTransport);
	let category: Writable<Category | undefined> = writable();
	const abortController = new AbortController();
    let editMode = false;

	const editedCategory = {
		name: ''
	};
	$: if (!editMode) {
        editedCategory.name = $category?.name ?? '';
	}

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
							value: editedCategory.name
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
        editedCategory.name = $category.name;
        editMode = true;
    }

    function stopEdit() {
        editMode = false;
    }
</script>

<LayoutGrid>
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<h2>Category</h2>
	</LayoutCell>
	{#if $category}
		<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
			<Textfield variant="outlined" disabled={!editMode} bind:value={editedCategory.name} label="Category name" style="width: 100%" />
		</LayoutCell>
		<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }} style="display: flex; justify-content: center">
			{#if editMode}
				<Button on:click={updateCategory} variant="outlined">
					<Label>Update category</Label>
				</Button>
				<Button on:click={stopEdit} variant="outlined">
					<Label>Cancel</Label>
				</Button>
			{:else}
				<Button on:click={startEdit} variant="outlined">
					<Label>Edit category</Label>
				</Button>
			{/if}
		</LayoutCell>
	{/if}
	<LayoutCell spanDevices={{ desktop: 12, tablet: 8, phone: 4 }}>
		<LinearProgress
			indeterminate
			closed={!!$category}
			aria-label="Category is being loaded..."/>
	</LayoutCell>
</LayoutGrid>