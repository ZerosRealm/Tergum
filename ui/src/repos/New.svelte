<script>
    import { createEventDispatcher} from 'svelte';
    import Modal from '../common/Modal.svelte';
    import { callAPI }  from '../common/API.js';

    const dispatch = createEventDispatcher();
    
    let showModal = false;

    let name = "";
    let repository = "";
    let password = "";
    let settings = "";

    function toggleModal() {
        showModal = !showModal;
    }

    function create() {
        let newSettings = settings.split("\n");

        if (newSettings.length == 1 && newSettings[0] == "") {
            newSettings = []
        }

        callAPI('/repo', {
            method: 'POST',
            body: JSON.stringify({
                name: name,
                repo: repository,
                password: password,
                settings: newSettings
            }),
        })
        .then(data => {
            toggleModal();
            dispatch('add', data.repo);
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
			New repository
		</h2>

		<label class="form-label">Name</label>
        <input type="text" class="form-control" name="name" placeholder="Display name" bind:value={name}>
        <div class="invalid-feedback" class:d-block={name == ""}>
            Please provide a name.
        </div>

        <label class="form-label mt-3">Repository</label>
        <input type="text" class="form-control" name="repo" placeholder="Repository" bind:value={repository}>
        <span><i><b>Note:</b> might be the whole connection string, eg. sftp:user@host:/srv/restic-repo</i></span>
        <div class="invalid-feedback" class:d-block={repository == ""}>
            Please provide the repository details.
        </div>

        <label class="form-label mt-3">Password</label>
        <input type="text" class="form-control" name="password" placeholder="Repository password" bind:value={password}>
        <div class="invalid-feedback" class:d-block={password == ""}>
            Please provide the password.
        </div>

        <label class="form-label mt-3">Settings</label>
        <textarea class="form-control" name="settings" rows="3" bind:value={settings}></textarea>
        <span><i><b>Note:</b> this is for extra environment variables, eg. for S3 settings</i></span>

        <div slot="buttons" class="float-end" style="display: inline-block;">
            <button type="button" class="btn btn-primary float-end" on:click={create} disabled={ (name == "" || repository == "" || password == "") }>Create</button>
            <button type="button" class="btn btn-secondary float-end mx-1" on:click={toggleModal}>Close</button>
        </div>
	</Modal>
{/if}