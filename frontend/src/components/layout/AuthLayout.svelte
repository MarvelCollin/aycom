<script lang="ts">
  import { useTheme } from "../../hooks/useTheme";
  import Logo from "../common/Logo.svelte";
  import Toast from "../common/Toast.svelte";
  import { onMount } from "svelte";
  import lightLogo from "../../assets/logo/light-logo.jpeg";
  import darkLogo from "../../assets/logo/dark-logo.jpeg";

  export let title = "";
  export let showLogo = true;
  export let showCloseButton = false;
  export let showBackButton = false;
  export let onBack = () => {};

  // Get theme store
  const { theme } = useTheme();

  // Reactive declaration to update isDarkMode when theme changes
  $: isDarkMode = $theme === "dark";

  onMount(() => {
    // Apply theme class to document when component mounts
    document.documentElement.classList.add(isDarkMode ? "dark-theme" : "light-theme");

    return () => {
      // Cleanup when component is destroyed
      document.documentElement.classList.remove(isDarkMode ? "dark-theme" : "light-theme");
    };
  });

  // Update theme class when isDarkMode changes
  $: {
    if (typeof document !== "undefined") {
      if (isDarkMode) {
        document.documentElement.classList.add("dark-theme");
        document.documentElement.classList.remove("light-theme");
      } else {
        document.documentElement.classList.add("light-theme");
        document.documentElement.classList.remove("dark-theme");
      }
    }
  }
</script>

<!-- Render the Toast component here -->
<Toast />

<div class="auth-container {isDarkMode ? "auth-container-dark" : ""}">
  <div class="auth-left">
    <div class="auth-left-logo">
      {#if isDarkMode}
        <img src={lightLogo} alt="AYCOM Logo" class="auth-logo-image" />
      {:else}
        <img src={darkLogo} alt="AYCOM Logo" class="auth-logo-image" />
      {/if}
    </div>
    <div class="auth-left-bg"></div>
  </div>

  <div class="auth-right aycom-auth-scroll-container">
    <div class="auth-form">
      <div class="auth-header">
        {#if showBackButton}
          <button
            class="auth-back-button"
            on:click={onBack}
            data-cy="back-button"
            aria-label="Go back"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
          </button>
        {/if}

        {#if showLogo}
          <div class="auth-logo">
            {#if isDarkMode}
              <img src={lightLogo} alt="AYCOM Logo" class="auth-header-logo-image" />
            {:else}
              <img src={darkLogo} alt="AYCOM Logo" class="auth-header-logo-image" />
            {/if}
          </div>
        {/if}

        {#if title}
          <h1 class="auth-title" data-cy="page-title">{title}</h1>
        {/if}
      </div>

      <div class="auth-scrollable-content">
        <slot />
      </div>
    </div>
  </div>
</div>

<style>
  .auth-left-logo {
    z-index: 10;
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .auth-logo-image {
    width: 60%;
    height: auto;
    max-width: 400px;
  }

  .auth-header-logo-image {
    width: 40px;
    height: 40px;
    object-fit: contain;
  }

  .auth-back-button {
    position: absolute;
    top: var(--space-4);
    left: var(--space-4);
    color: var(--color-primary);
    background: transparent;
    border: none;
    cursor: pointer;
    border-radius: 50%;
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background-color var(--transition-fast);
  }

  .auth-back-button:hover {
    background-color: var(--bg-hover);
  }

  /* Global styles for auth components */
  :global(.auth-btn) {
    width: 100%;
    padding: var(--space-3);
    border-radius: var(--radius-full);
    font-weight: 600;
    transition: background-color var(--transition-fast);
  }

  :global(.auth-btn-primary) {
    background-color: var(--color-primary);
    color: white;
  }

  :global(.auth-btn-primary:hover) {
    background-color: var(--color-primary-hover);
  }

  :global(.auth-btn-secondary) {
    background-color: transparent;
    color: var(--color-primary);
    border: 1px solid var(--border-color);
  }

  :global(.auth-btn-secondary:hover) {
    background-color: var(--bg-hover);
  }

  :global(.dark-theme .auth-btn-secondary) {
    border-color: var(--dark-border-color);
  }

  :global(.dark-theme .auth-btn-secondary:hover) {
    background-color: var(--dark-hover-bg);
  }
</style>