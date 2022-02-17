<script>
    import { onMount} from 'svelte';
    import { callAPI }  from '../common/API.js';
    import { addToast }  from '../common/toasts.js';

    onMount(async () => {
        getSettings();
	});

    let loading = true;
    let token = "";
    let enabled = false;

    function getSettings() {
        loading = true;
        callAPI('/setting/registration-enabled', {
            method: 'GET'
        })
        .then(data => {
            enabled = data.value;

            callAPI('/setting/registration-token', {
                method: 'GET'
            })
            .then(data => {
                token = data.value;
                loading = false;
            })
        })
    }

    function save() {
        callAPI('/setting/registration-enabled', {
            method: 'PUT',
            body: JSON.stringify({
                value: enabled
            })
        })
        .then(data => {
            enabled = data.value;

            callAPI('/setting/registration-token', {
                method: 'PUT',
                body: JSON.stringify({
                    value: token
                })
            })
            .then(data => {
                token = data.value;
                
                addToast({
                    type: 'success',
                    title: 'Saved!',
                    message: 'Settings saved.'
                });
            })
        })
    }

    function generatePSK() {
        let size = 64
        token = [...Array(size)].map(() => Math.floor(Math.random() * 16).toString(16)).join('');
    }
</script>
<style>
    h2 {
        font-size: 1.5em;
        margin-bottom: 0.5em;
    }
</style>

<h2>Registration</h2>
{#if loading}
<div class="spinner-grow position-absolute top-50 start-50 translate-middle" role="status">
    <span class="visually-hidden">Loading...</span>
</div>
{:else}
<div class="form-check form-switch">
    <input class="form-check-input" type="checkbox" bind:checked={enabled}>
    <label for="enabled" class="form-label">Enabled</label>
</div>

<label for="token" class="form-label mt-3">Token</label>
<div class="w-100 d-inline-flex">
    <input type="text" class="form-control" name="token" placeholder="Token" bind:value={token}>
    <button class="btn btn-primary ms-2" on:click={generatePSK}>
    <svg class="bi" width="16" height="16" fill="currentColor">
        <use xlink:href="css/bootstrap-icons.svg#arrow-clockwise" />
    </svg>
    </button>
</div>

<br>
<button type="button" class="btn btn-primary" on:click={save}>Save</button>
{/if}