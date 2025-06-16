<script lang="ts">
  import MainLayout from "../components/layout/MainLayout.svelte";
  import { useTheme } from "../hooks/useTheme";
  import { getProfile, updateProfile, getBlockedUsers, unblockUser } from "../api/user";
  import { onMount } from "svelte";
  import { toastStore } from "../stores/toastStore";
  import ThemeToggle from "../components/common/ThemeToggle.svelte";

  const { theme } = useTheme();
  $: isDarkMode = $theme === "dark";

  let isLoading = true;
  let isSaving = false;
  let isLoadingBlockedUsers = true;

  const fontSizeOptions = [
    { value: "small", label: "Small" },
    { value: "medium", label: "Medium" },
    { value: "large", label: "Large" }
  ];

  const fontColorOptions = [
    { value: "default", label: "Default" },
    { value: "blue", label: "Blue" },
    { value: "green", label: "Green" },
    { value: "purple", label: "Purple" }
  ];

  const notificationTypes = [
    { id: "like", label: "Like", description: "Notify me when someone likes my post" },
    { id: "repost", label: "Repost", description: "Notify me when someone reposts my post" },
    { id: "follow", label: "Follow", description: "Notify me when someone follows me" },
    { id: "mentions", label: "Mentions", description: "Notify me when someone mentions me" },
    { id: "community", label: "Community", description: "Notify me about community activities" }
  ];

  let fontSize = "medium";
  let fontColor = "default";

  $: textColor = getTextColor(fontColor, isDarkMode);

  function getTextColor(color, isDark) {
    if (color === "blue") {
      return isDark ? "#62a0ea" : "#1a5fb4";
    } else if (color === "green") {
      return isDark ? "#57e389" : "#2a7a30";
    } else if (color === "purple") {
      return isDark ? "#dc8add" : "#613583";
    }
    return isDark ? "var(--text-primary)" : "var(--text-primary)";
  }

  interface BlockedUser {
    id: string;
    username: string;
    display_name?: string;
    name?: string;
    profile_picture_url?: string;
    is_verified?: boolean;
    blocked_at?: string;
  }

  let blockedAccounts: BlockedUser[] = [];
  const blockedUsersPage = 1;
  const blockedUsersLimit = 20;

  let userData = {
    name: "",
    username: "",
    email: "",
    is_private: false,
    notifications: {
      like: true,
      repost: true,
      follow: true,
      mentions: true,
      community: false
    }
  };

  onMount(async () => {
    isLoading = true;

    // Load font preferences from localStorage
    try {
      const storedFontSize = localStorage.getItem("fontSize");
      if (storedFontSize) {
        fontSize = storedFontSize;
        applyFontSize(fontSize);
      }

      const storedFontColor = localStorage.getItem("fontColor");
      if (storedFontColor) {
        fontColor = storedFontColor;
        applyFontColor(fontColor);
      }
    } catch (error) {
      console.error("Error loading font preferences:", error);
    }

    try {
      const response = await getProfile();
      const profile = response.user || (response.data && response.data.user);

      if (profile) {
        userData = {
          name: profile.name || profile.display_name || "",
          username: profile.username || "",
          email: profile.email || "",
          is_private: profile.is_private === true,
          notifications: {
            like: profile.notifications?.like !== false,
            repost: profile.notifications?.repost !== false,
            follow: profile.notifications?.follow !== false,
            mentions: profile.notifications?.mentions !== false,
            community: profile.notifications?.community === true
          }
        };
      }
    } catch (error) {
      console.error("Error loading profile:", error);
      toastStore.showToast("Failed to load profile settings", "error");
    } finally {
      isLoading = false;
    }

    // Load blocked users
    await loadBlockedUsers();
  });

  // Function to load blocked users from API
  async function loadBlockedUsers() {
    isLoadingBlockedUsers = true;
    try {
      blockedAccounts = await getBlockedUsers(blockedUsersPage, blockedUsersLimit);
    } catch (error) {
      console.error("Error loading blocked users:", error);
      toastStore.showToast("Failed to load blocked accounts", "error");
    } finally {
      isLoadingBlockedUsers = false;
    }
  }

  // Apply font size to body
  function applyFontSize(size) {
    if (typeof document !== "undefined") {
      document.documentElement.classList.remove("font-small", "font-medium", "font-large");
      document.documentElement.classList.add(`font-${size}`);
    }
  }

  // Apply font color to body
  function applyFontColor(color) {
    if (typeof document !== "undefined") {
      document.documentElement.classList.remove("text-default", "text-blue", "text-green", "text-purple");
      document.documentElement.classList.add(`text-${color}`);

      // Apply to HTML element as well for better specificity
      const htmlElement = document.querySelector("html");
      if (htmlElement) {
        htmlElement.classList.remove("text-default", "text-blue", "text-green", "text-purple");
        htmlElement.classList.add(`text-${color}`);
      }

      // Apply color directly to key elements
      if (color !== "default") {
        setTimeout(() => {
          const headings = document.querySelectorAll(".settings-section-title, .page-title");
          const labels = document.querySelectorAll(".settings-form-group label, .settings-toggle-label span");

          const currentColor = getTextColor(color, isDarkMode);

          headings.forEach(heading => {
            (heading as HTMLElement).style.color = currentColor;
          });

          labels.forEach(label => {
            (label as HTMLElement).style.color = currentColor;
          });
        }, 10);
      }
    }
  }

  // Handle font size change
  function handleFontSizeChange(event) {
    const newSize = event.target.value;
    fontSize = newSize;
    localStorage.setItem("fontSize", newSize);
    applyFontSize(newSize);
  }

  // Handle font color change
  function handleFontColorChange(event) {
    const newColor = event.target.value;
    fontColor = newColor;
    localStorage.setItem("fontColor", newColor);
    applyFontColor(newColor);

    // Force a refresh to make sure the colors are applied immediately
    setTimeout(() => {
      if (typeof document !== "undefined") {
        // This will trigger style recalculation and apply the new colors
        document.body.style.display = "none";
        // This forces a reflow
        void document.body.offsetHeight;
        document.body.style.display = "";
      }
    }, 0);
  }

  // Handle unblock user - now uses the real API
  async function handleUnblockUser(userId) {
    try {
      const success = await unblockUser(userId);
      if (success) {
        blockedAccounts = blockedAccounts.filter(account => account.id !== userId);
        toastStore.showToast("User unblocked successfully", "success");
      } else {
        toastStore.showToast("Failed to unblock user", "error");
      }
    } catch (error) {
      console.error("Error unblocking user:", error);
      toastStore.showToast("Failed to unblock user", "error");
    }
  }

  // Handle account deactivation
  function handleDeactivateAccount() {
    if (confirm("Are you sure you want to deactivate your account? This action can be reversed by logging in again within 30 days.")) {
      // Would call API in real implementation
      toastStore.showToast("Account deactivation initiated", "info");
    }
  }

  async function handleSaveSettings() {
    isSaving = true;
    try {
      await updateProfile(userData);
      toastStore.showToast("Settings saved successfully", "success");
    } catch (error) {
      console.error("Error saving settings:", error);
      toastStore.showToast("Failed to save settings", "error");
    } finally {
      isSaving = false;
    }
  }
</script>

<MainLayout>
  <div class="page-header {isDarkMode ? "page-header-dark" : ""}">
    <h1 class="page-title" style={fontColor !== "default" ? `color: ${textColor}` : ""}>Settings</h1>
  </div>

  <div class="settings-container {isDarkMode ? "settings-container-dark" : ""}">
    {#if isLoading}
      <div class="settings-loading">
        <div class="settings-loading-spinner"></div>
        <p style={fontColor !== "default" ? `color: ${textColor}` : ""}>Loading settings...</p>
      </div>
    {:else}
      <div class="settings-sections">
        <!-- Security Section -->
        <div class="settings-section {isDarkMode ? "settings-section-dark" : ""}">
          <h2 class="settings-section-title" style={fontColor !== "default" ? `color: ${textColor}` : ""}>Security</h2>

          <div class="settings-toggle-group">
            <div class="settings-toggle-label">
              <span style={fontColor !== "default" ? `color: ${textColor}` : ""}>Private Account</span>
              <small>Only approved followers can see your content</small>
            </div>
            <label class="toggle">
              <input type="checkbox" bind:checked={userData.is_private}>
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>

        <!-- Display Settings Section -->
        <div class="settings-section {isDarkMode ? "settings-section-dark" : ""}">
          <h2 class="settings-section-title" style={fontColor !== "default" ? `color: ${textColor}` : ""}>Display</h2>

          <div class="settings-toggle-group theme-toggle-container">
            <div class="settings-toggle-label">
              <span style={fontColor !== "default" ? `color: ${textColor}` : ""}>Dark mode</span>
              <small>Switch between light and dark theme</small>
            </div>
            <ThemeToggle size="md" />
          </div>

          <div class="settings-form-group">
            <label for="fontSize" style={fontColor !== "default" ? `color: ${textColor}` : ""}>Font Size</label>
            <div class="settings-select-wrapper">
              <select
                id="fontSize"
                class="settings-select {isDarkMode ? "settings-select-dark" : ""}"
                value={fontSize}
                on:change={handleFontSizeChange}
              >
                {#each fontSizeOptions as option}
                  <option value={option.value}>{option.label}</option>
                {/each}
              </select>
            </div>
            <small>Adjust the text size for better readability</small>
          </div>

          <div class="settings-form-group">
            <label for="fontColor" style={fontColor !== "default" ? `color: ${textColor}` : ""}>Font Color</label>
            <div class="settings-select-wrapper">
              <select
                id="fontColor"
                class="settings-select {isDarkMode ? "settings-select-dark" : ""}"
                value={fontColor}
                on:change={handleFontColorChange}
              >
                {#each fontColorOptions as option}
                  <option value={option.value}>{option.label}</option>
                {/each}
              </select>
            </div>
            <small>Customize the text color for your preference</small>
          </div>
        </div>

        <!-- Your Account Section -->
        <div class="settings-section {isDarkMode ? "settings-section-dark" : ""}">
          <h2 class="settings-section-title" style={fontColor !== "default" ? `color: ${textColor}` : ""}>Your Account</h2>

          <div class="account-action-container">
            <div class="account-action-info">
              <h3 style={fontColor !== "default" ? `color: ${textColor}` : ""}>Deactivate Account</h3>
              <p>Temporarily disable your account. You can reactivate it anytime within 30 days by logging in.</p>
            </div>
            <button class="danger-button" on:click={handleDeactivateAccount}>
              Deactivate
            </button>
          </div>
        </div>

        <!-- Blocked Accounts Section -->
        <div class="settings-section {isDarkMode ? "settings-section-dark" : ""}">
          <h2 class="settings-section-title" style={fontColor !== "default" ? `color: ${textColor}` : ""}>Blocked Accounts</h2>

          {#if isLoadingBlockedUsers}
            <div class="settings-loading">
              <div class="settings-loading-spinner"></div>
              <p>Loading blocked accounts...</p>
            </div>
          {:else if blockedAccounts.length === 0}
            <p class="empty-list-message">You haven't blocked any accounts.</p>
          {:else}
            <div class="blocked-accounts-list">
              {#each blockedAccounts as account}
                <div class="blocked-account-item">
                  <div class="account-info">
                    <span class="account-name" style={fontColor !== "default" ? `color: ${textColor}` : ""}>{account.display_name || account.name}</span>
                    <small class="account-username">@{account.username}</small>
                  </div>
                  <button class="unblock-button" on:click={() => handleUnblockUser(account.id)}>
                    Unblock
                  </button>
                </div>
              {/each}
            </div>
          {/if}
        </div>

        <!-- Notification Preferences Section -->
        <div class="settings-section {isDarkMode ? "settings-section-dark" : ""}">
          <h2 class="settings-section-title" style={fontColor !== "default" ? `color: ${textColor}` : ""}>Notification Preferences</h2>

          {#each notificationTypes as notificationType}
            <div class="settings-toggle-group">
              <div class="settings-toggle-label">
                <span style={fontColor !== "default" ? `color: ${textColor}` : ""}>{notificationType.label}</span>
                <small>{notificationType.description}</small>
              </div>
              <label class="toggle">
                <input
                  type="checkbox"
                  bind:checked={userData.notifications[notificationType.id]}
                >
                <span class="toggle-slider"></span>
              </label>
            </div>
          {/each}
        </div>

        <div class="settings-actions">
          <button
            class="settings-save-btn"
            on:click={handleSaveSettings}
            disabled={isSaving}
          >
            {isSaving ? "Saving..." : "Save Settings"}
          </button>
        </div>
      </div>
    {/if}
  </div>
</MainLayout>

<style>
  .settings-container {
    padding: var(--space-4);
  }

  .settings-loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--space-8);
  }

  .settings-loading-spinner {
    width: 40px;
    height: 40px;
    border: 3px solid var(--border-color);
    border-top: 3px solid var(--color-primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: var(--space-4);
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  .settings-sections {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .settings-section {
    background-color: var(--bg-secondary);
    border-radius: var(--radius-lg);
    padding: var(--space-4);
  }

  .settings-section-dark {
    background-color: var(--dark-bg-secondary);
  }

  .settings-section-title {
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-bold);
    margin-bottom: var(--space-4);
    padding-bottom: var(--space-2);
    border-bottom: 1px solid var(--border-color);
  }

  .settings-form-group {
    margin-bottom: var(--space-4);
  }

  .settings-form-group label {
    display: block;
    margin-bottom: var(--space-2);
    font-weight: var(--font-weight-medium);
  }

  .settings-form-group small {
    display: block;
    font-size: var(--font-size-xs);
    color: var(--text-tertiary);
    margin-top: var(--space-1);
  }

  .settings-input,
  .settings-textarea {
    width: 100%;
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    background-color: var(--bg-primary);
    color: var(--text-primary);
    font-size: var(--font-size-base);
  }

  .settings-input-dark,
  .settings-textarea-dark {
    border-color: var(--border-color-dark);
    background-color: var(--bg-primary-dark);
    color: var(--text-primary-dark);
  }

  .settings-input:focus,
  .settings-textarea:focus {
    border-color: var(--color-primary);
    outline: none;
  }

  .settings-textarea {
    resize: vertical;
    min-height: 80px;
  }

  .settings-toggle-group {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--space-4);
  }

  .settings-toggle-group:last-child {
    margin-bottom: 0;
  }

  .settings-toggle-label {
    display: flex;
    flex-direction: column;
  }

  .settings-toggle-label span {
    font-weight: var(--font-weight-medium);
  }

  .settings-toggle-label small {
    font-size: var(--font-size-xs);
    color: var(--text-tertiary);
    margin-top: var(--space-1);
  }

  .toggle {
    position: relative;
    display: inline-block;
    width: 52px;
    height: 26px;
  }

  .toggle input {
    opacity: 0;
    width: 0;
    height: 0;
  }

  .toggle-slider {
    position: absolute;
    cursor: pointer;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: var(--border-color);
    transition: .4s;
    border-radius: 34px;
  }

  .toggle-slider:before {
    position: absolute;
    content: "";
    height: 20px;
    width: 20px;
    left: 3px;
    bottom: 3px;
    background-color: white;
    transition: .4s;
    border-radius: 50%;
  }

  input:checked + .toggle-slider {
    background-color: var(--color-primary);
  }

  input:checked + .toggle-slider:before {
    transform: translateX(26px);
  }

  .theme-toggle-container {
    margin-bottom: var(--space-4);
  }

  .settings-actions {
    margin-top: var(--space-4);
    display: flex;
    justify-content: flex-end;
  }

  .settings-save-btn {
    padding: var(--space-2) var(--space-5);
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: var(--radius-full);
    font-weight: var(--font-weight-bold);
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .settings-save-btn:hover {
    background-color: var(--color-primary-hover);
  }

  .settings-save-btn:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

  .settings-select-wrapper {
    position: relative;
    width: 100%;
  }

  .settings-select {
    width: 100%;
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    background-color: var(--bg-primary);
    color: var(--text-primary);
    font-size: var(--font-size-base);
    appearance: none;
    cursor: pointer;
  }

  .settings-select-dark {
    border-color: var(--border-color-dark);
    background-color: var(--bg-primary-dark);
    color: var(--text-primary-dark);
  }

  .settings-select:focus {
    border-color: var(--color-primary);
    outline: none;
  }

  .settings-select-wrapper::after {
    content: "▼";
    font-size: 0.8em;
    position: absolute;
    right: 12px;
    top: 50%;
    transform: translateY(-50%);
    pointer-events: none;
    color: var(--text-secondary);
  }

  /* Account Deactivation */
  .account-action-container {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-3) 0;
  }

  .account-action-info h3 {
    margin: 0 0 var(--space-1) 0;
    font-size: var(--font-size-base);
    font-weight: var(--font-weight-medium);
  }

  .account-action-info p {
    margin: 0;
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
  }

  .danger-button {
    padding: var(--space-2) var(--space-4);
    background-color: var(--color-danger);
    color: white;
    border: none;
    border-radius: var(--radius-md);
    font-weight: var(--font-weight-medium);
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .danger-button:hover {
    background-color: var(--color-danger-dark);
  }

  /* Blocked Accounts */
  .blocked-accounts-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .blocked-account-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-3);
    border-radius: var(--radius-md);
    background-color: var(--bg-primary);
    border: 1px solid var(--border-color);
  }

  .account-info {
    display: flex;
    flex-direction: column;
  }

  .account-name {
    font-weight: var(--font-weight-medium);
    margin-bottom: var(--space-1);
  }

  .account-username {
    color: var(--text-secondary);
    font-size: var(--font-size-xs);
  }

  .unblock-button {
    padding: var(--space-1) var(--space-3);
    background-color: transparent;
    color: var(--color-primary);
    border: 1px solid var(--color-primary);
    border-radius: var(--radius-md);
    font-size: var(--font-size-sm);
    cursor: pointer;
    transition: all 0.2s;
  }

  .unblock-button:hover {
    background-color: var(--color-primary);
    color: white;
  }

  .empty-list-message {
    color: var(--text-secondary);
    text-align: center;
    padding: var(--space-4);
  }

  /* Responsive adjustments */
  @media (max-width: 768px) {
    .account-action-container,
    .settings-toggle-group {
      flex-direction: column;
      align-items: flex-start;
      gap: var(--space-2);
    }

    .settings-toggle-group .toggle,
    .account-action-container .danger-button {
      margin-top: var(--space-2);
    }
  }
</style>
