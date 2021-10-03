$( document ).ready(function() {
    cache.currentPage = "backups"
    refresh()
});

function gotBackups(data) {
    let list = $("#backups table tbody")
    list.empty()
    data.backups.forEach(backup => {
        let item = $(`<tr>
                        <th scope="row">`+backup.ID+`</th>
                        <td>`+backup.Source+`</td>
                        <td>`+backup.Schedule+`</td>
                        <td>
                            <button class="btn btn-danger float-end ms-1" onclick="confirmationBox('Do you want to delete this backup?', () => deleteBackup(`+backup.ID+`))">
                                <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#trash"/></svg>
                            </button>
                            <button class="btn btn-link float-end ms-1" onclick="editBackup(`+backup.ID+`)" data-bs-toggle="modal" data-bs-target="#editBackup">
                                <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#pencil-square"/></svg>
                            </button>
                        </td>
                    </tr>`)

        list.append(item)
    });
}

function newBackup() {
    resetInvalidForms()
    let repo = $("#newBackup .card select[name='repo']").val()
    let source = $("#newBackup .card input[name='source']").val()
    let schedule = $("#newBackup .card input[name='schedule']").val()
    let exclude = $("#newBackup .card textarea[name='exclude']").val().split('\n')
    
    let repoInvalid = false
    if (repo == -1) {
        repoInvalid = true
        invalidateField("#newBackup .card select[name='repo']")
    }
    
    const regex = /(28|\*) (2|\*) (7|\*) (1|\*) (1|\*)/gm;
    let scheduleInvalid = false
    if (!regex.test(schedule)) {
        scheduleInvalid = true
        invalidateField("#newBackup .card input[name='schedule']")
    }

    if (hasInvalidFields("#newBackup .card") || scheduleInvalid || repoInvalid) {
        return
    }

    if (exclude.length == 1 && exclude[0] == "") {
        exclude = []
    }

    var msg = {
        type: "newbackup",
        target: parseInt(repo),
        source: source,
        schedule: schedule,
        exclude: exclude,
    };
    webSocket.send(JSON.stringify(msg));
    closeModal("#newBackup")
}

function prepareNewBackup() {
    let selectElem = $("#backups select[name='repo']")
    selectElem.empty()

    let newOption = $(`<option value="-1">None</option>`)
    selectElem.append(newOption)

    cache.repos.forEach(repo => {
        let newOption = $(`<option value="`+repo.ID+`">`+repo.Name+`</option>`)
        selectElem.append(newOption)
    });
}

function editBackup(id) {
    $("#editBackup input").val("")

    let idInput = $("#editBackup input[name='id']")
    let source = $("#editBackup input[name='source']")
    let schedule = $("#editBackup input[name='schedule']")
    let exclude = $("#editBackup textarea[name='exclude']")

    let data = null
    cache.backups.forEach(backup => {
        if (backup.ID == id) {
            data = backup
        }
    });

    if (data == null) {
        showError("Cannot edit - no backup with that ID.")
        return
    }

    let selectElem = $("#editBackup select[name='repo']")
    selectElem.empty()

    cache.repos.forEach(repo => {
        if (repo.ID == data.Target) {
            let newOption = $(`<option value="`+repo.ID+`" selected>`+repo.Name+`</option>`)
            selectElem.append(newOption)
        } else {
            let newOption = $(`<option value="`+repo.ID+`">`+repo.Name+`</option>`)
            selectElem.append(newOption)
        }
    });

    selectElem = $("#editBackup select[name='search']")
    selectElem.empty()

    newOption = $(`<option value="-1">None</option>`)
    selectElem.append(newOption)

    cache.agents.forEach(agent => {
        let newOption = $(`<option value="`+agent.ID+`">`+agent.Name+`</option>`)
        selectElem.append(newOption)
    });

    $("#editBackup .subscribers").empty()

    if (cache.subscribers[id]) {
        cache.subscribers[id].forEach(sub => {
            let newElem = $(`<div class="subscriber">
                <input name="id" type="hidden" value="`+sub.ID+`"/>
                <button type="button" class="btn btn-dark">`+sub.Name+`</button>
                <button id="btn-delete" type="button" class="btn btn-danger" onclick="$(this.parentElement).remove()">X</button>
            </div>`)
            $("#editBackup .subscribers").append(newElem)
        });
    }

    idInput.val(data.ID)
    source.val(data.Source)
    schedule.val(data.Schedule)
    if (data.Exclude) {
        exclude.val(data.Exclude.join("\n"))
    }
}

function addSubscriber() {
    let id = $("#editBackup select[name='search']").val()

    if (id == -1)  {
        return
    }

    if (($(`#editBackup .subscriber input[name="id"][value="`+id+`"]`)).length >= 1) {
        return
    }

    let foundAgent = null
    cache.agents.forEach(agent => {
        if (agent.ID == id) {
            foundAgent = agent
        }
    });

    if (foundAgent == null) {
        showError("Can't add subscriber - no agent with that ID.")
        return
    }

    let newElem = $(`<div class="subscriber">
            <input name="id" type="hidden" value="`+foundAgent.ID+`"/>
            <button type="button" class="btn btn-dark">`+foundAgent.Name+`</button>
            <button id="btn-delete" type="button" class="btn btn-danger" onclick="$(this.parentElement).remove()">X</button>
        </div>`)
    $("#editBackup .subscribers").append(newElem)
}

function updateBackup() {
    let id = $("#editBackup input[name='id']").val()
    let repo = $("#editBackup select[name='repo']").val()
    let source = $("#editBackup input[name='source']").val()
    let schedule = $("#editBackup input[name='schedule']").val()

    var msg = {
        type: "updatebackup",
        id: parseInt(id),
        target: parseInt(repo),
        source: source,
        schedule: schedule,
    };
    webSocket.send(JSON.stringify(msg));

    let subscribers = []
    $(`#editBackup .subscriber input[name="id"]`).each(function() {
        subscribers.push(parseInt($(this).val()))
    });

    msg = {
        type: "updatesubscribers",
        backup: parseInt(id),
        agents: subscribers
    };
    webSocket.send(JSON.stringify(msg));
}

function deleteBackup(id) {
    var msg = {
        type: "deletebackup",
        id: parseInt(id),
    };
    webSocket.send(JSON.stringify(msg));
}

function showJob(jobID) {
    let job = JSON.parse(atob(cache.jobs[jobID]))

    if (job == undefined || job == null) {
        showError("No job with that ID in cache.")
        return
    }

    let modal = $("#jobs #showJob .modal-body .info")
    modal.empty()

    // TODO: Convert bytes to MB or GB?
    for (const key in job) {
        if (key == "message_type") {
            continue
        }
        let val = job[key]

        if (key == "percent_done") {
            val = Math.floor(val*100)+"%"
        }

        let newElem = $(`
            <label class="form-label mt-3">`+ capitalize(key).replaceAll("_", " ") +`</label>
            <input type="text" class="form-control" value="`+val+`" disabled>
        `)
        modal.append(newElem)
    }
}

function gotJobs(data) {
    $(`#jobs table tbody`).empty()
    for (const job in data.jobs) {
        if (Object.hasOwnProperty.call(data.jobs, job)) {
            const elem = data.jobs[job]
            const msg = JSON.parse(atob(elem))
            let percent = 0
            let snapshot = ""
            if (msg.message_type == "status") {
                percent = Math.floor(msg.percent_done * 100)
            }
            if (msg.message_type == "summary") {
                percent = 100
                snapshot = msg.snapshot_id
            }

            let html = `<th scope="row">`+job+`</th>
                <td>
                    <div class="progress">
                        <div class="progress-bar" role="progressbar" style="width: `+percent+`%;" aria-valuenow="`+percent+`" aria-valuemin="0" aria-valuemax="100">`+percent+`%</div>
                    </div>
                </td>
                <td>
                    `+snapshot+`
                </td>
                <td>
                    <button class="btn btn-link float-end ms-1" onclick="showJob('`+job+`')" data-bs-toggle="modal" data-bs-target="#showJob">
                        <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#search"/></svg>
                    </button>
                </td>`

            let newElem = $(`<tr>`+html+`</tr>`)
            $(`#jobs table tbody`).append(newElem)
        }
    }
}

function gotJobProgress(data) {
    let jobElems = $(`#jobs table tbody`).children()
    
    let foundElem = null
    for (let i = 0; i < jobElems.length; i++) {
        const jobElem = jobElems[i];
        let children = $(jobElem).children()
        if (children.length != 0) {
            let id = children[0].innerText
            if (id == data.job) {
                foundElem = jobElem
            }
        }
    }

    let percent = 0
    let snapshot = ""
    if (data.msg.message_type == "status") {
        percent = Math.floor(data.msg.percent_done * 100)
    }
    if (data.msg.message_type == "summary") {
        percent = 100
        snapshot = data.msg.snapshot_id
    }

    let html = `<th scope="row">`+data.job+`</th>
        <td>
            <div class="progress">
                <div class="progress-bar" role="progressbar" style="width: `+percent+`%;" aria-valuenow="`+percent+`" aria-valuemin="0" aria-valuemax="100">`+percent+`%</div>
            </div>
        </td>
        <td>
            `+snapshot+`
        </td>
        <td>
            <button class="btn btn-link float-end ms-1" onclick="showJob('`+data.job+`')" data-bs-toggle="modal" data-bs-target="#showJob">
                <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#pencil-square"/></svg>
            </button>
        </td>`

    if (foundElem == null) {
        foundElem = $(`<tr>`+html+`</tr>`)
        $(`#jobs table`).append(foundElem)
        return
    }
    foundElem.innerHTML = html
}