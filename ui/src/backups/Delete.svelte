<script>
    import { createEventDispatcher} from 'svelte';
    import Modal from '../common/Modal.svelte';
    import { callAPI }  from '../common/API.js';
    
	const dispatch = createEventDispatcher();
    
    export let backup = {};
    let showModal = false;

    function toggleModal() {
        showModal = !showModal;
    }

    function confirm() {
        callAPI('/backup/'+backup.id, {
            method: 'DELETE'
        })
        .then(() => {
            toggleModal();
            dispatch('refresh', {});
        })
    }
</script>
<style>

</style>
<button class="btn btn-danger float-end ms-1" on:click={toggleModal}>
    <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#trash"/></svg>
</button>
{#if showModal}
    <Modal on:close={toggleModal}>
        <h2 slot="header">
			Delete backup
		</h2>
        
        Do you want to delete backup <code>{backup.source}</code>?
        
        <div slot="buttons" class="float-end" style="display: inline-block;">
            <button type="button" class="btn btn-danger float-end" on:click={confirm}>Delete</button>
            <button type="button" class="btn btn-secondary float-end mx-1" on:click={toggleModal}>Close</button>
        </div>
	</Modal>
{/if}