<script lang="ts">
	import { goto } from "$app/navigation";

    function gotoGroups() {
        goto('./groups');
    }
</script>

<h2>My main page</h2>
<button on:click={gotoGroups}>Open Groups</button>