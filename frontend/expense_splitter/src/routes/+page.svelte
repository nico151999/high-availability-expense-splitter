<script lang="ts">
	import { goto } from "$app/navigation";
	import { t } from "$lib/localization";
	import Button, { Label } from "@smui/button";
	import LayoutGrid, {Cell} from "@smui/layout-grid";

    function gotoGroups() {
        goto('./groups');
    }
</script>

<LayoutGrid>
	<Cell span={12}>
        <h1>Welcome</h1>
        <Button on:click={gotoGroups} touch variant="outlined">
            <Label>{$t("root.openGroups")}</Label>
        </Button>
    </Cell>
</LayoutGrid>