<script lang="ts">
  import LeftSide from './LeftSide.svelte';
  import RightSide from './RightSide.svelte';
  import Toast from '../common/Toast.svelte';
  import DebugPanel from '../common/DebugPanel.svelte';
  import ComposeTweetModal from '../social/ComposeTweetModal.svelte';
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
  import MenuIcon from 'svelte-feather-icons/src/icons/MenuIcon.svelte';
  import SearchIcon from 'svelte-feather-icons/src/icons/SearchIcon.svelte';
  import XIcon from 'svelte-feather-icons/src/icons/XIcon.svelte';

  export let username = "";
  export let displayName = "";
  export let avatar = "https://secure.gravatar.com/avatar/0?d=mp";
  export let trends: ITrend[] = [];
  export let suggestedFollows: ISuggestedFollow[] = [];
  
  export let showLeftSidebar = true;
  export let showRightSidebar = true;
  export let pageTitle = "";

  // Setup viewport detection
  let isMobile = false;
  let isTablet = false;
  let isSmallDesktop = false;
  let windowWidth = 0;
  let showComposeModal = false;
  let showMobileMenu = false;
  let showSearchBar = false;
  let searchQuery = '';

  onMount(() => {
    // Check viewport size on mount and on resize
    const checkViewport = () => {
      windowWidth = window.innerWidth;
      isMobile = windowWidth < 768;
      isTablet = windowWidth >= 768 && windowWidth < 992;
      isSmallDesktop = windowWidth >= 992 && windowWidth < 1080;
    };
    
    checkViewport();
    window.addEventListener('resize', checkViewport);
    
    return () => {
      window.removeEventListener('resize', checkViewport);
    };
  });

  const { theme } = useTheme();
  $: isDarkMode = $theme === 'dark';

  function handleToggleComposeModal() {
    showComposeModal = !showComposeModal;
    dispatch('toggleComposeModal');
  }

  function handleNewPost(event) {
    showComposeModal = false;
    dispatch('posted', event.detail);
  }
  
  function toggleMobileMenu() {
    showMobileMenu = !showMobileMenu;
  }
  
  function closeMobileMenu() {
    showMobileMenu = false;
  }
  
  function toggleSearchBar() {
    showSearchBar = !showSearchBar;
    if (showSearchBar) {
      setTimeout(() => {
        document.getElementById('mobile-search-input')?.focus();
      }, 100);
    }
  }
  
  function handleSearch(e) {
    if (e.key === 'Enter' && searchQuery.trim()) {
      window.location.href = `/explore?q=${encodeURIComponent(searchQuery.trim())}`;
    }
  }
  
  function clearSearch() {
    searchQuery = '';
    document.getElementById('mobile-search-input')?.focus();
  }

  // Get the current path for active link styling
  let currentPath = '';
  onMount(() => {
    currentPath = window.location.pathname;
  });

  const dispatch = createEventDispatcher();
</script>

<div class="app-container {isDarkMode ? 'app-container-dark dark-theme' : ''}">
  {#if isMobile}
    <div class="page-header-mobile {isDarkMode ? 'page-header-mobile-dark' : ''}">
      <button 
        class="page-header-mobile-menu"
        on:click={toggleMobileMenu}
        aria-label="Toggle menu"
      >
        <MenuIcon size="24" />
      </button>
      
      <div class="page-header-mobile-title">
        {#if pageTitle}
          {pageTitle}
        {:else}
          <div class="page-header-mobile-logo">
            {#if isDarkMode}
              <img src="/assets/light-logo.jpeg" alt="AYCOM" class="mobile-logo-img" />
            {:else}
              <img src="/assets/dark-logo.jpeg" alt="AYCOM" class="mobile-logo-img" />
            {/if}
          </div>
        {/if}
      </div>
      
      <button 
        class="page-header-mobile-search"
        on:click={toggleSearchBar}
        aria-label="Search"
      >
        <SearchIcon size="24" />
      </button>
    </div>
    
    {#if showSearchBar}
      <div class="mobile-search-container {isDarkMode ? 'mobile-search-container-dark' : ''}">
        <div class="mobile-search-form">
          <div class="mobile-search-input-wrapper">
            <SearchIcon size="18" />
            <input 
              type="text" 
              id="mobile-search-input"
              placeholder="Search" 
              class="mobile-search-input"
              bind:value={searchQuery}
              on:keydown={handleSearch}
            />
            {#if searchQuery}
              <button class="mobile-search-clear" on:click={clearSearch}>
                <XIcon size="16" />
              </button>
            {/if}
          </div>
        </div>
      </div>
    {/if}
  {/if}
  
  <div class="app-layout">
    {#if showLeftSidebar}
      <aside class="sidebar {isDarkMode ? 'sidebar-dark' : ''}">
        {#if !isMobile || (isMobile && showMobileMenu)}
          <LeftSide 
            {username}
            {displayName}
            {avatar}
            on:toggleComposeModal={handleToggleComposeModal}
            isCollapsed={isTablet || isSmallDesktop}
            isMobileMenu={isMobile && showMobileMenu}
            on:closeMobileMenu={closeMobileMenu}
          />
        {/if}
      </aside>
    {/if}
    
    <div class="content-wrapper">
      <main class="main-content {isDarkMode ? 'main-content-dark' : ''}">
        <slot></slot>
      </main>
      
      {#if showRightSidebar && isTablet}
        <div class="tablet-widgets">
          <RightSide 
            {isDarkMode}
            {trends}
            {suggestedFollows}
            isTabletView={true}
          />
        </div>
      {/if}
    </div>
    
    {#if showRightSidebar && !isMobile && !isTablet}
      <aside class="widgets-container {isDarkMode ? 'widgets-container-dark' : ''}">
        <RightSide 
          {isDarkMode}
          {trends}
          {suggestedFollows}
        />
      </aside>
    {/if}
  </div>
  
  <!-- Mobile navigation bar for smaller screens -->
  {#if isMobile}
    <nav class="mobile-nav {isDarkMode ? 'mobile-nav-dark' : ''}">
      <a href="/feed" class="mobile-nav-item {currentPath === '/feed' ? 'active' : ''}">
        <div class="mobile-nav-icon">
          <HomeIcon size="20" />
        </div>
        <span class="mobile-nav-label">Home</span>
      </a>
      <a href="/explore" class="mobile-nav-item {currentPath === '/explore' ? 'active' : ''}">
        <div class="mobile-nav-icon">
          <HashIcon size="20" />
        </div>
        <span class="mobile-nav-label">Explore</span>
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
        <span class="mobile-nav-label">Alerts</span>
      </a>
      <a href="/profile" class="mobile-nav-item {currentPath === '/profile' ? 'active' : ''}">
        <div class="mobile-nav-icon">
          <UserIcon size="20" />
        </div>
        <span class="mobile-nav-label">Profile</span>
      </a>
    </nav>
  {/if}
  
  {#if showMobileMenu && isMobile}
    <div 
      class="mobile-menu-overlay"
      on:click={closeMobileMenu}
      role="button"
      tabindex="-1"
      aria-label="Close menu"
    ></div>
  {/if}
  
  <ComposeTweetModal 
    isOpen={showComposeModal}
    {avatar}
    on:close={() => showComposeModal = false}
    on:posted={handleNewPost}
  />
  
  <Toast />
  <DebugPanel />
</div>

<style>
  .content-wrapper {
    grid-area: main;
    display: flex;
    flex-direction: column;
    width: 100%;
    min-width: 0;
  }
  
  .tablet-widgets {
    width: 100%;
    margin-top: var(--space-4);
  }
  
  .page-header-mobile {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-3) var(--space-4);
    border-bottom: 1px solid var(--border-color);
    position: sticky;
    top: 0;
    z-index: var(--z-header);
    background-color: var(--bg-primary);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
  }
  
  .page-header-mobile-dark {
    background-color: var(--dark-bg-primary);
    border-bottom: 1px solid var(--border-color-dark);
  }
  
  .page-header-mobile-menu,
  .page-header-mobile-search {
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background: transparent;
    border: none;
    color: var(--text-primary);
    cursor: pointer;
    transition: background-color var(--transition-fast);
  }
  
  .page-header-mobile-menu:hover,
  .page-header-mobile-search:hover {
    background-color: var(--bg-hover);
  }
  
  .page-header-mobile-title {
    font-weight: var(--font-weight-bold);
    font-size: var(--font-size-lg);
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .mobile-logo-img {
    height: 32px;
    width: 32px;
    object-fit: contain;
  }
  
  .mobile-menu-overlay {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    left: 0;
    background-color: rgba(0, 0, 0, 0.5);
    z-index: calc(var(--z-sidebar) - 1);
    backdrop-filter: blur(4px);
    -webkit-backdrop-filter: blur(4px);
  }
  
  .mobile-search-container {
    padding: var(--space-2) var(--space-4);
    border-bottom: 1px solid var(--border-color);
    background-color: var(--bg-primary);
  }
  
  .mobile-search-container-dark {
    background-color: var(--dark-bg-primary);
    border-bottom: 1px solid var(--border-color-dark);
  }
  
  .mobile-search-form {
    width: 100%;
  }
  
  .mobile-search-input-wrapper {
    display: flex;
    align-items: center;
    background-color: var(--bg-secondary);
    border-radius: var(--radius-full);
    padding: 0 var(--space-3);
  }
  
  .mobile-search-input {
    flex: 1;
    border: none;
    background: transparent;
    padding: var(--space-2);
    font-size: var(--font-size-base);
    color: var(--text-primary);
    outline: none;
    width: 100%;
  }
  
  .mobile-search-clear {
    background: transparent;
    border: none;
    color: var(--text-secondary);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--space-1);
    border-radius: 50%;
    cursor: pointer;
  }
  
  .mobile-search-clear:hover {
    background-color: var(--bg-hover);
    color: var(--color-primary);
  }
  
  @media (min-width: 769px) {
    .mobile-menu-overlay {
      display: none;
    }
  }
</style>