<script>
    import { onMount, createEventDispatcher} from 'svelte';
    import Modal from '../common/Modal.svelte';
    import { addToast }  from '../common/toasts.js';
    import { callAPI }  from '../common/API.js';

    const dispatch = createEventDispatcher();

    onMount(async () => {
        getRepos();
	});
    
    let showModal = false;

    let target = -1;
    let source = "";
    let schedule = "* * * * *";
    let exclude = "";

    function toggleModal() {
        showModal = !showModal;
    }

    let repos = [];
    function getRepos() {
        callAPI('/repo', {
            method: 'GET'
        })
        .then(data => {
            repos = data.repos;
        })
    }

    function confirm() {
        const regex = /^(\*|([0-9]|1[0-9]|2[0-9]|3[0-9]|4[0-9]|5[0-9])|\*\/([0-9]|1[0-9]|2[0-9]|3[0-9]|4[0-9]|5[0-9])) (\*|([0-9]|1[0-9]|2[0-3])|\*\/([0-9]|1[0-9]|2[0-3])) (\*|([1-9]|1[0-9]|2[0-9]|3[0-1])|\*\/([1-9]|1[0-9]|2[0-9]|3[0-1])) (\*|([1-9]|1[0-2])|\*\/([1-9]|1[0-2])) (\*|([0-6])|\*\/([0-6]))$/gm;
        if (!regex.test(schedule)) {
            addToast({
                type: "error",
                title: "Invalid schedule",
                message: "Schedule does not match the cron syntax"
            })
            return;
        }

        callAPI('/backup', {
            method: 'POST',
            body: JSON.stringify({
                target: parseInt(target),
                source: source,
                schedule: schedule,
                exclude: exclude.split('\n')
            })
        })
        .then(data => {
            toggleModal();
            dispatch('add', data.backup);
        })
    }
</script>
<style>

</style>
<button class="btn btn-primary" type="button" on:click={toggleModal}>
    New
</button>
{#if showModal}
    <Modal on:close={toggleModal}>
		<h2 slot="header">
			New backup
		</h2>

        <label for="repo" class="form-label">Repository</label>
        <select name="repo" class="form-control searchbox" style="width: 100%;" bind:value={target}>
            <option value="-1" selected>None</option>
            {#each repos as repo}
                <option value={repo.id}>{repo.name}</option>
            {/each}
        </select>
        <div class=" invalid-feedback">
            Please select a repository.
        </div>

        <label for="source" class="form-label mt-3">Source</label>
        <input type="text" class="form-control" name="source" placeholder="Source" bind:value={source}>
        <div class="invalid-feedback">
            Please provide a source.
        </div>

        <label for="schedule" class="form-label mt-3">Schdule</label>
        <input type="text" class="form-control" name="schedule" placeholder="* * * * *" bind:value={schedule}>
        <div class="invalid-feedback">
            Please provide a valid schedule.
        </div>

        <label for="exclude" class="form-label mt-3">Exclude</label>
        <textarea class="form-control" name="exclude" rows="3" bind:value={exclude}></textarea>
        <span><i><b>Note:</b> new line for each exclusion</i></span>

        <div slot="buttons" class="float-end" style="display: inline-block;">
            <button type="button" class="btn btn-primary float-end" on:click={confirm} disabled={ (target == -1 || source == "" || schedule == "") }>Create</button>
            <button type="button" class="btn btn-secondary float-end mx-1" on:click={toggleModal}>Close</button>
        </div>
	</Modal>
{/if}