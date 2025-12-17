import { fetchData, changePage, fetchHome } from "./fetch_and_navigate.js";
import { UpdateUI } from "./update_UI_elements.js";
import { fireCalendarListeners } from "./calendar.js";
import { selectActiveFeed, goHome } from "./navigation.js";
import { showMainNotification } from "./notifications.js";
import {
  applyCustomTheme,
  showThemesClickable,
  pickTheme,
  themeList,
} from "./consoleThemes.js";
import "./chat.js";

// Show all themes in console
showThemesClickable();

// Pick theme #3 (example)
pickTheme(3);

// Apply a specific theme manually
// applyCustomTheme("catppuccin", "mocha");

export let activePage, activePageElement;

// Create a new custom event
export const newContentLoaded = new CustomEvent("newContentLoaded");

window.addEventListener("popstate", (event) => {
  if (!event.state || typeof event.state !== "object") {
    // Handle missing or invalid state gracefully
    showMainNotification("Navigation state is missing or invalid.");
    return;
  }
  const { entity, id } = event.state;
  const page = entity + "Page";
  console.custom.info("Popped state:", event.state);
  console.custom.info("entity:", entity);
  console.custom.info("id:", id);
  if (entity === "home") {
    try {
      setActivePage(entity);
      changePage(page);
      return fetchHome();
    } catch (err) {
      showMainNotification("Failed to fetch data.");
      // Optionally log error or handle further
    }
  } else if (entity && id) {
    try {
      setActivePage(entity);
      changePage(page);
      return fetchData(entity, id);
    } catch (err) {
      showMainNotification("Failed to fetch data.");
      // Optionally log error or handle further
    }
  } else {
    // Use a custom error class for HTTP-like errors
    class HttpError extends Error {
      constructor(message, status) {
        super(message);
        this.name = "HttpError";
        this.status = status;
      }
    }
    const error = new HttpError("Invalid URL: entity or id missing", 400);
    showMainNotification(error.message);
    // Optionally log error or handle further
  }
});

document.addEventListener("DOMContentLoaded", (event) => {
  // console.log("fired DOMContentLoaded");
  setActivePage("home");
  UpdateUI();
  selectActiveFeed();
});

document.addEventListener("newContentLoaded", (event) => {
  // console.log("%cfired", expect, "newContentLoaded");
  UpdateUI();
});

export function setActivePage(dest) {
  activePage = dest + "-page";
  activePageElement = document.querySelector(`#${dest}-page`);
  // console.info("%cactivePageElement:", expect, activePageElement);
}

// variables
//user information
// const homePageUserContainer = document.querySelector("#home-page-users");
// if (homePageUserContainer !== null) {
//   homePageUserContainer.addEventListener("click", (e) => {
//     console.log("%cclicked ", expect, e.target);
//     if (e.target.matches(".card")) {
//       navigateToPage("user", e.target);
//     }
//   });
// }
