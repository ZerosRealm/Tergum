<script>
    import Modal from '../common/Modal.svelte';
    import { callAPI }  from '../common/API.js';
    
    export let agent = {};
    let data = agent;
    let showModal = false;

    function toggleModal() {
        showModal = !showModal;
    }

    function save() {
        callAPI('/agent/'+data.id, {
            method: 'PUT',
            body: JSON.stringify({
                name: data.name,
                ip: data.ip,
                port: parseInt(data.port),
                psk: data.psk
            })
        })
        .then(data => {
            toggleModal();
            agent = data.agent;
        })
    }

    function generatepsk() {
        let size = 64
        data.psk = [...Array(size)].map(() => Math.floor(Math.random() * 16).toString(16)).join('');
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
			Edit agent
		</h2>
        <label for="name" class="form-label">Name</label>
        <input type="text" class="form-control" name="name" placeholder="E.g. hostname" bind:value={data.name}>
        <div class="invalid-feedback" class:d-block={data.name == ""}>
            Please provide a name.
        </div>

        <label for="ip" class="form-label mt-3">IP</label>
        <input type="text" class="form-control" name="ip" placeholder="ip address" bind:value={data.ip}>
        <div class="invalid-feedback" class:d-block={data.ip == ""}>
            Please provide an IP.
        </div>

        <label for="port" class="form-label mt-3">port</label>
        <input type="text" class="form-control" name="port" placeholder="port" bind:value={data.port}>
        <div class="invalid-feedback" class:d-block={data.port == ""}>
            Please provide a port.
        </div>

        <label for="psk" class="form-label mt-3">PSK</label>
        <div class="w-100 d-inline-flex">
            <input type="text" class="form-control" name="psk" placeholder="Pre-shared key" bind:value={data.psk}>
            <button class="btn btn-primary ms-2" on:click={generatepsk}>
                <svg class="bi" width="16" height="16" fill="currentColor">
                    <use xlink:href="css/bootstrap-icons.svg#arrow-clockwise" /></svg>
            </button>
        </div>
        <div class="invalid-feedback" class:d-block={data.psk == ""}>
            Please provide a pre-shared key.
        </div>
        
        <div slot="buttons" class="float-end" style="display: inline-block;">
            <button type="button" class="btn btn-primary float-end" on:click={save} disabled={ (data.name == "" || data.ip == "" || data.port == "" || data.psk == "") }>Save</button>
            <button type="button" class="btn btn-secondary float-end mx-1" on:click={toggleModal}>Close</button>
        </div>
	</Modal>
{/if}