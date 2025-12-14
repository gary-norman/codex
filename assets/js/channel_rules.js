
// INFO was a DOMContentLoaded function
// ------- channel rules -------
export function listenToRules() {
    const addButton = document.querySelector("#add-unsubmitted-rule");
    const submitButton = document.querySelector("#edit-channel-rules-btn");
    const addedRulesWrapper = document.querySelector("#rules-wrapper-added");
    const removedRulesWrapper = document.querySelector("#rules-wrapper-removed");
    const inputField = document.querySelector("#create-unsubmitted-rule");
    const hiddenInput = document.querySelector("#rules-hidden-input");
    const existingRulesContainer = document.querySelector(
        "#rules-wrapper-existing",
    );

    // Early return if required elements don't exist (popover not on this page)
    if (!addButton || !submitButton || !addedRulesWrapper || !removedRulesWrapper ||
        !inputField || !hiddenInput || !existingRulesContainer) {
        return;
    }

    let existingRules = existingRulesContainer.querySelectorAll(
        '[id^="existing-channel-rule-"]',
    );

    existingRules.forEach((element) =>
        element.addEventListener("click", (e) => {
            removeExistingRule(e.target.id);
        }),
    );

    let rulesList = [];
    let addRuleCounter = 0;

    addButton.addEventListener("click", addRule);
    inputField.addEventListener("keydown", handleKeyPress);

    function addRule() {
        const ruleText = inputField.value.trim();

        if (ruleText) {
            const ruleId = `ruleItem-${addRuleCounter++}`;
            createRuleItem(ruleId, ruleText, "add");

            rulesList.push({ id: ruleId, text: ruleText });
            updateHiddenInput();

            inputField.value = "";
        }
    }

    function removeExistingRule(ruleId) {
        const item = document.getElementById(ruleId);
        const ruleText = item.innerText.trim();
        console.log("existing ruleText: ", ruleText);

        if (ruleText) {
            console.log("remove rule: ", ruleId);
            createRuleItem(ruleId, ruleText, "remove");
            rulesList.push({ id: ruleId, text: ruleText });
            updateHiddenInput();
        }
        console.log("removing ", item);
        item.remove();
    }

    function createRuleItem(ruleId, ruleText, process) {
        const ruleItem = document.createElement("li");
        ruleItem.classList.add("rule-item");
        ruleItem.classList.add("flex-space-between");
        ruleItem.id = ruleId;

        // Rule text span
        const ruleTextSpan = document.createElement("span");
        ruleTextSpan.textContent = ruleText;
        ruleTextSpan.classList.add("rule-text");
        if (process === "add") {
            console.log("process add: text = ", ruleText, "ID = ", ruleItem.id);
            ruleTextSpan.addEventListener("click", () => editRule(ruleId, ruleText));
        } else {
            console.log("process remove: text = ", ruleText, "ID = ", ruleItem.id);
            ruleItem.addEventListener("click", () => removeRule(ruleId));
        }

        // Delete button
        const deleteButton = document.createElement("button");
        deleteButton.classList.add("delete-rule-btn");
        deleteButton.classList.add("btn-channel");
        deleteButton.classList.add("btn-sm");
        deleteButton.classList.add("btn-icoonly");
        deleteButton.innerHTML = `<span class="btn-minus" role="contentinfo" aria-description="Remove Rule"></span>`;
        deleteButton.addEventListener("click", (event) => {
            event.stopPropagation(); // Prevent triggering edit on click
            removeRule(ruleId);
        });

        ruleItem.appendChild(ruleTextSpan);
        ruleItem.appendChild(deleteButton);
        console.log("process query: ", process);
        if (process === "remove") {
            removedRulesWrapper.appendChild(ruleItem);
        } else {
            addedRulesWrapper.appendChild(ruleItem);
        }
    }

    function editRule(ruleId, oldText) {
        const ruleItem = document.getElementById(ruleId);
        if (!ruleItem) return;

        const editInput = document.createElement("input");
        editInput.type = "text";
        editInput.value = oldText;
        editInput.classList.add("rule-edit-input");

        ruleItem.replaceChildren(editInput);
        editInput.focus();

        editInput.addEventListener("keydown", (event) => {
            if (event.key === "Enter") {
                saveEditedRule(ruleId, editInput.value);
            } else if (event.key === "Escape") {
                cancelEdit(ruleId, oldText);
            }
        });

        editInput.addEventListener("blur", () =>
            saveEditedRule(ruleId, editInput.value),
        );
    }

    function saveEditedRule(ruleId, newText) {
        if (!newText.trim())
            return cancelEdit(
                ruleId,
                rulesList.find((r) => r.id === ruleId)?.text || "",
            );

        const ruleIndex = rulesList.findIndex((rule) => rule.id === ruleId);
        if (ruleIndex !== -1) {
            rulesList[ruleIndex].text = newText;
            updateHiddenInput();
        }

        createRuleItem(ruleId, newText);
        document.querySelector(".rule-edit-input")?.remove();
    }

    function cancelEdit(ruleId, oldText) {
        createRuleItem(ruleId, oldText);
        document.querySelector(".rule-edit-input")?.remove();
    }

    function removeRule(ruleId) {
        console.log("removeRule: ", ruleId);
        rulesList = rulesList.filter((rule) => rule.id !== ruleId);
        const ruleItem = document.getElementById(ruleId);
        console.log("ruleItem ID: ", ruleItem.id);

        if (ruleItem) ruleItem.remove();

        if (ruleItem.id.startsWith("existing-channel-rule-")) {
            console.log("removing existing rule: ", ruleItem.innerText);
            const ruleTextSpan = document.createElement("span");
            ruleTextSpan.textContent = ruleItem.innerText;
            ruleTextSpan.id = ruleId;
            existingRulesContainer.appendChild(ruleTextSpan);

            existingRules = existingRulesContainer.querySelectorAll(
                '[id^="existing-channel-rule-"]',
            );
            existingRules.forEach((element) =>
                element.addEventListener("click", (e) => {
                    removeExistingRule(e.target.id);
                }),
            );
        }

        updateHiddenInput();
    }

    function updateHiddenInput() {
        hiddenInput.value = JSON.stringify(rulesList);
    }

    function handleKeyPress(event) {
        if (event.key === "Enter") {
            if (event.ctrlKey || event.metaKey) {
                submitButton.click();
            } else {
                addButton.click();
                event.preventDefault();
            }
        }
    }
}