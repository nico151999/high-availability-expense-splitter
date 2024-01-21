<script lang="ts">
	import axios from 'axios';
	import TopAppBar, { Row, Section, Title, AutoAdjust } from '@smui/top-app-bar';
	import Drawer, {
		AppContent,
		Content,
		Header,
		Title as DrawerTitle,
		Subtitle,
		Scrim
	} from '@smui/drawer';
	import List, { Item, Text, Graphic, Separator, Subheader } from '@smui/list';
	import IconButton from '@smui/icon-button';
	import { t } from '$lib/localization';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';

	let topAppBar: TopAppBar;
	let drawerOpen = false;
</script>

<Drawer variant="modal" bind:open={drawerOpen}>
	<Header>
		<DrawerTitle>Placeholder title</DrawerTitle>
		<Subtitle>A placeholder for a drawer</Subtitle>
	</Header>
	<Content>
		<List>
			<Item>
				<Graphic class="material-icons" aria-hidden="true">inbox</Graphic>
				<Text>Placeholder Item 1</Text>
			</Item>
			<Item>
				<Graphic class="material-icons" aria-hidden="true">star</Graphic>
				<Text>Placeholder Item 2</Text>
			</Item>
			<Separator />
			<Subheader tag="h6">Further items</Subheader>
			<Item>
				<Graphic class="material-icons" aria-hidden="true">inbox</Graphic>
				<Text>Placeholder Item 3</Text>
			</Item>
			<Item>
				<Graphic class="material-icons" aria-hidden="true">star</Graphic>
				<Text>Placeholder Item 4</Text>
			</Item>
		</List>
	</Content>
</Drawer>
<Scrim />
<AppContent>
	<header>
		<TopAppBar bind:this={topAppBar} variant="standard">
			<Row>
				<Section>
					<IconButton
						on:click={() => (drawerOpen = !drawerOpen)}
						class="material-icons"
						aria-label={$t('global.toggleDrawer')}>menu</IconButton>
					<Title>{$t('global.title')}</Title>
				</Section>
				<Section align="end" toolbar>
					{#if $page.url.pathname !== '/'}
						<IconButton on:click={() => goto('./')} class="material-icons" aria-label={$t("global.navigateLevelUp")}>arrow_upward</IconButton>
					{/if}
				</Section>
			</Row>
		</TopAppBar>
	</header>
	<main>
		<AutoAdjust {topAppBar}>
			<slot />
		</AutoAdjust>
	</main>
</AppContent>
