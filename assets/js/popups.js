// modals
// import { resetInputStyle, toggleUserInteracted } from "./update_UI_elements.js";
import { activePageElement } from "./main.js";
import {
  applyCustomTheme,
  showThemesClickable,
  pickTheme,
  themeList,
} from "./consoleThemes.js";

export function closePostForm() {
  console.info("closePostForm");
  const formPost = document.querySelector("#form-post");
  const cancelPost = document.querySelector("#cancel-post");
  if (formPost && cancelPost) {
    cancelPost.popoverTargetElement = formPost;
    cancelPost.popoverTargetAction = "hide";
  }
}

const formChannel = document.querySelector("#form-channel");
const cancelChannel = document.querySelector("#cancel-channel");
if (formChannel && cancelChannel) {
  cancelChannel.popoverTargetElement = formChannel;
  cancelChannel.popoverTargetAction = "hide";
}

const formChannelRules = document.querySelector("#form-edit-channel-rules");
const cancelChannelRules = document.querySelector("#cancel-channel-rules");
if (formChannelRules && cancelChannelRules) {
  cancelChannelRules.popoverTargetElement = formChannelRules;
  cancelChannelRules.popoverTargetAction = "hide";
}

document.addEventListener("DOMContentLoaded", () => {
  toggleModals();
  setupChatScrolling();
});

// Scroll chat to bottom when popover opens
function setupChatScrolling() {
  // Get all chat popovers
  const chatPopovers = document.querySelectorAll('[id^="form-chat-"]');

  chatPopovers.forEach((chatPopover) => {
    chatPopover.addEventListener('toggle', (event) => {
      // Check if popover is being opened
      if (event.newState === 'open') {
        const chatContainer = chatPopover.querySelector('.chat-messages-container');
        if (chatContainer) {
          // Use setTimeout to ensure content is fully rendered
          setTimeout(() => {
            chatContainer.scrollTop = chatContainer.scrollHeight;
          }, 0);
        }
      }
    });
  });
}

export function toggleModals() {
  const loginModal = document.querySelector("#container-form-login");

  // Get the <span> element that closes the modal
  const closeLoginModal = loginModal.getElementsByClassName("close")[0];

  // Get the buttons that open the modals
  const openLoginModal = document.querySelector("#btn-open-login-modal");
  const openLoginModalFallback = document.querySelector(
    "#btn-open-login-modal-fallback",
  );

  // open modals
  // TODO refactor the open and close modals
  if (openLoginModal) {
    openLoginModal.addEventListener("click", () => {
      loginModal.style.display = "block";

      // Delay attaching the outside click listener
      setTimeout(() => {
        document.addEventListener("click", handleClickOutsideModals);
      }, 0);
    });

    openLoginModalFallback.addEventListener(
      "click",
      () => (loginModal.style.display = "block"),
    );
  } else {
    console.warn("openLoginModal not found");
  }

  function handleClickOutsideModals(e) {
    console.custom.warn("e.target", e.target);

    // if the target is not the modal itself, which is of class modal-content
    if (e.target === loginModal || e.target === closeLoginModal) {
      console.custom.info("clicked outside modals");

      loginModal.style.display = "none";
      resetInputStyle();
      document.removeEventListener("click", handleClickOutsideModals);
    }
  }
}
