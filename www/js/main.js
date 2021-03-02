var webSocket = new WebSocket("ws://127.0.0.1/ws", "1");
var cache = {
    repos: [],
    backups: [],
    agents: [],
    subscribers: [],
    snapshots: {
        repo: 0,
        data: []
    }
}

$( document ).ready(function() {
    webSocket.onopen = function(event) {
        refresh()
    };
});

function refresh() {
    requestData("getbackups")
    requestData("getagents")
    requestData("getrepos")
    requestData("getsubscribers")
}

function showError(msg) {
    msg = capitalize(msg)
    let newAlert = $(`<div class="alert alert-danger" role="alert" style="margin-top: 12px;">
        <span>`+msg+`</span>
        <button style='margin-left: 5px;' type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
    </div>`)
    $("#alerts").append(newAlert)
}

function showSuccess(msg) {
    msg = capitalize(msg)
    let newAlert = $(`<div class="alert alert-success" role="alert" style="margin-top: 12px;">
        <span>`+msg+`</span>
        <button style='margin-left: 5px;' type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
    </div>`)
    $("#alerts").append(newAlert)
}

const capitalize = (s) => {
  if (typeof s !== 'string') return ''
  return s.charAt(0).toUpperCase() + s.slice(1)
}

function requestData(type) {
    var msg = {
        type: type,
    };
    webSocket.send(JSON.stringify(msg));
}

webSocket.onmessage = function (event) {
    if (event.data == "") {
        return
    }
    let data = JSON.parse(event.data)
    switch (data.type) {
        case "getbackups":
            gotBackups(data)
            break;
        case "getrepos":
            gotRepos(data)
            break
        case "getagents":
            gotAgents(data)
            break
        case "getsubscribers":
            cache.subscribers = data.subscribers
            break
        case "getsnapshots":
            console.log(data)
            gotSnapshots(data)
            break
        case "error":
            showError(data.message)
            break
        case "success":
            showSuccess(data.message)
            refresh()
            break
        default:
            break;
    }
}

function gotBackups(data) {
    let list = $("#backups table tbody")
    list.empty()
    data.backups.forEach(backup => {
        let item = $(`<tr>
                        <th scope="row">`+backup.ID+`</th>
                        <td>`+backup.Source+`</td>
                        <td>`+backup.Schedule+`</td>
                        <td>
                            <button class="btn btn-danger float-end ms-1" onclick="deleteBackup(`+backup.ID+`)">
                                <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#trash"/></svg>
                            </button>
                            <button class="btn btn-link float-end ms-1" onclick="editBackup(`+backup.ID+`)" data-bs-toggle="modal" data-bs-target="#editBackup">
                                <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#pencil-square"/></svg>
                            </button>
                        </td>
                    </tr>`)

        list.append(item)
    });
    cache.backups = data.backups
}

function newBackup() {
    let repo = $("#backups .card select[name='repo']").val()
    let source = $("#backups .card input[name='source']").val()
    let schedule = $("#backups .card input[name='schedule']").val()

    if (repo == -1) {
        showError("You need to select a repository.")
        return
    }

    var msg = {
        type: "newbackup",
        target: parseInt(repo),
        source: source,
        schedule: schedule,
    };
    webSocket.send(JSON.stringify(msg));
}

function editBackup(id) {
    $("#editBackup input").val("")

    let idInput = $("#editBackup input[name='id']")
    let repo = $("#editBackup select[name='repo']")
    let source = $("#editBackup input[name='source']")
    let schedule = $("#editBackup input[name='schedule']")

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
    // repo.val(data.Target)
    source.val(data.Source)
    schedule.val(data.Schedule)
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

function gotAgents(data) {
    let table = $("#agents table tbody")
    table.empty()
    data.agents.forEach(agent => {
        let item = $(`<tr>
                        <th scope="row">`+agent.ID+`</th>
                        <td>`+agent.Name+`</td>
                        <td>`+agent.IP+`</td>
                        <td>`+agent.Port+`</td>
                        <td>
                            <button class="btn btn-danger float-end ms-1" onclick="deleteAgent(`+agent.ID+`)">
                                <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#trash"/></svg>
                            </button>
                            <button class="btn btn-link float-end ms-1" onclick="editAgent(`+agent.ID+`)" data-bs-toggle="modal" data-bs-target="#editAgent">
                                <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#pencil-square"/></svg>
                            </button>
                        </td>
                    </tr>`)

        table.append(item)
    });

    cache.agents = data.agents
}

function newAgent() {
    let name = $("#agents .card input[name='name']").val()
    let ip = $("#agents .card input[name='ip']").val()
    let port = $("#agents .card input[name='port']").val()
    let psk = $("#agents .card input[name='psk']").val()

    var msg = {
        type: "newagent",
        name: name,
        ip: ip,
        port: parseInt(port),
        psk: psk,
    };
    console.log(msg)
    webSocket.send(JSON.stringify(msg));
}

function editAgent(id) {
    $("#editAgent input").val("")

    let idInput = $("#editAgent input[name='id']")
    let name = $("#editAgent input[name='name']")
    let ip = $("#editAgent input[name='ip']")
    let port = $("#editAgent input[name='port']")
    let psk = $("#editAgent input[name='psk']")

    let data = null
    cache.agents.forEach(agent => {
        if (agent.ID == id) {
            data = agent
        }
    });

    if (data == null) {
        showError("Cannot edit - no agent with that ID.")
        return
    }

    idInput.val(data.ID)
    name.val(data.Name)
    ip.val(data.IP)
    port.val(data.Port)
    psk.val(data.PSK)
}

function updateAgent() {
    let id = $("#editAgent input[name='id']").val()
    let name = $("#editAgent input[name='name']").val()
    let ip = $("#editAgent input[name='ip']").val()
    let port = $("#editAgent input[name='port']").val()
    let psk = $("#editAgent input[name='psk']").val()

    var msg = {
        type: "updateagent",
        id: parseInt(id),
        name: name,
        ip: ip,
        port: parseInt(port),
        psk: psk,
    };
    webSocket.send(JSON.stringify(msg));
}

function deleteAgent(id) {
    var msg = {
        type: "deleteagent",
        id: parseInt(id),
    };
    webSocket.send(JSON.stringify(msg));
}

function gotRepos(data) {
    let table = $("#repos div.new table tbody")
    table.empty()
    data.repos.forEach(repo => {
        let item = $(`<tr>
                        <th scope="row">`+repo.ID+`</th>
                        <td>`+repo.Name+`</td>
                        <td>`+repo.Repo+`</td>
                        <td>
                            <button class="btn btn-danger float-end ms-1" onclick="deleteRepo(`+repo.ID+`)">
                                <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#trash"/></svg>
                            </button>
                            <button class="btn btn-link float-end ms-1" onclick="editRepo(`+repo.ID+`)" data-bs-toggle="modal" data-bs-target="#editRepo">
                                <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#pencil-square"/></svg>
                            </button>
                            <button class="btn btn-link float-end" onclick="getSnapshots(`+repo.ID+`)">
                                <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#search"/></svg>
                            </button>
                        </td>
                    </tr>`)

        table.append(item)
    });

    cache.repos = data.repos
}

function newRepo() {
    let repo = $("#repos .card input[name='repo']").val()
    let password = $("#repos .card input[name='password']").val()
    let settings = $("#repos .card textarea[name='settings']").val().split('\n')

    if (settings.length == 1 && settings[0] == "") {
        settings = []
    }

    var msg = {
        type: "newrepo",
        repo: repo,
        password: password,
        settings: settings,
    };
    webSocket.send(JSON.stringify(msg));
}

function editRepo(id) {
    $("#editRepo input, #editRepo textarea").val("")

    let idInput = $("#editRepo input[name='id']")
    let name = $("#editRepo input[name='name']")
    let repo = $("#editRepo input[name='repo']")
    let password = $("#editRepo input[name='password']")
    let settings = $("#editRepo textarea[name='settings']")

    let data = null
    cache.repos.forEach(repo => {
        if (repo.ID == id) {
            data = repo
        }
    });

    if (data == null) {
        showError("Cannot edit - no repo with that ID.")
        return
    }

    idInput.val(data.ID)
    name.val(data.Name)
    repo.val(data.Repo)
    password.val(data.Password)
    if (data.Settings) {
        settings.val(data.Settings.join("\n"))
    }
}

function updateRepo() {
    let id = $("#editRepo input[name='id']").val()
    let name = $("#editRepo input[name='name']").val()
    let repo = $("#editRepo input[name='repo']").val()
    let password = $("#editRepo input[name='password']").val()
    let settings = $("#editRepo textarea[name='settings']").val().split('\n')

    if (settings.length == 1 && settings[0] == "") {
        settings = []
    }

    var msg = {
        type: "updaterepo",
        id: parseInt(id),
        name: name,
        repo: repo,
        password: password,
        settings: settings,
    };
    webSocket.send(JSON.stringify(msg));
}

function deleteRepo(id) {
    var msg = {
        type: "deleteRepo",
        id: parseInt(id),
    };
    webSocket.send(JSON.stringify(msg));
}

function getSnapshots(id) {
    var msg = {
        type: "getSnapshots",
        id: parseInt(id),
    };
    webSocket.send(JSON.stringify(msg));
}

function gotSnapshots(data) {
    let table = $("#snapshots table tbody")
    table.empty()
    
    data.snapshots.forEach(snapshot => {
        let item = $(`<tr>
                        <th scope="row">`+snapshot.ID+`</th>
                        <td>`+snapshot.Time+`</td>
                        <td>`+snapshot.Host+`</td>
                        <td>`+snapshot.Tags+`</td>
                        <td>`+snapshot.Paths+`</td>
                        <td>
                            <button class="btn btn-danger float-end ms-1" onclick="deleteSnapshot(`+snapshot.ID+`)">
                                <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#trash"/></svg>
                            </button>
                            <button type="button" class="btn btn-link" onclick="prepareRestore('`+snapshot.ID+`')" data-bs-dismiss="modal">
                                <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#download"/></svg>
                            </button>
                        </td>
                    </tr>`)

        table.append(item)
    });
    $(`#snapshots button[data-bs-toggle="modal"][data-bs-target="#snapshots"]`).click()

    cache.snapshots.repo = data.repo
    cache.snapshots.data = data.snapshots
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

function prepareRestore(id) {
    let snapshotID = $(`#restoreSnapshot input[name="id"]`)
    let repoID = $(`#restoreSnapshot input[name="repo"]`)
    let paths = $(`#restoreSnapshot input[name="paths"]`)
    let selectAgent = $(`#restoreSnapshot select[name="agent"]`)
    selectAgent.empty()

    let data = null
    cache.snapshots.data.forEach(snapshot => {
        if (snapshot.ID == id) {
            data = snapshot
        }
    });
    
    if (data == null) {
        showError("No snapshot with that ID in cache.")
        return
    }

    let newOption = $(`<option value="-1" selected>None</option>`)
    selectAgent.append(newOption)

    cache.agents.forEach(agent => {
        let newOption = $(`<option value="`+agent.ID+`">`+agent.Name+`</option>`)
        selectAgent.append(newOption)
    });

    snapshotID.val(id)
    repoID.val(cache.snapshots.repo)
    paths.val(data.Paths)

    $(`#restoreSnapshot button[data-bs-toggle="modal"][data-bs-target="#restoreSnapshot"]`).click()
}

function restoreSnapshot() {
    let repo = $(`#restoreSnapshot input[name="repo"]`).val()
    let snapshot = $(`#restoreSnapshot input[name="id"]`).val()
    let paths = $(`#restoreSnapshot input[name="paths"]`).val()
    let agent = $(`#restoreSnapshot select[name="agent"]`).val()
    let target = $(`#restoreSnapshot input[name="target"]`).val()
    let include = $(`#restoreSnapshot input[name="include"]`).val()
    let exclude = $(`#restoreSnapshot input[name="exclude"]`).val()

    if (agent == -1) {
        showError("You need to select an agent.")
        return
    }

    var msg = {
        type: "restoreSnapshot",
        repo: parseInt(repo),
        snapshot: snapshot,
        paths: paths,
        agent: parseInt(agent),
        target: target,
        include: include,
        exclude: exclude,
    };
    webSocket.send(JSON.stringify(msg));
}