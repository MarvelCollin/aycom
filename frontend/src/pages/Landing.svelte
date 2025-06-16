<script lang="ts">
  import { useTheme } from "../hooks/useTheme";
  import { onMount } from "svelte";
  import lightLogo from "../assets/logo/light-logo.jpeg";
  import darkLogo from "../assets/logo/dark-logo.jpeg";

  const { theme } = useTheme();

  $: isDarkMode = $theme === "dark";

  onMount(() => {
    document.documentElement.classList.add(isDarkMode ? "dark-theme" : "light-theme");

    return () => {
      document.documentElement.classList.remove(isDarkMode ? "dark-theme" : "light-theme");
    };
  });

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

<div class="aycom-landing-wrapper">
  <div class="aycom-landing-grid">
    <div class="aycom-landing-sidebar">
      <div class="aycom-landing-logo-container">
        {#if isDarkMode}
          <img src={lightLogo} alt="AYCOM Logo" class="aycom-landing-logo-img" />
        {:else}
          <img src={darkLogo} alt="AYCOM Logo" class="aycom-landing-logo-img" />
        {/if}
      </div>
      <div class="aycom-landing-background-pattern"></div>
    </div>

    <div class="aycom-landing-main">
      <div class="aycom-landing-card">
        <div class="aycom-landing-card-header">
          <h1 class="aycom-landing-title">Happening now</h1>
          <h2 class="aycom-landing-subtitle">Join today.</h2>
          <p class="aycom-landing-description">Connect, share, engage.</p>
        </div>

        <div class="aycom-landing-card-actions">
          <a href="/register" class="aycom-landing-primary-button">Create account</a>
          <div class="aycom-landing-divider"></div>
          <div class="aycom-landing-sign-in-section">
            <p class="aycom-landing-sign-in-text">Already have an account?</p>
            <a href="/login" class="aycom-landing-secondary-button">Sign in</a>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  .aycom-landing-wrapper {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 100vh;
    background-color: var(--bg-primary);
    transition: background-color var(--transition-normal);
  }

  .aycom-landing-grid {
    display: flex;
    width: 100%;
    height: 100vh;
    max-width: 1440px;
    margin: 0 auto;
  }

  .aycom-landing-sidebar {
    display: none;
    flex: 1;
    background-color: var(--color-primary);
    justify-content: center;
    align-items: center;
    position: relative;
    overflow: hidden;
  }

  .aycom-landing-logo-container {
    z-index: 10;
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .aycom-landing-logo-img {
    width: 60%;
    height: auto;
    max-width: 300px;
    object-fit: contain;
  }

  .aycom-landing-background-pattern {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-image: url('/assets/auth-pattern.svg');
    background-size: cover;
    opacity: 0.1;
  }

  .aycom-landing-main {
    flex: 1;
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 2rem;
  }

  .aycom-landing-card {
    width: 100%;
    max-width: 380px;
    padding: 2.5rem;
    background-color: var(--bg-primary);
    border-radius: 1rem;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    display: flex;
    flex-direction: column;
    gap: 2.5rem;
  }

  .aycom-landing-card-header {
    text-align: center;
  }

  .aycom-landing-title {
    font-size: 2rem;
    font-weight: 800;
    margin-bottom: 0.75rem;
    color: var(--text-primary);
  }

  .aycom-landing-subtitle {
    font-size: 1.5rem;
    font-weight: 700;
    margin-bottom: 0.5rem;
    color: var(--text-primary);
  }

  .aycom-landing-description {
    font-size: 1rem;
    color: var(--text-secondary);
  }

  .aycom-landing-card-actions {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .aycom-landing-divider {
    height: 1px;
    background-color: var(--border-color);
    margin: 0 auto;
    width: 80%;
  }

  .aycom-landing-primary-button {
    display: block;
    width: 100%;
    padding: 0.875rem;
    border-radius: 9999px;
    background-color: var(--color-primary);
    color: white;
    font-weight: 600;
    font-size: 1rem;
    text-align: center;
    text-decoration: none;
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
    transition: all 0.2s ease;
  }

  .aycom-landing-primary-button:hover {
    background-color: var(--color-primary-hover);
    transform: translateY(-1px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
  }

  .aycom-landing-sign-in-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.75rem;
  }

  .aycom-landing-sign-in-text {
    text-align: center;
    color: var(--text-secondary);
    font-size: 0.9rem;
  }

  .aycom-landing-secondary-button {
    display: block;
    width: 100%;
    padding: 0.875rem;
    border-radius: 9999px;
    background-color: transparent;
    color: var(--color-primary);
    border: 1px solid var(--border-color);
    font-weight: 600;
    font-size: 1rem;
    text-align: center;
    text-decoration: none;
    transition: all 0.2s ease;
  }

  .aycom-landing-secondary-button:hover {
    background-color: var(--bg-hover);
    transform: translateY(-1px);
  }

  :global(.dark-theme) .aycom-landing-wrapper {
    background-color: var(--dark-bg-primary);
  }

  :global(.dark-theme) .aycom-landing-card {
    background-color: var(--dark-bg-secondary);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  }

  :global(.dark-theme) .aycom-landing-divider {
    background-color: var(--dark-border-color);
  }

  :global(.dark-theme) .aycom-landing-secondary-button {
    border-color: var(--dark-border-color);
  }

  :global(.dark-theme) .aycom-landing-secondary-button:hover {
    background-color: var(--dark-hover-bg);
  }

  @media (min-width: 768px) {
    .aycom-landing-sidebar {
      display: flex;
    }

    .aycom-landing-card {
      padding: 3rem;
      box-shadow: 0 6px 16px rgba(0, 0, 0, 0.1);
    }
  }

  @media (max-width: 767px) {
    .aycom-landing-main {
      padding: 1.5rem;
      align-items: center;
    }

    .aycom-landing-card {
      max-width: 100%;
      padding: 2rem;
      margin: 0 1rem;
    }

    .aycom-landing-title {
      font-size: 1.75rem;
    }

    .aycom-landing-subtitle {
      font-size: 1.25rem;
    }
  }
</style>