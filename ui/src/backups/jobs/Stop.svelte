<script>
    import { createEventDispatcher} from 'svelte';
    import Modal from '../../common/Modal.svelte';
    import { callAPI }  from '../../common/API.js';

    const dispatch = createEventDispatcher();

    export let job = {};
    
    let showModal = false;

    function toggleModal() {
        showModal = !showModal;
    }

    function confirm() {
        callAPI('/job/'+job.id, {
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

{#if job.message_type == "status"}
    <button class="btn btn-link float-end text-danger" type="button" on:click={toggleModal}>
        <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#x-circle-fill"/></svg>
    </button>
{/if}
{#if showModal}
    <Modal on:close={toggleModal}>
		<h2 slot="header">
			Stop job
		</h2>

        Do you want to stop this running job?<br/>
        This only affects this job, not the other jobs in the queue.

        <div slot="buttons" class="float-end" style="display: inline-block;">
            <button type="button" class="btn btn-danger float-end" on:click={confirm}>Stop</button>
            <button type="button" class="btn btn-secondary float-end mx-1" on:click={toggleModal}>Close</button>
        </div>
	</Modal>
{/if}