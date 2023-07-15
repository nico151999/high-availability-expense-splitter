<script lang="ts">
	import { goto } from "$app/navigation";

    function goLevelUp() {
        goto('./');
    }
</script>

<button on:click={goLevelUp}>Navigate level up</button>
<slot></slot>