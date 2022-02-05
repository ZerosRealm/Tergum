<script>
    import Modal from '../../common/Modal.svelte';

    export let job = {};
    
    let showModal = false;

    function toggleModal() {
        showModal = !showModal;
    }

    const capitalize = (s) => {
        if (typeof s !== 'string') return ''
        return s.charAt(0).toUpperCase() + s.slice(1)
    }
</script>
<style>

</style>
<button class="btn btn-link float-end" type="button" on:click={toggleModal}>
    <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#search"/></svg>
</button>
{#if showModal}
    <Modal on:close={toggleModal}>
		<h2 slot="header">
			Job information
		</h2>

        {#each Object.entries(job) as [key, val]}
            {#if key != "message_type"}
                {#if key == "percent_done"}
                    <label for="{key}" class="form-label mt-3">{capitalize(key).replaceAll("_", " ")}</label>
                    <input type="text" class="form-control" value="{Math.floor(val*100)}%" disabled>
                {:else}
                    <label for="{key}" class="form-label mt-3">{capitalize(key).replaceAll("_", " ")}</label>
                    <input type="text" class="form-control" value="{val}" disabled>
                {/if}
            {/if}
        {/each}

        <div slot="buttons" class="float-end" style="display: inline-block;">
            <button type="button" class="btn btn-secondary float-end mx-1" on:click={toggleModal}>Close</button>
        </div>
	</Modal>
{/if}