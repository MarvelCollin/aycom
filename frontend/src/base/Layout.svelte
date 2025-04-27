<script>
  import { onMount } from 'svelte';
  import ThemeToggle from '../components/common/ThemeToggle.svelte';
  import { useTheme } from '../hooks/useTheme';
  import Header from '../components/Header.svelte';
  import Footer from '../components/common/Footer.svelte';

  // Access the current theme and toggle function
  const { theme, toggleTheme } = useTheme();

  // Define props
  export let showHeader = true;
  export let showFooter = true;
  export let fullWidth = false;
  export let centerContent = false;

  // Reference to the main container
  let mainContainer;

  onMount(() => {
    // Any initialization logic can go here
    document.documentElement.classList.add('theme-transition');
    
    // Apply initial theme
    document.documentElement.setAttribute('data-theme', $theme);
  });
</script>

<div class="layout-wrapper {$theme === 'dark' ? 'bg-gray-900' : 'bg-gray-50'} min-h-screen">
  {#if showHeader}
    <Header />
  {/if}

  <main 
    bind:this={mainContainer} 
    class="main-container {fullWidth ? 'w-full' : 'container mx-auto px-4'} {centerContent ? 'flex justify-center items-center' : ''}"
  >
    <div class="content-area flex-1">
      <!-- The main content will be injected here -->
      <slot />
    </div>
  </main>

  {#if showFooter}
    <Footer />
  {/if}

  <!-- Theme toggle button - fixed position -->
  <div class="theme-toggle-wrapper pos-fixed bottom-8 right-8 z-50">
    <ThemeToggle />
  </div>
</div>

<style>
  /* Local styles */
  .layout-wrapper {
    display: flex;
    flex-direction: column;
    transition: background-color 0.3s ease;
  }

  .main-container {
    display: flex;
    flex-direction: column;
    flex: 1;
    padding-top: 2rem;
    padding-bottom: 2rem;
  }

  /* Theme transition */
  :global(.theme-transition) {
    transition: background-color 0.3s ease, 
                color 0.3s ease,
                border-color 0.3s ease;
  }

  /* Responsive adjustments */
  @media (min-width: 768px) {
    .main-container {
      padding-top: 3rem;
      padding-bottom: 3rem;
    }
  }

  /* Twitter-style layout support */
  :global(.twitter-layout) {
    display: flex;
    flex-direction: column;
    width: 100%;
  }

  @media (min-width: 768px) {
    :global(.twitter-layout) {
      flex-direction: row;
    }

    :global(.twitter-layout__left) {
      width: 45%;
    }

    :global(.twitter-layout__right) {
      width: 55%;
    }
  }
</style>