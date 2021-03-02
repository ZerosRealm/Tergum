var webSocket = new WebSocket("ws://127.0.0.1/ws", "1");
var cache = {
    currentPage: "index",
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
            cache.backups = data.backups
            if (cache.currentPage == "backups") {
                gotBackups(data)
            }
            break;
        case "getrepos":
            cache.repos = data.repos
            if (cache.currentPage == "repos") {
                gotRepos(data)
            }
            break
        case "getagents":
            cache.agents = data.agents
            if (cache.currentPage == "agents") {
                gotAgents(data)
            }
            break
        case "getsubscribers":
            cache.subscribers = data.subscribers
            break
        case "getsnapshots":
            if (cache.currentPage == "repos") {
                cache.snapshots.repo = data.repo
                cache.snapshots.data = data.snapshots
                gotSnapshots(data)
            }
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