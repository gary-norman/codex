let connection;

let isConnecting = false;  // Flag to track connection state

class Event {
    constructor(type, payload){
        this.type = type;
        this.payload = payload;
    }
}

function routeEvent(event) {
    if (event.type === undefined ) {
        alert("no type field in the event")
    }

    switch(event.type) {
        case "new message":
            console.log("message received", event.payload)
            break;
        default:
            alert("unsupported message type")
    }
}

function sendEvent(eventName, payload) {
    const event = new Event(eventName, payload);
    connection.send(JSON.stringify(event));
}

function connectWebSocket() {
    if (isConnecting || (connection && connection.readyState === WebSocket.OPEN)) {
        return;
    }

    isConnecting = true;

    connection = new WebSocket("ws://localhost:8888/ws");

    connection.onopen = function() {
        console.log("✅ WebSocket OPEN - readyState:", connection.readyState);
        isConnecting = true;
    };

    connection.onmessage = function(evt) {
        const eventData = JSON.parse(evt.data);

        const event = Object.assign(new Event, eventData);

        routeEvent(event); 
    };

    connection.onerror = function(error) {
        console.error("❌ WebSocket ERROR - readyState:", connection.readyState);
        console.error("Error details:", error);
        isConnecting = false;
    };

    connection.onclose = function(event) {
        console.log("❌ WebSocket CLOSED");
        console.log("Close code:", event.code);
        console.log("Close reason:", event.reason);
        console.log("Was clean:", event.wasClean);
        isConnecting = false;
    };
}




document.addEventListener("DOMContentLoaded", () => {
    // Initial connection
    connectWebSocket();



    const sendMessageButton = document.getElementById("sendMessage");
    sendMessageButton.addEventListener("click", sendMessage);
});

export function sendMessage() {
    const messageInput = document.getElementById("messageInput");

    if (messageInput != null) {
        sendEvent("send_message", messageInput.value);
        return
    }


}
