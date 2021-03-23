$( document ).ready(function() {
    cache.currentPage = "agents"
    refresh()
});

const genRanHex = size => [...Array(size)].map(() => Math.floor(Math.random() * 16).toString(16)).join('');

function generatePSK() {
    let psk = $("#agents .card input[name='psk']")
    psk.val(genRanHex(32))
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
                            <button class="btn btn-danger float-end ms-1" onclick="confirmationBox('Do you want to delete this agent?', () => deleteAgent(`+agent.ID+`))">
                                <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#trash"/></svg>
                            </button>
                            <button class="btn btn-link float-end ms-1" onclick="editAgent(`+agent.ID+`)" data-bs-toggle="modal" data-bs-target="#editAgent">
                                <svg class="bi" width="16" height="16" fill="currentColor"><use xlink:href="css/bootstrap-icons.svg#pencil-square"/></svg>
                            </button>
                        </td>
                    </tr>`)

        table.append(item)
    });
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