import {
  showMainNotification,
  showInlineNotification,
} from "./notifications.js";
import {
  applyCustomTheme,
  showThemesClickable,
  pickTheme,
  themeList,
} from "./consoleThemes.js";
import {connectWebSocket} from "./websocket.js";

// registration form
const regForm = document.querySelector("#form-register");
const regPass = document.querySelector("#register_password");
const liValidNum = document.querySelector("#li-valid-num");
const liValidUpper = document.querySelector("#li-valid-upper");
const liValidLower = document.querySelector("#li-valid-lower");
const liValid8 = document.querySelector("#li-valid-8");
const regPassRpt = document.querySelector("#register_password-rpt");
const validList = regForm.querySelector("ul");

// login/register forms
const notifierLogin = document.querySelector("#login-title");
const notifierRegister = document.querySelector("#register-title");
const formLogin = document.querySelector("#form-login");
const formRegister = document.querySelector("#form-register");
const formForgot = document.querySelector("#form-forgot");
let forgotVisible = false;

// login/register buttons
// TODO overhaul the naming of these buttons
const registerForm = document.querySelector("#register-form");
const loginForm = document.querySelector("#login-form");
const logoutFormButton = document.querySelector("#logout");
const btnsLogin = document.querySelectorAll('[id^="btn_login-"]');
const btnsRegister = document.querySelectorAll('[id^="btn_register-"]');
const btnsForgot = document.querySelector("#btn_forgot");

// login/register modal
const loginHeader = document.querySelector("#modal-header-logreg");

// toggle login and register forms
function toggleLoginRegistration(target) {
  if (target === "btn_login-1" || target === "btn_login-2") {
    loginHeader.innerText = "sign in to codex";
  }
  if (target === "btn_register-1" || target === "btn_register-2") {
    loginHeader.innerText = "register for codex";
  }
  if (target === "btn_login-2") {
    console.log("btn-log-2");
    formLogin.classList.remove("display-off");
    formForgot.classList.add("display-off");
    forgotVisible = false;
  } else if (target === "btn_register-2") {
    console.log("btn-reg-2");
    formRegister.classList.remove("display-off");
    formForgot.classList.add("display-off");
    forgotVisible = false;
  } else {
    console.log("Toggling login and register forms");
    formLogin.classList.toggle("display-off");
    formRegister.classList.toggle("display-off");
  }
}
// toggle forgot password form
function forgot() {
  formLogin.classList.add("display-off");
  formRegister.classList.add("display-off");
  formForgot.classList.remove("display-off");
  loginHeader.innerText = "reset password";
  forgotVisible = true;
}
// validate each requirement of the password
function validatePass() {
  if (regPass.value.match(/[0-9]/g)) {
    liValidNum.classList.add("li-valid");
  }
  if (regPass.value.match(/[A-Z]/g)) {
    liValidUpper.classList.add("li-valid");
  }
  if (regPass.value.match(/[a-z]/g)) {
    liValidLower.classList.add("li-valid");
  }
  if (regPass.value.length >= 8) {
    liValid8.classList.add("li-valid");
  }
  if (!regPass.value.match(/[0-9]/g)) {
    liValidNum.classList.remove("li-valid");
  }
  if (!regPass.value.match(/[A-Z]/g)) {
    liValidUpper.classList.remove("li-valid");
  }
  if (!regPass.value.match(/[a-z]/g)) {
    liValidLower.classList.remove("li-valid");
  }
  if (regPass.value.length < 8) {
    liValid8.classList.remove("li-valid");
  }
}
// confirm password validation
function confirmPass() {
  if (regPassRpt.value !== regPass.value || regPassRpt.value === "") {
    return regPassRpt.classList.add("pass-nomatch");
  }
  regPassRpt.classList.remove("pass-nomatch");
}

// retrieve the csrf_token cookie and explicitly set the X-CSRF-Token header in requests
export function getCSRFToken() {
  const match = document.cookie
    .split("; ")
    .find((row) => row.startsWith("csrf_token"));
  if (!match) {
    console.warn("CSRF token not found in cookies");
    return null;
  }
  return match.substring("csrf_token=".length);
}

// login / register / forgot
btnsLogin.forEach((button) =>
  button.addEventListener("click", (e) => toggleLoginRegistration(e.target.id)),
);
btnsRegister.forEach((button) =>
  button.addEventListener("click", (e) => toggleLoginRegistration(e.target.id)),
);
btnsForgot.addEventListener("click", forgot);
//
// TODO get these working

// validate password
regPass.addEventListener("input", validatePass);
// check passwords match
regPassRpt.addEventListener("focusout", confirmPass);
// reverse the order of the password validation list delay
regPass.addEventListener("focusin", () => {
  setTimeout(() => {
    validList.classList.remove("ul-forwards");
    validList.classList.add("ul-reverse");
  }, 1000);
});
regPass.addEventListener("focusout", () => {
  setTimeout(() => {
    validList.classList.remove("ul-reverse");
    validList.classList.add("ul-forwards");
  }, 1000);
});

// --- register ---
if (registerForm) {
  registerForm.addEventListener("submit", async function (event) {
    event.preventDefault(); // Prevent the default form submission
    const form = event.target;
    const formData = new FormData(form); // Collect form data
    const registerUser = formData.get("register_user");
    const registerEmail = formData.get("register_email");
    const registerPassword = formData.get("register_password");
    const registerPasswordRpt = formData.get("register_password-rpt");

    if (!registerUser) {
      showInlineNotification(
        notifierRegister,
        "",
        "please choose a username (5-16 characters)",
        false,
        "dummy",
      );
      return;
    }

    if (!registerEmail) {
      showInlineNotification(
        notifierRegister,
        "",
        "please enter a valid email address",
        false,
        "dummy",
      );
      return;
    }

    if (!registerPassword) {
      showInlineNotification(
        notifierRegister,
        "",
        "please choose a password (min. 8 characters, 1 uppercase, 1 lowercase, 1 number)",
        false,
        "dummy",
      );
      return;
    }

    if (!registerPasswordRpt) {
      showInlineNotification(
        notifierRegister,
        "",
        "please repeat the password",
        false,
        "dummy",
      );
      return;
    }

    try {
      const response = await fetch("/register", {
        method: "POST",
        // headers: {
        //   "Content-Type": "application/json",
        // },
        body: formData,
        // body: JSON.stringify({
        //   register_user: document.getElementById("register_user").value,
        //   register_email: document.getElementById("register_email").value,
        //   register_password: document.getElementById("register_password").value,
        // }),
        cache: "no-store",
      });

      const data = await response.json();
      console.custom.info("data:", data);

      if (!response.ok || data.message === "registration failed!") {
        showInlineNotification(
          notifierRegister,
          "",
          data.message,
          false,
          "dummy",
        );
        return;
      }
      showInlineNotification(notifierRegister, "", data.message, true, "dummy");
      setTimeout(() => {
        window.location.href = "/";
      }, 2000);
    } catch (error) {
      console.custom.error("Registration failed:", error);
      showMainNotification("An error occured during registration.");
    }
  });
}

// --- login ---
if (loginForm) {
  loginForm.addEventListener("submit", async function (event) {
    event.preventDefault();

    const usernameInput = document.getElementById("loginFormUser");
    const passwordInput = document.getElementById("loginFormPassword");
    const username = usernameInput.value.trim();
    const password = passwordInput.value;

    if (!username || !password) {
      showInlineNotification(
        notifierLogin,
        "",
        "Username and password required.",
        false,
        "dummy",
      );
      return;
    }

    const csrfToken = getCSRFToken();

    try {
      const response = await fetch("/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "x-csrf-token": csrfToken,
        },
        body: JSON.stringify({ username, password }),
        cache: "no-store",
      });

      const data = await response.json();

      const errorMessages = new Set([
        "incorrect password",
        "user not found",
        "failed to create cookies",
      ]);

      if (!response.ok || errorMessages.has(data.message)) {
        showInlineNotification(
          notifierLogin,
          "",
          data.message || "Login failed.",
          false,
          "dummy",
        );
        return;
      }

      showInlineNotification(notifierLogin, "", data.message, true, "dummy");

      // Connect WebSocket with OTP
      if (data?.otp) {
        connectWebSocket(data.otp);
      }

      setTimeout(() => {
        window.location.href = "/";
      }, 2000);
    } catch (error) {
      console.error("Login failed:", error);
      showMainNotification("An error occurred during login.");
    }
  });
}

// --- logout ---
if (logoutFormButton) {
  logoutFormButton.addEventListener("click", function (event) {
    event.preventDefault();

    fetch("/logout", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({}),
      cache: "no-store",
    })
      .then((response) => {
        if (response.ok) {
          return response.json(); // Parse JSON response
        } else {
          throw new Error("Logout failed.");
        }
      })
      .then((data) => {
        // Show notification based on the response from the server
        showMainNotification(data.message);
        // Optional: Redirect the user after showing the notification
        setTimeout(() => {
          window.location.href = "/"; // Replace with your desired location
        }, 3500);
      })
      .catch((error) => {
        console.error("Error:", error);
        showMainNotification("An error occurred during logout.");
      });
  });
}

const passwordInput = document.getElementById("register_password");
const togglePasswordButton = document.getElementById("password_toggle");
const passwordInputRpt = document.getElementById("register_password-rpt");
const togglePasswordButtonRpt = document.getElementById("password_toggle-rpt");

// Map buttons to their corresponding inputs for cleaner lookups
const buttonInputMap = new Map([
  [togglePasswordButton, passwordInput],
  [togglePasswordButtonRpt, passwordInputRpt],
]);

let visibilityTimeouts = new WeakMap(); // Track timeouts per button

[togglePasswordButton, togglePasswordButtonRpt].forEach((button) => {
  button.addEventListener("click", () => {
    const input = buttonInputMap.get(button);
    const isVisible = input.type === "text";
    const newType = isVisible ? "password" : "text";

    input.type = newType;
    button.setAttribute("aria-pressed", !isVisible);
    button.setAttribute(
      "aria-label",
      isVisible ? "Show password" : "Hide password",
    );

    // Clear existing timeout for this button
    clearTimeout(visibilityTimeouts.get(button));

    if (!isVisible) {
      const timeout = setTimeout(() => {
        input.type = "password";
        button.setAttribute("aria-pressed", "false");
        button.setAttribute("aria-label", "Show password");
      }, 5000);

      visibilityTimeouts.set(button, timeout);
    }
  });
});
