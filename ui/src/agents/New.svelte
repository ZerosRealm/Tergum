<script>
    import { createEventDispatcher} from 'svelte';
    import Modal from '../common/Modal.svelte';
    import { callAPI }  from '../common/API.js';

    const dispatch = createEventDispatcher();
    
    let showModal = false;

    let name = "";
    let IP = "";
    let port = "";
    let psk = "";

    function toggleModal() {
        showModal = !showModal;
    }

    function confirm() {

        callAPI('/agent', {
            method: 'POST',
            body: JSON.stringify({
                name: name,
                ip: IP,
                port: parseInt(port),
                psk: psk
            })
        })
        .then(data => {
            toggleModal();
            dispatch('add', data.agent);
        })
    }

    function generatePSK() {
        let size = 64
        psk = [...Array(size)].map(() => Math.floor(Math.random() * 16).toString(16)).join('');
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
			New agent
		</h2>

        <label for="name" class="form-label">Name</label>
        <input type="text" class="form-control" name="name" placeholder="E.g. hostname" bind:value={name}>
        <div class="invalid-feedback" class:d-block={name == ""}>
            Please provide a name.
        </div>

        <label for="IP" class="form-label mt-3">IP</label>
        <input type="text" class="form-control" name="ip" placeholder="IP address" bind:value={IP}>
        <div class="invalid-feedback" class:d-block={IP == ""}>
            Please provide an IP.
        </div>

        <label for="port" class="form-label mt-3">Port</label>
        <input type="text" class="form-control" name="port" placeholder="Port" bind:value={port}>
        <div class="invalid-feedback" class:d-block={port == ""}>
            Please provide a port.
        </div>

        <label for="psk" class="form-label mt-3">PSK</label>
        <div class="w-100 d-inline-flex">
            <input type="text" class="form-control" name="psk" placeholder="Pre-shared key" bind:value={psk}>
            <button class="btn btn-primary ms-2" on:click={generatePSK}>
                <svg class="bi" width="16" height="16" fill="currentColor">
                    <use xlink:href="css/bootstrap-icons.svg#arrow-clockwise" /></svg>
            </button>
        </div>
        <div class="invalid-feedback" class:d-block={psk == ""}>
            Please provide a PSK.
        </div>

        <div slot="buttons" class="float-end" style="display: inline-block;">
            <button type="button" class="btn btn-primary float-end" on:click={confirm} disabled={ (name == "" || IP == "" || port == "" || psk == "") }>Create</button>
            <button type="button" class="btn btn-secondary float-end mx-1" on:click={toggleModal}>Close</button>
        </div>
	</Modal>
{/if}