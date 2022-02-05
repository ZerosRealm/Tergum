<script>
    import { createEventDispatcher} from 'svelte';
    import Modal from '../../common/Modal.svelte';
    import { callAPI }  from '../../common/API.js';
    
	const dispatch = createEventDispatcher();
    
    export let repo = {};
    export let snapshots = [];
    let showModal = false;
    let loading = false;

    function toggleModal() {
        showModal = !showModal;
    }

    function confirm() {
        loading = true;
        queue = snapshots.length;
        snapshots.forEach(snapshot => {
            callAPI('/repo/'+repo.id+'/snapshot/'+snapshot.id, {
                method: 'DELETE'
            })
            .then(() => {
                queue--;

                if (queue === 0) {
                    loading = false;
                    dispatch('refresh', {});
                }
            })
        });
        toggleModal();
    }
</script>
<style>
    .spinner-grow {
        width: 1rem;
        height: 1rem;
    }
</style>
<button class="btn btn-danger float-end ms-1" disabled={(snapshots.length==0 || loading)} on:click={toggleModal}>
    {#if loading}
        <div class="spinner-grow" role="status">
            <span class="visually-hidden">Loading...</span>
        </div>
    {/if}
    Reset <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#trash"/></svg>
</button>
{#if showModal}
    <Modal on:close={toggleModal}>
        <h2 slot="header">
			Reset repository
		</h2>
        
        Do you want to <b>reset</b> the repository <b>{repo.Name}</b>?<br/>
        This will delete <b>ALL SNAPSHOTS</b> in the repository.
        
        <div slot="buttons" class="float-end" style="display: inline-block;">
            <button type="button" class="btn btn-danger float-end" on:click={confirm}>Reset</button>
            <button type="button" class="btn btn-secondary float-end mx-1" on:click={toggleModal}>Close</button>
        </div>
	</Modal>
{/if}