<script>
    import { createEventDispatcher } from 'svelte';
    import { callAPI }  from '../../common/API.js';
    import Modal from '../../common/Modal.svelte';
    import Delete from './Delete.svelte'
    import Reset from './Reset.svelte'
    import Restore from './Restore.svelte'
    
    const dispatch = createEventDispatcher();
    
    export let repo = {};
    let loading = true;
    let showModal = false;
    let snapshots = [];

    function getSnapshots() {
        callAPI('/repo/'+repo.id+'/snapshot', {
            method: 'GET'
        })
        .then(data => {
            loading = false;
            snapshots = data.snapshots;
        })
    }

    function toggleModal() {
        showModal = !showModal;

        if (showModal) {
            getSnapshots();
        }
    }

    function refresh(e) {
        loading = true;
        getSnapshots();
    }
</script>
<style>
    
</style>
<button class="btn btn-link float-end" on:click={toggleModal}>
    <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#search" /></svg>
</button>
{#if showModal}
    <Modal on:close={toggleModal} fit={true}>
        <h2 slot="header">
			Snapshots
		</h2>
        
        <table class="table">
            <thead>
                <tr>
                    <th scope="col">#</th>
                    <th scope="col">Time</th>
                    <th scope="col">Host</th>
                    <th scope="col">Tags</th>
                    <th scope="col">Paths</th>
                    <th scope="col" style='text-align:right;'>Actions</th>
                </tr>
            </thead>
            <tbody>
                {#if loading}
                <div class="spinner-grow position-absolute top-50 start-50 translate-middle" role="status">
                    <span class="visually-hidden">Loading...</span>
                </div>
                {/if}
                {#if !loading}
                    {#if snapshots.length == 0}
                        <b>No snapshots found.</b>
                    {/if}
                    {#each snapshots as snapshot}
                        <tr>
                            <th scope="row">{snapshot.id.substring(0,8)}</th>
                            <td>{snapshot.time}</td>
                            <td>{snapshot.hostname}</td>
                            <td>{snapshot.tags}</td>
                            <td>{snapshot.paths}</td>
                            <td>
                                <Delete {repo} {snapshot} on:refresh={refresh} on:click={toggleModal} />
                                <Restore {repo} {snapshot} on:refresh={refresh} />
                            </td>
                        </tr>
                        {/each}
                        {/if}
                    </tbody>
                </table>
                
        <div slot="buttons" class="float-end" style="display: inline-block;">
            <button type="button" class="btn btn-secondary float-end mx-1" on:click={toggleModal}>Close</button>
            <Reset {repo} {snapshots} on:refresh={refresh} />
        </div>
	</Modal>
{/if}