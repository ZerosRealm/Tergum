var webSocket = new WebSocket("ws://127.0.0.1/ws");
var cache = {
    currentPage: "index",
    repos: [],
    backups: [],
    agents: [],
    subscribers: [],
    jobs: [],
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
    requestData("getjobs")
}

function showError(msg) {
    msg = capitalize(msg)
    let newAlert = $(`<div class="alert alert-danger" role="alert">
        <span>`+msg+`</span>
        <button style='margin-left: 5px;' type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
    </div>`)
    $("#alerts").append(newAlert)
}

function showSuccess(msg) {
    msg = capitalize(msg)
    let newAlert = $(`<div class="alert alert-success" role="alert">
        <span>`+msg+`</span>
        <button style='margin-left: 5px;' type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
    </div>`)
    $("#alerts").append(newAlert)
}

function hasInvalidFields(form) {
    let foundInvalid = false
    $(form+" input[required]").each(function(i){
        if ($(this).val() == "") {
            foundInvalid = true
            invalidateField(this)
        }
    });

    return foundInvalid
}

function invalidateField(elem) {
    $(elem).addClass("invalid")

    let maxSearch = 5
    let invalidFeedback = null

    let search = $(`.invalid-feedback[for="`+$(elem).attr("name")+`"]`)
    console.log(search)
    if (search.length != 0) {
        invalidFeedback = search
    }

    while (invalidFeedback == null || !$(invalidFeedback).hasClass("invalid-feedback")) {
        if (maxSearch == 0) {
            break
        }

        if (invalidFeedback == null) {
            invalidFeedback = $(elem)[0].nextElementSibling
        } else {
            invalidFeedback = $(invalidFeedback)[0].nextElementSibling
        }
        
        maxSearch--
    }
    $(invalidFeedback).show()
}

function closeModal(modal) {
    $(modal+` input, `+modal+` textarea`).val("")

    $(modal).append($(`<button class="d-none" data-bs-toggle="collapse" data-bs-target="`+modal+`" type="button"></button>`))
    $(modal+` button[data-bs-toggle="collapse"]`).click()
    $(modal+` button[data-bs-toggle="collapse"]`).remove()
}

function resetInvalidForms() {
    $(".form-control.invalid").removeClass("invalid")
    $(".invalid-feedback").hide()
}

var gConfirmFunc = null
function confirmationBox(msg, returnFunc) {
    gConfirmFunc = returnFunc
    let elem = `
    <button type="button" class="btn btn-primary invisible" data-bs-toggle="modal" data-bs-target="#confirmModal"></button>    
    <div class="modal fade" id="confirmModal" tabindex="-1" aria-hidden="true">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Confirmation</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                </div>
                <div class="modal-body">
                    `+msg+`
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">No</button>
                    <button type="button" class="btn btn-primary" onclick="gConfirmFunc()" data-bs-dismiss="modal">Yes</button>
                </div>
            </div>
        </div>
    </div>`
    $("body").append($(elem))
    $(`button[data-bs-toggle="modal"][data-bs-target="#confirmModal"]`).click()
    $(`#confirmModal button`).on("click", function() {
        $("#confirmModal").remove()
        $(`button[data-bs-toggle="modal"][data-bs-target="#confirmModal"]`).remove()
    })
}

const capitalize = (s) => {
  if (typeof s !== 'string') return ''
  return s.charAt(0).toUpperCase() + s.slice(1)
}

function requestData(type) {
    if (webSocket.readyState == 0) {
        return
    }

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
    switch (data.type.toLowerCase()) {
        case "getbackups":
            cache.backups = data.backups
            if (cache.currentPage == "backups") {
                gotBackups(data)
            }
            break;
        case "getjobs":
            cache.jobs = data.jobs
            if (cache.currentPage == "backups") {
                gotJobs(data)
            }
            break;
        case "jobprogress":
            if (cache.currentPage == "backups") {
                gotJobProgress(data)
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