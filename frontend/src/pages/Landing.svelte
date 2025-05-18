<script lang="ts">
  import { useTheme } from '../hooks/useTheme';
  import { onMount } from 'svelte';
  import lightLogo from '../assets/logo/light-logo.jpeg';
  import darkLogo from '../assets/logo/dark-logo.jpeg';
  
  // Get theme store
  const { theme } = useTheme();
  
  // Reactive declaration to update isDarkMode when theme changes
  $: isDarkMode = $theme === 'dark';
  
  onMount(() => {
    // Apply theme class to document when component mounts
    document.documentElement.classList.add(isDarkMode ? 'dark-theme' : 'light-theme');
    
    return () => {
      // Cleanup when component is destroyed
      document.documentElement.classList.remove(isDarkMode ? 'dark-theme' : 'light-theme');
    };
  });
  
  // Update theme class when isDarkMode changes
  $: {
    if (typeof document !== 'undefined') {
      if (isDarkMode) {
        document.documentElement.classList.add('dark-theme');
        document.documentElement.classList.remove('light-theme');
      } else {
        document.documentElement.classList.add('light-theme');
        document.documentElement.classList.remove('dark-theme');
      }
    }
  }
</script>

<div class="landing-container {isDarkMode ? 'landing-container-dark' : ''}">
  <div class="landing-content">
    <div class="landing-left">
      <div class="landing-logo">
        {#if isDarkMode}
          <img src={lightLogo} alt="AYCOM Logo" class="logo-image" />
        {:else}
          <img src={darkLogo} alt="AYCOM Logo" class="logo-image" />
        {/if}
      </div>
      <div class="landing-bg"></div>
    </div>
    
    <div class="landing-right">
      <div class="landing-form">
        <div class="landing-header">
          <h1 class="landing-title">Happening now</h1>
          <p class="landing-subtitle">Join today.</p>
          <p class="landing-text">Connect, share, engage.</p>
        </div>
        
        <div class="landing-buttons">
          <a href="/register" class="landing-btn landing-btn-primary">Create account</a>
          <p class="landing-account-text">Already have an account?</p>
          <a href="/login" class="landing-btn landing-btn-secondary">Sign in</a>
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  /* Landing page styling */
  .landing-container {
    display: flex;
    min-height: 100vh;
    background-color: var(--bg-primary);
    transition: background-color var(--transition-normal);
  }

  .landing-container-dark {
    background-color: var(--dark-bg-primary);
  }

  .landing-content {
    display: flex;
    width: 100%;
    height: 100vh;
  }

  /* Left side with logo/branding */
  .landing-left {
    display: none;
    flex: 1;
    background-color: var(--color-primary);
    justify-content: center;
    align-items: center;
    position: relative;
    overflow: hidden;
  }

  .landing-logo {
    z-index: 10;
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .logo-image {
    width: 60%;
    height: auto;
    max-width: 400px;
  }

  .landing-bg {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-image: url('/assets/auth-pattern.svg');
    background-size: cover;
    opacity: 0.1;
  }

  /* Right side with content - IMPROVED */
  .landing-right {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    padding: var(--space-4);
    width: 100%;
    max-width: 100%;
    margin: 0 auto;
    position: relative;
    overflow-y: auto;
  }

  .landing-form {
    width: 100%;
    max-width: 400px;
    margin: 0 auto;
    padding: var(--space-4);
    background-color: var(--bg-primary);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-sm);
    transition: background-color var(--transition-normal), box-shadow var(--transition-normal);
  }

  .landing-header {
    margin-bottom: var(--space-8);
  }

  .landing-title {
    font-size: var(--font-size-4xl);
    font-weight: 800;
    margin-bottom: var(--space-4);
    color: var(--text-primary);
    transition: color var(--transition-normal);
  }

  .landing-subtitle {
    font-size: var(--font-size-2xl);
    font-weight: 700;
    margin-bottom: var(--space-2);
    color: var(--text-primary);
    transition: color var(--transition-normal);
  }

  .landing-text {
    color: var(--text-secondary);
    margin-bottom: var(--space-6);
    transition: color var(--transition-normal);
  }

  /* Button styles */
  .landing-buttons {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .landing-btn {
    display: block;
    width: 100%;
    padding: var(--space-3);
    border-radius: var(--radius-full);
    font-weight: 600;
    font-size: var(--font-size-md);
    text-align: center;
    cursor: pointer;
    transition: background-color var(--transition-fast), border-color var(--transition-fast);
    text-decoration: none;
  }

  .landing-btn-primary {
    background-color: var(--color-primary);
    color: white;
    border: none;
  }

  .landing-btn-primary:hover {
    background-color: var(--color-primary-hover);
  }

  .landing-btn-secondary {
    background-color: transparent;
    color: var(--color-primary);
    border: 1px solid var(--border-color);
    transition: background-color var(--transition-fast), border-color var(--transition-fast);
  }

  .landing-btn-secondary:hover {
    background-color: var(--bg-hover);
  }

  .landing-account-text {
    text-align: center;
    color: var(--text-secondary);
    margin-bottom: var(--space-2);
    transition: color var(--transition-normal);
  }

  /* Dark mode specific adjustments */
  :global(.dark-theme) .landing-form {
    background-color: var(--dark-bg-secondary);
    box-shadow: var(--shadow-md-dark);
  }

  :global(.dark-theme) .landing-btn-secondary {
    border-color: var(--dark-border-color);
  }

  :global(.dark-theme) .landing-btn-secondary:hover {
    background-color: var(--dark-hover-bg);
  }

  /* Responsive styles */
  @media (min-width: 768px) {
    .landing-left {
      display: flex;
    }
    
    .landing-right {
      max-width: 50%;
      padding: var(--space-6);
    }
    
    .landing-form {
      padding: var(--space-6);
      box-shadow: var(--shadow-md);
    }
  }

  @media (max-width: 767px) {
    .landing-right {
      padding: var(--space-4);
    }
    
    .landing-form {
      max-width: 100%;
      box-shadow: none;
      padding: var(--space-2);
    }
    
    .landing-title {
      font-size: var(--font-size-3xl);
    }
    
    .landing-subtitle {
      font-size: var(--font-size-xl);
    }
  }
</style> 