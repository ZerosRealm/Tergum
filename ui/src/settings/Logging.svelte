<script>
    import { onMount} from 'svelte';
    import { callAPI }  from '../common/API.js';

    onMount(async () => {
        getSettings();
	});

    let loading = true;
    let settings = {};

    function getSettings() {
        loading = true;
        callAPI('/setting/logging', {
            method: 'GET'
        })
        .then(data => {
            loading = false;
            settings = data;
        })
    }

    function save() {
        loading = true;
        callAPI('/setting/logging', {
            method: 'PUT',
            body: JSON.stringify(settings)
        })
        .then(data => {
            loading = false;
            settings = data;
        })
    }
</script>
<style>
    h2 {
        font-size: 1.5em;
        margin-bottom: 0.5em;
    }
</style>

<h2>Logging</h2>
{#if loading}
<div class="spinner-grow position-absolute top-50 start-50 translate-middle" role="status">
    <span class="visually-hidden">Loading...</span>
</div>
{:else}
<label for="level" class="form-label">Level</label>
<select id="level" class="form-select" aria-label="Logging level" bind:value={settings.level}>
    <option value="error">Error</option>
    <option value="warning">Warning</option>
    <option value="info">Info</option>
    <option value="debug">Debug</option>
    <option value="trace">Trace</option>
</select>
<span><i><b>Note:</b> Error is the least verbose, only showing errors, debug being the most verbose showing all levels, usually used in testing.</i></span>

<button type="button" class="btn btn-primary float-end" on:click={save}>Save</button>
{/if}