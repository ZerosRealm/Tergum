<script>
    import { createEventDispatcher, onMount } from 'svelte';
    import Modal from '../../common/Modal.svelte';
    import { callAPI }  from '../../common/API.js';
    
	const dispatch = createEventDispatcher();
    
    export let repo = {};
    export let snapshot = {};
    let showModal = false;
    let agent = -1;
    let destination = "";
    let exclude = "";
    let include = "";

    let agents = [];

    function getAgents() {
        callAPI('/agent', {
            method: 'GET'
        })
        .then(data => {
            agents = data.agents;
        })
    }

    function toggleModal() {
        showModal = !showModal;

        if (showModal) {
            getAgents();
        }
    }

    function confirm() {
        callAPI('/repo/'+repo.id+'/snapshot/'+snapshot.id+'/restore', {
            method: 'POST',
            body: {
                agent: agent,
                target: target,
                include: include,
                exclude: exclude,
                paths: snapshot.paths
            }
        })
        .then(() => {
            toggleModal();
            dispatch('refresh', {});
        })
    }
</script>
<style>
    
</style>

<button class="btn btn-link" on:click={toggleModal}>
    <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#download"/></svg>
</button>
{#if showModal}
    <Modal on:close={toggleModal}>
        <h2 slot="header">
			Restore repository
		</h2>
        
        <!-- svelte-ignore a11y-label-has-associated-control -->
        <label class="form-label">Snapshot</label>
        <input type="text" class="form-control" name="id" bind:value={snapshot.id} disabled>

        <!-- svelte-ignore a11y-label-has-associated-control -->
        <label class="form-label mt-3">Paths</label>
        <input type="text" class="form-control" name="paths" bind:value={snapshot.paths} disabled>

        <!-- svelte-ignore a11y-label-has-associated-control -->
        <label class="form-label mt-3">Agent</label>
        <select name="agent" class="searchbox" bind:value={agent} style="width: 100%;">
            <option value="-1" selected>None</option>
            {#each agents as agent}
                <option value={agent.ID}>{agent.Name}</option>
            {/each}
        </select>

        <!-- svelte-ignore a11y-label-has-associated-control -->
        <label class=" form-label mt-3">Destination</label>
        <input type="text" class="form-control" name="destination" bind:value={destination}>
        
        <!-- svelte-ignore a11y-label-has-associated-control -->
        <label class="form-label mt-3">Include</label>
        <input type="text" class="form-control" name="include" bind:value={include}>
        
        <!-- svelte-ignore a11y-label-has-associated-control -->
        <label class="form-label mt-3">Exclude</label>
        <input type="text" class="form-control" name="exclude" bind:value={exclude}>
        
        <div slot="buttons" class="float-end" style="display: inline-block;">
            <button type="button" class="btn btn-primary float-end" on:click={confirm} disabled={( agent == -1 )}>Restore</button>
            <button type="button" class="btn btn-secondary float-end mx-1" on:click={toggleModal}>Close</button>
        </div>
	</Modal>
{/if}