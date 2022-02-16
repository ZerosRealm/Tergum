<script>
    import { onMount} from 'svelte';
    import { callAPI }  from '../common/API.js';

    import LogLine from './LogLine.svelte';

    onMount(async () => {
        getLogs();
	});

    let loading = true;
    let logs = [];

    let filteredLogs = [];
    let shownLogs = [];

    function getLogs() {
        loading = true;
        callAPI('/log', {
            method: 'GET'
        })
        .then(data => {
            logs = data.logs;
            filteredLogs = logs;
            updatePages();
            updateLogsForPage();

            loading = false;
        })
    }

    let amountPerPage = 10;
    let currentPage = 1;
    let pages = [];

    function updateLogsForPage() {
        let start = (currentPage - 1) * amountPerPage;
        let end = start + amountPerPage;
        shownLogs = filteredLogs.slice(start, end);
    }

    function updatePages() {
        pages = [];
        let amountOfPages = Math.ceil(filteredLogs.length / amountPerPage);
        for (let i = 1; i <= amountOfPages; i++) {
            pages.push(i);
        }
    }

    function changePage(page) {
        if (currentPage == page) {
            return
        }

        currentPage = page;
        updateLogsForPage();
    }

    function nextPage() {
        if (currentPage != Math.ceil(filteredLogs.length / amountPerPage)) {
            currentPage++;
            updateLogsForPage();
        }
    }

    function prevPage() {
        if (currentPage > 1) {
            currentPage--;
            updateLogsForPage();
        }
    }

    let filter = "";
    function filterLogs() {
        currentPage = 1;
        if (filter == "") {
            filteredLogs = logs;
            shownLogs = filteredLogs;

            updatePages();
            updateLogsForPage();
            return
        }

        filteredLogs = logs.filter(log => {
            return log.level == filter;
        });
        shownLogs = filteredLogs;
        updatePages();
        updateLogsForPage();
    }
</script>
<style>
    h2 {
        font-size: 1.5em;
        margin-bottom: 0.5em;
    }

    .page-item:hover {
        cursor: pointer;
    }
</style>

<h2>Logs</h2>
{#if loading}
<div class="spinner-grow position-absolute top-50 start-50 translate-middle" role="status">
    <span class="visually-hidden">Loading...</span>
</div>
{:else}

<label for="level" class="form-label">Level filter</label>
<select id="level" class="form-select" on:blur={filterLogs} aria-label="Filter level" bind:value={filter}>
    <option value="">All</option>
    <option value="error">Error</option>
    <option value="warning">Warning</option>
    <option value="info">Info</option>
    <option value="debug">Debug</option>
</select>

<table class="table">
    <thead>
        <tr>
            <th style="width: 2%;text-align:center;">Level</th>
            <th style="width: 10%;text-align:center;">Time</th>
            <th>Log</th>
        </tr>
    </thead>
    <tbody>
        {#each shownLogs as log}
        <tr class:table-danger={log.level == "error" || log.level == "fatal" || log.level == "panic"} class:table-warning={log.level == "warn"} class:table-secondary={log.level == "info"} class:table-light={log.level == "debug"}>
            <td>{log.level}</td>
            <td>{log.ts}</td>
            <LogLine bind:log={log} />
        </tr>
        {/each}
    </tbody>
</table>
<div class="center">
    <nav aria-label="Logs pagination">
        <ul class="pagination justify-content-center">
            <li class="page-item" on:click={prevPage} class:disabled={currentPage==1}>
                <a class="page-link" aria-label="Previous" href="#">
                    <span aria-hidden="true">&laquo;</span>
                </a>
            </li>
            
            {#each pages as page}
            <li class="page-item" on:click={()=>{ changePage(page) }} class:active={currentPage==page}><a class="page-link" href="#">{page}</a></li>
            {/each}

            <li class="page-item" on:click={nextPage} class:disabled={currentPage==Math.ceil(filteredLogs.length / amountPerPage)}>
                <a class="page-link" aria-label="Next" href="#">
                    <span aria-hidden="true">&raquo;</span>
                </a>
            </li>
        </ul>
    </nav>
</div>
{/if}