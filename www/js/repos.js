$( document ).ready(function() {
    cache.currentPage = "repos"
});

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
}

function newRepo() {
    let name = $("#repos .card input[name='name']").val()
    let repo = $("#repos .card input[name='repo']").val()
    let password = $("#repos .card input[name='password']").val()
    let settings = $("#repos .card textarea[name='settings']").val().split('\n')

    if (settings.length == 1 && settings[0] == "") {
        settings = []
    }

    var msg = {
        type: "newrepo",
        name: name,
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