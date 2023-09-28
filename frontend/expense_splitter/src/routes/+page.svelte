<script lang="ts">
	import { goto } from "$app/navigation";
	import { t } from "$lib/localization";

    function gotoGroups() {
        goto('./groups');
    }
</script>

<h2>My main page</h2>
<button on:click={gotoGroups}>{$t("root.openGroups")}</button>