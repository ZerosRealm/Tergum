<script>
    import { onMount } from 'svelte';
    import { format  as dateFormat } from 'fecha';
    import socket  from '../../common/websocket.js';
    import { callAPI }  from '../../common/API.js';

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
        
        msg = job.progress
        if (msg == null) {
            return;
        }

        let percent = 0
        let snapshot = ""
        if (msg.message_type == "status") {
            percent = Math.floor(msg.percent_done * 100)
        }
        if (msg.message_type == "summary") {
            percent = 100
            snapshot = msg.snapshot_id
        }

        msg.percent = percent
        msg.snapshot = snapshot
        delete job.progress
        jobs[job.id] = Object.assign(job, msg)
    }

    function refresh(e) {
        loading = true;
        getJobs();
    }
</script>
<style>
 .job-done {
     background-color: #004F39;
 }
</style>
<h2>Jobs</h2>
<table class="table">
    <thead>
        <tr>
            <th scope="col">#</th>
            <th scope="col">Start</th>
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
                        <div class="progress progress-bar" class:job-done={job.percent==100} role="progressbar" style="width: {job.percent}%;" aria-valuenow="{job.percent}" aria-valuemin="0" aria-valuemax="100">{job.percent}%</div>
                    </td>
                    <td>{job.snapshot}</td>
                    <td>
                        <View job={job} on:refresh={refresh} />
                        <Stop job={job} on:refresh={refresh} />
                    </td>
                </tr>
            {/each}
        {/if}
    </tbody>
</table>