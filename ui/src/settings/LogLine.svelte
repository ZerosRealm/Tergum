<script>
    export let log = {};
    let open = false;
</script>
<style>
    td:hover {
        cursor: pointer;
    }
</style>

<td class="log-line" on:click={() => { open = !open }}>
    <b>{log.msg}</b>

    {#if open}
        <div class="data">
            {#each Object.entries(log) as [key, val]}
                <div>
                    <b>{key}</b>
                    <!-- TODO: Possible XSS by injecting HTML into the file logs. -->
                    <pre>{@html val}</pre>
                </div>
            {/each}
        </div>
    {/if}
</td>