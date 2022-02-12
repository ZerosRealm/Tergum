<script>
    import { createEventDispatcher } from 'svelte';
    import Modal from '../common/Modal.svelte';
    import { callAPI }  from '../common/API.js';

    let showModal = false;
    let forget = {};

    // NOTICE: Hardcoded ID, since currently only support one.
    const forgetID = 0;
    function getForget() {
        callAPI('/forget/'+forgetID, {
            method: 'GET'
        })
        .then(data => {
            forget = data.forget;
        })
    }

    function toggleModal() {
        showModal = !showModal;

        if (showModal) {
            getForget();
        }
    }

    function save() {
       callAPI('/forget/'+forgetID, {
            method: 'PUT',
            body: JSON.stringify({
                enabled: forget.enabled,
                lastX: forget.lastX,
                hourly: forget.hourly,
                daily: forget.daily,
                weekly: forget.weekly,
                monthly: forget.monthly,
                yearly: forget.yearly,
            })
        })
        .then(data => {
            forget = data.forget;
            toggleModal();
        })
    }
</script>
<style>
    
</style>

<button class="btn btn-primary" type="button" on:click={toggleModal}>
    Forget policy<svg class="bi ms-2" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#clock" /></svg>
</button>
{#if showModal}
    <Modal on:close={toggleModal}>
        <h2 slot="header">
			Forget policy
            <div class="form-check form-switch fs-5 float-end">
                <input class="form-check-input" type="checkbox" bind:checked={forget.enabled}>
            </div>
		</h2>
        <span><i><b>Note:</b> keeping this disabled will keep all backups forever.</i></span>

        <label for="lastX" class="form-label mt-3">Keep last X snapshots</label>
        <input type="number" class="form-control" name="lastX" placeholder="lastX" bind:value={forget.lastX}>

        <label for="hourly" class="form-label mt-3">Hourly</label>
        <input type="number" class="form-control" name="hourly" placeholder="hourly" bind:value={forget.hourly}>

        <label for="daily" class="form-label mt-3">Daily</label>
        <input type="number" class="form-control" name="daily" placeholder="daily" bind:value={forget.daily}>

        <label for="weekly" class="form-label mt-3">Weekly</label>
        <input type="number" class="form-control" name="weekly" placeholder="weekly" bind:value={forget.weekly}>

        <label for="monthly" class="form-label mt-3">Monthly</label>
        <input type="number" class="form-control" name="monthly" placeholder="monthly" bind:value={forget.monthly}>

        <label for="yearly" class="form-label mt-3">Yearly</label>
        <input type="number" class="form-control" name="yearly" placeholder="yearly" bind:value={forget.yearly}>
        
        <div slot="buttons" class="float-end" style="display: inline-block;">
            <button type="button" class="btn btn-primary float-end" on:click={save}>Save</button>
            <button type="button" class="btn btn-secondary float-end mx-1" on:click={toggleModal}>Close</button>
        </div>
	</Modal>
{/if}