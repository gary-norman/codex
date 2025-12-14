let socket;

let isConnecting = false;  // Flag to track connection state

function connectWebSocket() {
    if (isConnecting || (socket && socket.readyState === WebSocket.OPEN)) {
        return;
    }

    isConnecting = true;
    console.log("Attempting to connect to ws://localhost:8888/ws");

    socket = new WebSocket("ws://localhost:8888/ws");

    socket.onopen = function() {
        console.log("✅ WebSocket OPEN - readyState:", socket.readyState);
        isConnecting = false;
    };

    socket.onmessage = function(event) {
        console.log("📨 Message received:", event.data);
        const outputMessage = document.getElementById("outputMessage");
        outputMessage.innerHTML += `<p>${event.data}</p>`;
    };

    socket.onerror = function(error) {
        console.error("❌ WebSocket ERROR - readyState:", socket.readyState);
        console.error("Error details:", error);
        isConnecting = false;
    };

    socket.onclose = function(event) {
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
    const message = messageInput.value;

    console.log("sent!")

    // Check if WebSocket is open before sending
    if (socket.readyState === WebSocket.OPEN) {
        socket.send(message);
        messageInput.value = "";
    } else {
        console.error("WebSocket is not open. State:", socket.readyState);
        alert("Connection is closed. Reconnecting...");
        // Optionally reconnect
        connectWebSocket();
    }
}
