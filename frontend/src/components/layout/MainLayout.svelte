<script lang="ts">
  import LeftSide from './LeftSide.svelte';
  import RightSide from './RightSide.svelte';
  import Toast from '../common/Toast.svelte';
  import DebugPanel from '../common/DebugPanel.svelte';
  import { useTheme } from '../../hooks/useTheme';
  import type { ITrend, ISuggestedFollow } from '../../interfaces/ISocialMedia';
  import { createEventDispatcher } from 'svelte';
  import { onMount } from 'svelte';
  
  // Icons for mobile navigation
  import HomeIcon from 'svelte-feather-icons/src/icons/HomeIcon.svelte';
  import HashIcon from 'svelte-feather-icons/src/icons/HashIcon.svelte';
  import BellIcon from 'svelte-feather-icons/src/icons/BellIcon.svelte';
  import MailIcon from 'svelte-feather-icons/src/icons/MailIcon.svelte';
  import UserIcon from 'svelte-feather-icons/src/icons/UserIcon.svelte';
  import PlusIcon from 'svelte-feather-icons/src/icons/PlusIcon.svelte';

  export let username = "";
  export let displayName = "";
  export let avatar = "https://secure.gravatar.com/avatar/0?d=mp";
  export let trends: ITrend[] = [];
  export let suggestedFollows: ISuggestedFollow[] = [];
  
  export let showLeftSidebar = true;
  export let showRightSidebar = true;

  // Setup mobile detection
  let isMobile = false;

  onMount(() => {
    // Check if the viewport is mobile on mount
    const checkMobile = () => {
      isMobile = window.innerWidth < 768;
    };
    
    checkMobile();
    window.addEventListener('resize', checkMobile);
    
    return () => {
      window.removeEventListener('resize', checkMobile);
    };
  });

  const { theme } = useTheme();
  $: isDarkMode = $theme === 'dark';

  function handleToggleComposeModal() {
    dispatch('toggleComposeModal');
  }

  // Get the current path for active link styling
  let currentPath = '';
  onMount(() => {
    currentPath = window.location.pathname;
  });

  const dispatch = createEventDispatcher();
</script>

<div class="app-container {isDarkMode ? 'app-container-dark dark-theme' : ''}">
  <div class="app-layout">
    {#if showLeftSidebar}
      <LeftSide 
        {username}
        {displayName}
        {avatar}
        on:toggleComposeModal={handleToggleComposeModal}
      />
    {/if}
    
    <main class="main-content {isDarkMode ? 'main-content-dark' : ''}">
      <slot></slot>
    </main>
    
    {#if showRightSidebar}
      <div class="widgets-container">
        <RightSide 
          {isDarkMode}
        />
      </div>
    {/if}
  </div>
  
  <!-- Mobile navigation bar for smaller screens -->
  {#if isMobile}
    <nav class="mobile-nav {isDarkMode ? 'mobile-nav-dark' : ''}">
      <a href="/feed" class="mobile-nav-item {currentPath === '/feed' ? 'active' : ''}">
        <div class="mobile-nav-icon">
          <HomeIcon size="20" />
        </div>
      </a>
      <a href="/explore" class="mobile-nav-item {currentPath === '/explore' ? 'active' : ''}">
        <div class="mobile-nav-icon">
          <HashIcon size="20" />
        </div>
      </a>
      <button 
        class="mobile-compose-btn"
        on:click={handleToggleComposeModal}
        aria-label="Compose new post"
      >
        <PlusIcon size="20" />
      </button>
      <a href="/notifications" class="mobile-nav-item {currentPath === '/notifications' ? 'active' : ''}">
        <div class="mobile-nav-icon">
          <BellIcon size="20" />
        </div>
      </a>
      <a href="/profile" class="mobile-nav-item {currentPath === '/profile' ? 'active' : ''}">
        <div class="mobile-nav-icon">
          <UserIcon size="20" />
        </div>
      </a>
    </nav>
  {/if}
  
  <Toast />
  <DebugPanel />
</div>

<style>
  /* Mobile compose button */
  .mobile-compose-btn {
    background-color: var(--color-primary);
    color: white;
    border: none;
    width: 44px;
    height: 44px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
    margin-top: -20px;
    transition: background-color 0.2s;
    cursor: pointer;
  }
  
  .mobile-compose-btn:hover,
  .mobile-compose-btn:active {
    background-color: var(--color-primary-hover);
  }
</style>