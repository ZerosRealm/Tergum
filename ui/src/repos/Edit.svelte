<script>
    import { onMount, createEventDispatcher } from 'svelte';
    import Modal from '../common/Modal.svelte';
    import { callAPI }  from '../common/API.js';
    
    const dispatch = createEventDispatcher();
    
    export let repo = {};
    let data = repo;
    let showModal = false;

    onMount(async () => {
        console.log(repo)
	});

    function toggleModal() {
        showModal = !showModal;
    }

    function save() {
        let newSettings = [];

        if (repo.Settings) {
            newSettings = data.settings.split("\n");
        }

        if (newSettings.length == 1 && newSettings[0] == "") {
            newSettings = []
        }

        callAPI('/repo/'+repo.id, {
            method: 'PUT',
            body: JSON.stringify({
                name: data.name,
                repo: data.repo,
                password: data.password,
                settings: newSettings
            }),
        })
        .then(data => {
            toggleModal();
            repo = data.repo;
        })
    }
</script>
<style>

</style>
<button class="btn btn-link float-end ms-1" on:click={toggleModal}>
    <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#pencil-square" /></svg>
</button>
{#if showModal}
    <Modal on:close={toggleModal}>
        <h2 slot="header">
			Edit repository
		</h2>
        <label class="form-label">Name</label>
        <input type="text" class="form-control" name="name" placeholder="Display name" bind:value={data.name}>
        <div class="invalid-feedback" class:d-block={data.name == ""}>
            Please provide a name.
        </div>

        <label class="form-label mt-3">Repository</label>
        <input type="text" class="form-control" name="repo" placeholder="Repository" bind:value={data.repo}>
        <span><i><b>Note:</b> might be the whole connection string, eg. sftp:user@host:/srv/restic-repo</i></span>
        <div class="invalid-feedback" class:d-block={data.repo == ""}>
            Please provide the repository details.
        </div>

        <label class="form-label mt-3">Password</label>
        <input type="text" class="form-control" name="password" placeholder="Repository password" bind:value={data.password}>
        <div class="invalid-feedback" class:d-block={data.password == ""}>
            Please provide the password.
        </div>

        <label class="form-label mt-3">Settings</label>
        <textarea class="form-control" name="settings" rows="3" bind:value={data.settings}></textarea>
        <span><i><b>Note:</b> this is for extra environment variables, eg. for S3 settings</i></span>
        
        <div slot="buttons" class="float-end" style="display: inline-block;">
            <button type="button" class="btn btn-primary float-end" on:click={save} disabled={ (data.name == "" || data.repo == "" || data.password == "") }>Save</button>
            <button type="button" class="btn btn-secondary float-end mx-1" on:click={toggleModal}>Close</button>
        </div>
	</Modal>
{/if}