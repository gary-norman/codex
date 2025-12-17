import { sendChatMessage } from './websocket.js';
import { showInlineNotification } from './notifications.js';

export function setupChatFormHandlers() {
    // Find all chat forms
    const chatForms = document.querySelectorAll('form[id^="form-send-chat-"]');

    chatForms.forEach(form => {
        // Extract chatID from form ID (form-send-chat-{UUID})
        const formID = form.id;
        const chatID = formID.replace('form-send-chat-', '');

        // Get the chat popover and title element for notifications
        const chatPopover = document.getElementById(`form-chat-${chatID}`);
        const titleElement = chatPopover ? chatPopover.querySelector('#chat-popover-title') : null;

        // Add submit event listener
        form.addEventListener('submit', async (e) => {
            e.preventDefault();

            // Get message input
            const messageInput = form.querySelector(`#chat-input-${chatID}`);
            if (!messageInput) {
                console.error(`Message input not found for chat ${chatID}`);
                return;
            }

            const message = messageInput.value.trim();
            if (!message) {
                console.warn("Empty message, not sending");
                return;
            }

            // Clear input immediately (optimistic UI)
            const originalValue = messageInput.value;
            messageInput.value = '';

            // Try to send via WebSocket
            const success = sendChatMessage(chatID, message);

            if (!success) {
                // WebSocket failed, restore input and show error
                messageInput.value = originalValue;
                console.error("Failed to send message via WebSocket");

                // Show error notification in chat title
                if (titleElement) {
                    const originalTitle = titleElement.textContent;
                    showInlineNotification(
                        titleElement,
                        originalTitle,
                        'WebSocket not connected',
                        false, // success = false (shows error color)
                        'invisible-notify',
                        3000
                    );
                }
            }
        });
    });

    console.log(`✅ Set up chat handlers for ${chatForms.length} chat forms`);
}

// Setup handlers for "Start New Chat" buttons
export function setupStartChatHandlers() {
    const startChatButtons = document.querySelectorAll('.btn-start-chat');

    startChatButtons.forEach(button => {
        button.addEventListener('click', async (e) => {
            e.preventDefault();

            const buddyID = button.dataset.userId;
            const username = button.dataset.username;

            console.log(`Creating chat with ${username}...`);

            try {
                const response = await fetch('/api/chats/create', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    credentials: 'include',
                    body: JSON.stringify({
                        buddy_id: buddyID
                    })
                });

                if (!response.ok) {
                    console.error('Failed to create chat:', response.statusText);
                    showInlineNotification(
                        document.getElementById('chat-popover-title'),
                        '',
                        'Failed to create chat',
                        false,
                        'invisible-notify',
                        3000
                    );
                    return;
                }

                const data = await response.json();
                console.log('Chat created/found:', data);

                // Reload page to show the new chat
                window.location.reload();

            } catch (error) {
                console.error('Error creating chat:', error);
            }
        });
    });

    console.log(`✅ Set up handlers for ${startChatButtons.length} start chat buttons`);
}

// Initialize on DOM load
document.addEventListener('DOMContentLoaded', () => {
    setupChatFormHandlers();
    setupStartChatHandlers();
});
