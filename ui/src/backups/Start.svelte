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
        callAPI('/job', {
            method: 'POST',
            body: JSON.stringify({
                backup: backup.id,
            })
        })
        .then(() => {
            toggleModal();
            dispatch('refresh', {});
        })
    }
</script>
<style>
.btn.btn-link {
    color: #3B4252 !important;
    
}

.btn.btn-link:hover {
    color: #fff !important;
    background-color: #3B4252 !important;
}

</style>

<button class="btn btn-link float-end text-primary" type="button" on:click={toggleModal}>
    <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#caret-right-fill"/></svg>
</button>
{#if showModal}
    <Modal on:close={toggleModal}>
		<h2 slot="header">
			Start backup
		</h2>

        Do you want to start this backup up?

        <div slot="buttons" class="float-end" style="display: inline-block;">
            <button type="button" class="btn btn-primary float-end" on:click={confirm}>Start</button>
            <button type="button" class="btn btn-secondary float-end mx-1" on:click={toggleModal}>Close</button>
        </div>
	</Modal>
{/if}