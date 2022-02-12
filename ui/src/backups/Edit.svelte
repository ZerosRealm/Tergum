<script>
    import { createEventDispatcher } from 'svelte';
    import Modal from '../common/Modal.svelte';
    import { callAPI }  from '../common/API.js';
    
    export let backup = {};
    let data = backup;
    let showModal = false;
    let chosenAgent = -1;

    let subcribersChanged = false;

    let subscribers = [];
    function getSubscribers() {
        callAPI('/backup/'+backup.id+'/agent', {
            method: 'GET'
        })
        .then(data => {
            subscribers = data.agents;
        })
    }
    
    let agents = [];
    function getAgents() {
        callAPI('/agent', {
            method: 'GET'
        })
        .then(data => {
            agents = data.agents;
        })
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

    function toggleModal() {
        showModal = !showModal;

        if (showModal) {
            getRepos();
            getSubscribers();
            getAgents();
        }
    }

    function save() {
        let newExclude = [];

        if (backup.exclude != null && Object.prototype.toString.call(backup.exclude) !== "[object Array]") {
            newExclude = backup.exclude.split('\n');
        } else if (backup.exclude != null) {
            newExclude = backup.exclude;
        }

        if (newExclude.length == 1 && newExclude[0] == '') {
            newExclude = [];
        }

        callAPI('/backup/'+data.id, {
            method: 'PUT',
            body: JSON.stringify({
                target: data.target,
                source: data.source,
                schedule: data.schedule,
                exclude: newExclude
            })
        })
        .then(data => {
            backup = data.backup;

            let newSubscribers = [];
            subscribers.forEach(subscriber => {
                if (subscriber.id != chosenAgent) {
                    newSubscribers.push(subscriber.id);
                }
            });

            if (!subcribersChanged) {
                toggleModal();
                return
            }

            callAPI('/backup/'+backup.id+'/agent', {
                method: 'PUT',
                body: JSON.stringify({
                    agents: newSubscribers
                })
            })
            .then(() => {
                toggleModal();
            })
        })
    }

    function addSubscriber() {
        if (chosenAgent == -1) {
            return;
        }

        let found = false;
        subscribers.forEach(subscriber => {
            if (subscriber.id == chosenAgent) {
                found = true;
                return;
            }
        });

        if (found) {
            return;
        }

        let foundAgent = null;
        agents.forEach(agent => {
            if (agent.id == chosenAgent) {
                foundAgent = agent;
            }
        });

        if (foundAgent == null) {
            return;
        }

        subscribers = [...subscribers, foundAgent];
        chosenAgent = "-1";
        subcribersChanged = true;
    }

    function removeSubscriber(id) {
        let newSubscribers = [];
        subscribers.forEach(subscriber => {
            if (subscriber.id != id) {
                newSubscribers.push(subscriber);
            }
        });
        subscribers = newSubscribers;
        subcribersChanged = true;
    }
</script>
<style>
    .subscriber {
        width: 100%;
        display: inline-flex;
    }

    .subscriber .btn-dark {
        flex: 1;
        border-width: 0;
        background-color: #3B4252;
    }

    .subscriber .btn-danger {
        margin-left: 5px;
    }

    .subscriber {
        width: 100%;
        margin-top: 10px;
    }

    .search {
        width: 100%;
        display: inline-flex !important;
    }

    .search .btn {
        margin-left: 5px;
    }

    /* .searchbox {
        display: block;
        width: 100%;
        padding: .375rem .4rem;
        font-size: 1rem;
        font-weight: 400;
        line-height: 1.5;
        color: #006D77;
        background-color: #fff;
        background-clip: padding-box;
        border: 1px solid #ced4da;
        -webkit-appearance: none;
        -moz-appearance: none;
        appearance: none;
        border-radius: .25rem;
        transition: border-color .15s ease-in-out,
        box-shadow .15s ease-in-out;
    } */
</style>
<button class="btn btn-link float-end ms-1" on:click={toggleModal}>
    <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#pencil-square" /></svg>
</button>
{#if showModal}
    <Modal on:close={toggleModal}>
        <h2 slot="header">
			Edit backup
		</h2>
        
        <label for="repo" class="form-label">Repository</label>
        <select name="repo" class="form-control searchbox" style="width: 100%;" bind:value={data.target}>
            <option value="-1" selected>None</option>
            {#each repos as repo}
                <option value={repo.id}>{repo.name}</option>
            {/each}
        </select>
        <div class=" invalid-feedback">
            Please select a repository.
        </div>

        <label for="source" class="form-label mt-3">Source</label>
        <input type="text" class="form-control" name="source" placeholder="Source" bind:value={data.source}>
        <div class="invalid-feedback">
            Please provide a source.
        </div>

        <label for="schedule" class="form-label mt-3">Schdule</label>
        <input type="text" class="form-control" name="schedule" placeholder="* * * * *" bind:value={data.schedule}>
        <div class="invalid-feedback">
            Please provide a valid schedule.
        </div>

        <label for="exclude" class="form-label mt-3">Exclude</label>
        <textarea class="form-control" name="exclude" rows="3" bind:value={data.exclude}></textarea>
        <span><i><b>Note:</b> new line for each exclusion</i></span>

        <h4>Subscribers</h4>
        <div class="search">
            <select name="search" class="searchbox" style="width: 100%;" bind:value={chosenAgent}>
                <option value="-1">None</option>
                {#each agents as agent}
                    <option value={agent.id}>{agent.name}</option>
                {/each}
            </select>
            <button type=" button" class="btn btn-primary" on:click={addSubscriber}>Add</button>
        </div>
        <div class="subscribers">
            {#each subscribers as sub}
                <div class="subscriber">
                    <button type="button" class="btn btn-dark">{sub.name}</button>
                    <button id="btn-delete" type="button" class="btn btn-danger" on:click={removeSubscriber(sub.id)}>X</button>
                </div>
            {/each}
        </div>
        
        <div slot="buttons" class="float-end" style="display: inline-block;">
            <button type="button" class="btn btn-primary float-end" on:click={save} disabled={ (data.target == -1 || data.source == "" || data.schedule == "") }>Save</button>
            <button type="button" class="btn btn-secondary float-end mx-1" on:click={toggleModal}>Close</button>
        </div>
	</Modal>
{/if}