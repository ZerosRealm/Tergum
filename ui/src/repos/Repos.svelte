<script>
    import { onMount} from 'svelte';
    import { callAPI }  from '../common/API.js';
    import Snapshots from './snapshots/Snapshots.svelte'
    import New from './New.svelte'
    import Edit from './Edit.svelte'
    import Delete from './Delete.svelte'

    let loading = true;

    onMount(async () => {
        getRepos();
	});

    function getRepos() {
        callAPI('/repo', {
            method: 'GET'
        })
        .then(data => {
            loading = false;
            repos = data.repos;
        })
    }

    let repos = [];

    function refresh(e) {
        loading = true;
        getRepos();
    }

    function add(e) {
        repos = [...repos, e.detail];
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
                <th scope="col">Repo</th>
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
                {#each repos as repo}
                <tr>
                    <th scope="row">{repo.id}</th>
                    <td>{repo.name}</td>
                    <td>{repo.repo}</td>
                    <td>
                        <Delete bind:repo={repo} on:refresh={refresh} />
                        <Edit bind:repo={repo} on:refresh={refresh} />
                        <Snapshots bind:repo={repo} />
                    </td>
                </tr>
                {/each}
            {/if}
        </tbody>
    </table>
</div>