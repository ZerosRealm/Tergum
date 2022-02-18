<script>
    import { onMount } from 'svelte';
    import { format  as dateFormat } from 'fecha';
    import socket  from '../../common/websocket.js';
    import { callAPI }  from '../../common/API.js';
    import { addToast }  from '../../common/toasts.js';

    import View from './View.svelte'
    import Stop from './Stop.svelte'

    const nullDate = "0001-01-01T00:00:00Z"

    let loading = true;

    onMount(async () => {
        getJobs();
	});

    let jobs = {};
    function getJobs() {
        loading = true;
        callAPI('/job', {
            method: 'GET'
        })
        .then(data => {
            jobs = data.jobs;
            for (const id in jobs) {
                console.log(jobs[id]);
                parseJob(jobs[id]);
            }
            loading = false;
        })
    }

    socket.subscribe(event => {
        if (event.data == "") {
            return
        }

        let data = JSON.parse(event.data);
        if (data.type.toLowerCase() != "job_progress") {
            return;
        }
        parseJob(data.job);

    });

    function parseJob(job) {
        // Check if jobs contains this job
        if (!(job.id in jobs)) {
            jobs[job.id] = job;
        }

        if (job.progress == null) {
            jobs[job.id] = job
            return;
        }

        let percent = 0
        let snapshot = ""
        if (job.progress.message_type == "status") {
            percent = Math.floor(job.progress.percent_done * 100)
        }
        if (job.progress.message_type == "summary") {
            percent = 100
        }

        job.progress.percent = percent

        // delete job.progress
        jobs[job.id] = job
    }

    socket.subscribe(event => {
        if (event.data == "") {
            return
        }

        let data = JSON.parse(event.data);
        if (data.type.toLowerCase() != "job_error") {
            return;
        }
        msg = data.msg

        addToast({
            type: "error",
            title: "Error!",
            message: msg,
        })
    });

    function refresh(e) {
        loading = true;
        getJobs();
    }
</script>
<style>
 .job-done {
     background-color: #004F39;
 }

.job-error {
    background-color: #8B0000;
}
</style>
<h2>Jobs</h2>
<table class="table">
    <thead>
        <tr>
            <th scope="col">#</th>
            <th scope="col">Start</th>
            <th scope="col">End</th>
            <th scope="col">Progress</th>
            <th scope="col">Snapshot</th>
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
            {#each Object.entries(jobs).reverse() as [id, job]}
                <tr>
                    <th scope="row">{id}</th>
                    <td>
                        {#if job.start_time == nullDate || job.start_time == undefined}
                            Never
                        {:else}
                            {dateFormat((new Date(job.start_time)), "YYYY-MM-DD HH:mm:ss")}
                        {/if}
                    </td>
                    <td>
                        {#if job.end_time == nullDate || job.end_time == undefined}
                            Never
                        {:else}
                            {dateFormat((new Date(job.end_time)), "YYYY-MM-DD HH:mm:ss")}
                        {/if}
                    </td>
                    <td>
                        {#if job.progress != null && !job.aborted}
                            <div class="progress progress-bar" class:job-done={job.progress.percent==100} role="progressbar" style="width: {job.progress.percent}%;" aria-valuenow="{job.progress.percent}" aria-valuemin="0" aria-valuemax="100">{job.progress.percent}%</div>
                        {:else}
                            {#if !job.aborted}
                                <div class="progress progress-bar" role="progressbar" style="width: 0%;" aria-valuenow="0" aria-valuemin="0" aria-valuemax="100">0%</div>
                            {:else}
                                <div class="progress progress-bar job-error" role="progressbar" style="width: 100%;" aria-valuenow="100" aria-valuemin="0" aria-valuemax="100">ABORTED</div>
                            {/if}
                        {/if}
                    </td>
                    <td>
                        {#if job.progress != null && job.progress.snapshot_id != null}
                            {job.progress.snapshot_id}
                        {/if}
                    </td>
                    <td>
                        <View job={job} on:refresh={refresh} />
                        <Stop job={job} on:refresh={refresh} />
                    </td>
                </tr>
            {/each}
        {/if}
    </tbody>
</table>