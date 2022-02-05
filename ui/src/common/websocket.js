import { writable } from 'svelte/store';

const messageStore = writable({ data: "" });

console.log("Websocket endpoint", WS)
const socket = new WebSocket(WS);

// Connection opened
socket.addEventListener('open', function (event) {
    console.log("Websocket open.");
});

// Listen for messages
socket.addEventListener('message', function (event) {
    messageStore.set({ data: event.data });
});

const sendMessage = (message) => {
	// If we are still connecting, delay message.
	if (socket.readyState == 0) {
		setTimeout(() => {
			sendMessage(message);
		}, 100);
	}
	if (socket.readyState == 1) {
		socket.send(message);
	}
}


export default {
	subscribe: messageStore.subscribe,
	sendMessage
}

