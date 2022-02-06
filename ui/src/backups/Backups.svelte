<script>
    import { onMount} from 'svelte';
    import { format  as dateFormat } from 'fecha';
    import { callAPI }  from '../common/API.js';

    import New from './New.svelte'
    import Edit from './Edit.svelte'
    import Start from './Start.svelte'
    import Delete from './Delete.svelte'
    import Jobs from './jobs/Jobs.svelte'

    let loading = true;
    const nullDate = "0001-01-01T00:00:00Z"

    onMount(async () => {
        getBackups();
	});

    function getBackups() {
        loading = true;
        callAPI('/backup', {
            method: 'GET'
        })
        .then(data => {
            loading = false;
            backups = data.backups;
        })
    }

    let backups = [];

    function refresh(e) {
        loading = true;
        getBackups();
    }

    function add(e) {
        backups = [...backups, e.detail];
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
                <th scope="col">Source</th>
                <th scope="col">Schedule</th>
                <th scope="col">Last backup</th>
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
                {#each backups as backup}
                <tr>
                    <th scope="row">{backup.id}</th>
                    <td>{backup.source}</td>
                    <td>{backup.schedule}</td>
                    <td>
                        {#if backup.last_run == nullDate}
                            Never
                        {:else}
                            {dateFormat((new Date(backup.last_run)), "YYYY-MM-DD HH:mm:ss")}
                        {/if}
                    </td>
                    <td>
                        <Delete bind:backup={backup} on:refresh={refresh} />
                        <Edit bind:backup={backup} on:refresh={refresh} />
                        <Start bind:backup={backup} on:refresh={refresh} />
                    </td>
                </tr>
                {/each}
            {/if}
        </tbody>
    </table>

    <Jobs />
</div>