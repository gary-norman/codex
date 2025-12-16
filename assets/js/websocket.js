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
    if (!connection || connection.readyState !== WebSocket.OPEN) {
        console.error("WebSocket not connected!");
        alert("WebSocket not connected. Please wait or refresh the page.");
        return;
    }

    const event = new Event(eventName, payload);
    connection.send(JSON.stringify(event));
    console.log("Event sent:", event);
}

export function connectWebSocket(otp) {
    if (isConnecting || (connection && connection.readyState === WebSocket.OPEN)) {
        console.log("⚠️ Already connecting or connected");
        return;
    }

    isConnecting = true;

    const wsUrl = `ws://localhost:8888/ws?otp=${otp}`;
    console.log("🚀 Attempting to connect to:", wsUrl);

    connection = new WebSocket(wsUrl);

    connection.onopen = function() {
        console.log("✅ WebSocket OPEN - readyState:", connection.readyState);
        isConnecting = false;
    };

    connection.onmessage = function(evt) {
        console.log("📨 Message received:", evt.data);
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
        connection = null;
    };
}

// NEW: Function to initialize WebSocket for already-logged-in users
export async function initWebSocket() {
    try {
        console.log("🔑 Fetching OTP for authenticated user...");

        const response = await fetch('/api/ws-otp', {
            credentials: 'include', // Send cookies
            headers: {
                'Accept': 'application/json'
            }
        });

        console.log("📥 OTP fetch response status:", response.status);

        if (!response.ok) {
            if (response.status === 401) {
                console.log("ℹ️ User not authenticated, skipping WebSocket");
            } else {
                console.error("❌ Failed to fetch OTP:", response.status);
            }
            return;
        }

        const data = await response.json();
        console.log("✅ OTP received:", data.otp);
        connectWebSocket(data.otp);

    } catch (error) {
        console.error("❌ Error fetching OTP:", error);
    }
}

// Call on page load for already-logged-in users
document.addEventListener('DOMContentLoaded', () => {
    console.log("🔄 Page loaded, checking for existing session...");
    void initWebSocket();
});

// Existing sendMessage function
document.addEventListener("DOMContentLoaded", () => {
    const sendMessageButton = document.getElementById("sendMessage");
    if (sendMessageButton) {
        sendMessageButton.addEventListener("click", sendMessage);
    }
});

export function sendMessage() {
    const messageInput = document.getElementById("messageInput");

    if (messageInput != null) {
        sendEvent("send_message", messageInput.value);
        return;
    }
}