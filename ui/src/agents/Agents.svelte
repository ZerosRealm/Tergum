<script>
    import { onMount } from 'svelte';
    import New from './New.svelte'
    import Edit from './Edit.svelte'
    import Delete from './Delete.svelte'
    import { callAPI }  from '../common/API.js';

    let loading = true;
    let agents = [];

    onMount(async () => {
        getAgents();
	});

    function getAgents() {
        loading = true;
        callAPI('/agent', {
            method: 'GET'
        })
        .then(data => {
            loading = false;
            agents = data.agents;
        })
    }


    function refresh(e) {
        loading = true;
        getAgents();
    }

    function add(e) {
        agents = [...agents, e.detail];
    }
</script>
<style>

</style>
<div>
    <New on:refresh={refresh} on:add={add} />
    <table class="table">
        <thead>
            <tr>
                <th scope="col">#</th>
                <th scope="col">Name</th>
                <th scope="col">IP</th>
                <th scope="col">Port</th>
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
                {#each agents as agent}
                <tr>
                    <th scope="row">{agent.id}</th>
                    <td>{agent.name}</td>
                    <td>{agent.ip}</td>
                    <td>{agent.port}</td>
                    <td>
                        <Delete bind:agent={agent} on:refresh={refresh} />
                        <Edit bind:agent={agent} on:refresh={refresh} />
                    </td>
                </tr>
                {/each}
            {/if}
        </tbody>
    </table>
</div>