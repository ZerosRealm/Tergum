<script>
    import { onMount} from 'svelte';
    import Modal from '../../../common/Modal.svelte';
    import { callAPI }  from '../../../common/API.js';

    import Folder from './Folder.svelte';

    export let repo = {};
    export let snapshot = {};

    onMount(async () => {
        // getNodes();
	});

    let loading = true;
    let showModal = false;

    let directories = [];

    function getNodes() {
        loading = true;
        callAPI('/repo/'+repo.id+'/snapshot/'+snapshot.id+'/list', {
            method: 'GET'
        })
        .then(data => {
            directories = data.directories;
            console.log(directories);

            loading = false;
        })
    }

    function toggleModal() {
        showModal = !showModal;

        if (showModal) {
            getNodes();
        }
    }
</script>

<style>
    :global(.modal-full) {
        max-width: 100vw !important;
    }

    .folder {
        display: block;
    }

    pre {
        margin: 0px;
        overflow: initial;
        display: inline-block;
    }
</style>

<button class="btn btn-link" on:click={toggleModal}>
    <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#search"/></svg>
</button>

{#if showModal}
    <Modal extraClass="modal-full" on:close={toggleModal}>
        <h2 slot="header">
			Explorer - <pre>{snapshot.id.substring(0,8)}</pre>
		</h2>
        
        {#if loading}
        <div class="spinner-grow position-absolute top-50 start-50 translate-middle" role="status">
            <span class="visually-hidden">Loading...</span>
        </div>
        This can take awhile.
        {:else}
        
            {#each directories as dir}
                <div class="folder">
                    <Folder name={dir.name} files={dir.files} />
                </div>
            {/each}

        {/if}
        
        <div slot="buttons" class="float-end" style="display: inline-block;">
            <!-- <button type="button" class="btn btn-primary float-end" on:click={confirm} disabled={( agent == -1 )}>Restore</button> -->
            <button type="button" class="btn btn-secondary float-end mx-1" on:click={toggleModal}>Close</button>
        </div>
	</Modal>
{/if}