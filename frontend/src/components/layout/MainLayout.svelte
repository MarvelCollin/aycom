<script lang="ts">
  import LeftSide from './LeftSide.svelte';
  import RightSide from './RightSide.svelte';
  import Toast from '../common/Toast.svelte';
  import DebugPanel from '../common/DebugPanel.svelte';
  import ComposeTweetModal from '../social/ComposeTweetModal.svelte';
  import { useTheme } from '../../hooks/useTheme';
  import type { ITrend } from '../../interfaces/ITrend';
  import type { ISuggestedFollow } from '../../interfaces/ISocialMedia';
  import { createEventDispatcher } from 'svelte';
  import { onMount, onDestroy } from 'svelte';
  import { notificationWebsocketStore } from '../../stores/notificationWebsocketStore';
  import { isAuthenticated } from '../../utils/auth';
  import { createLoggerWithPrefix } from '../../utils/logger';
  import { notificationStore } from '../../stores/notificationStore';
  import { page } from '../../stores/routeStore';
  
  const logger = createLoggerWithPrefix('MainLayout');
  
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

  export let isDarkMode = false;
  export let isNavOpen = false;
  export let avatar = "";
  export let username = "";
  export let displayName = "";
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

  // Get unread notification count
  let unreadNotificationCount = 0;
  notificationStore.unreadCount.subscribe(count => {
    unreadNotificationCount = count;
  });

  // Track if notification WebSocket is connected
  let notificationWsConnected = false;

  // List of paths where notification WebSocket should be active
  const notificationEnabledPaths = ['/feed', '/notifications'];

  // Handle WebSocket connections based on current path
  function handleWebSocketConnection(path) {
    const shouldConnect = isAuthenticated() && notificationEnabledPaths.includes(path);
    
    if (shouldConnect && !notificationWsConnected) {
      logger.info(`Connecting to notification WebSocket on path: ${path}`);
      notificationWebsocketStore.connect();
      notificationWsConnected = true;
    } else if (!shouldConnect && notificationWsConnected) {
      logger.info(`Disconnecting notification WebSocket on path: ${path}`);
      notificationWebsocketStore.disconnect();
      notificationWsConnected = false;
    }
  }

  // Subscribe to page changes
  const unsubscribePageStore = page.subscribe(pageInfo => {
    if (pageInfo && pageInfo.route) {
      handleWebSocketConnection(pageInfo.route.id);
    }
  });

  onMount(() => {
    const checkViewport = () => {
      windowWidth = window.innerWidth;
      isMobile = windowWidth < 768;
      isTablet = windowWidth >= 768 && windowWidth < 992;
      isSmallDesktop = windowWidth >= 992 && windowWidth < 1200;
    };
    
    checkViewport();
    window.addEventListener('resize', checkViewport);
    
    const currentPath = window.location.pathname;
    handleWebSocketConnection(currentPath);
    
    return () => {
      window.removeEventListener('resize', checkViewport);
    };
  });
  
  // Clean up subscriptions when component is destroyed
  onDestroy(() => {
    // Unsubscribe from page store
    if (unsubscribePageStore) {
      unsubscribePageStore();
    }
    
    // Disconnect WebSocket if connected
    if (notificationWsConnected) {
      notificationWebsocketStore.disconnect();
      notificationWsConnected = false;
    }
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
        <MenuIcon size="22" />
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
        <SearchIcon size="22" />
      </button>
    </div>
    
    {#if showSearchBar}
      <div class="mobile-search-container {isDarkMode ? 'mobile-search-container-dark' : ''}">
        <div class="mobile-search-form">
          <div class="mobile-search-input-wrapper">
            <SearchIcon size="16" />
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
                <XIcon size="14" />
              </button>
            {/if}
          </div>
        </div>
      </div>
    {/if}
  {/if}
  
  <div class="app-layout">
    {#if showLeftSidebar}
      <aside class="sidebar {isDarkMode ? 'sidebar-dark' : ''} {isMobile && showMobileMenu ? 'visible' : ''}">
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
          {#if unreadNotificationCount > 0}
            <div class="mobile-notification-badge">
              {unreadNotificationCount > 99 ? '99+' : unreadNotificationCount}
            </div>
          {/if}
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
      on:keydown={(e) => { if (e.key === 'Escape') closeMobileMenu(); }}
      role="button"
      tabindex="0"
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
    padding: var(--space-3) var(--space-3);
    border-bottom: 1px solid var(--border-color);
    position: sticky;
    top: 0;
    z-index: var(--z-header);
    background-color: rgba(var(--bg-primary-rgb, 255, 255, 255), 0.95);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
  }
  
  .page-header-mobile-dark {
    background-color: rgba(var(--dark-bg-primary-rgb, 25, 25, 25), 0.95);
    border-bottom: 1px solid var(--border-color-dark);
  }
  
  .page-header-mobile-menu,
  .page-header-mobile-search {
    width: 36px;
    height: 36px;
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
    height: 28px;
    width: 28px;
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
    padding: var(--space-2) var(--space-3);
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
  
  /* Mobile navigation styling */
  .mobile-nav {
    box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.1);
  }
  
  .mobile-nav-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: var(--text-secondary);
    text-decoration: none;
    font-size: var(--font-size-sm);
    padding: var(--space-1);
    flex: 1;
  }
  
  .mobile-nav-item.active {
    color: var(--color-primary);
  }
  
  .mobile-nav-icon {
    position: relative;
    margin-bottom: var(--space-1);
  }
  
  .mobile-nav-label {
    font-size: var(--font-size-xs);
    margin-top: 2px;
  }
  
  .mobile-compose-btn {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background-color: var(--color-primary);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    box-shadow: var(--shadow-md);
    margin-top: -15px;
    margin-bottom: var(--space-1);
    cursor: pointer;
  }
  
  .mobile-notification-badge {
    position: absolute;
    top: -5px;
    right: -5px;
    background-color: var(--color-primary);
    color: white;
    font-size: var(--font-size-xs);
    font-weight: var(--font-weight-bold);
    border-radius: 50%;
    min-width: 16px;
    height: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0 var(--space-1);
  }
  
  .app-container {
    min-height: 100vh;
    max-width: 100vw;
    overflow-x: hidden;
    display: flex;
    flex-direction: column;
    background-color: var(--bg-primary);
    color: var(--text-primary);
  }
  
  .app-container-dark {
    background-color: var(--dark-bg-primary);
    color: var(--dark-text-primary);
  }
  
  .app-layout {
    display: grid;
    grid-template-columns: 1fr 3fr 1fr;
    grid-template-areas: "sidebar main widgets";
    min-height: 100vh;
    width: 100%;
    margin: 0 auto;
    max-width: 1440px;
  }
  
  /* Adjust sidebar width */
  @media (min-width: 1281px) {
    .app-layout {
      grid-template-columns: 1fr 3fr 1fr;
      max-width: 1280px;
    }
  }
  
  @media (max-width: 1280px) {
    .app-layout {
      grid-template-columns: 1fr 3fr 1fr;
      max-width: 100%;
    }
  }
  
  @media (max-width: 992px) {
    .app-layout {
      grid-template-columns: 1fr 4fr;
      grid-template-areas: "sidebar main";
    }
    
    .widgets-container {
      display: none;
    }
  }
  
  @media (max-width: 768px) {
    .app-layout {
      grid-template-columns: 1fr;
      grid-template-areas: "main";
      padding-bottom: 60px; /* Make room for mobile nav */
    }
    
    .sidebar {
      display: none;
    }
    
    .sidebar.visible {
      display: block;
      position: fixed;
      top: 0;
      left: 0;
      bottom: 0;
      width: 80%;
      max-width: 280px;
      z-index: var(--z-sidebar);
      transform: translateX(0);
    }
  }
  
  .sidebar {
    grid-area: sidebar;
    position: sticky;
    top: 0;
    height: 100vh;
    overflow-y: auto;
    border-right: 1px solid var(--border-color);
    z-index: var(--z-sidebar);
    background-color: var(--bg-primary);
    padding-left: var(--space-2);
    padding-right: var(--space-1);
  }
  
  .sidebar-dark {
    background-color: var(--dark-bg-primary);
    border-right: 1px solid var(--border-color-dark);
  }
  
  .main-content {
    width: 100%;
    max-width: 670px;
    margin: 0 auto;
    border-right: 1px solid var(--border-color);
    border-left: 1px solid var(--border-color);
    min-height: 100vh;
    overflow-x: hidden;
  }
  
  .main-content-dark {
    border-right: 1px solid var(--border-color-dark);
    border-left: 1px solid var(--border-color-dark);
  }
  
  .widgets-container {
    grid-area: widgets;
    position: sticky;
    top: 0;
    height: 100vh;
    overflow-y: auto;
    padding: var(--space-4);
    z-index: var(--z-sidebar);
  }
  
  .widgets-container-dark {
    background-color: var(--dark-bg-primary);
  }
  
  /* Custom styles for content containment */
  :global(.view-replies-button) {
    padding: var(--space-2) 0;
    width: auto;
    max-width: 95%;
    margin: 0 auto;
    text-align: center;
    color: var(--color-primary);
    background-color: transparent;
    border: none;
    cursor: pointer;
    transition: background-color var(--transition-fast);
  }
  
  :global(.reply-container) {
    width: 100%;
    max-width: 100%;
    border-top: 1px solid var(--border-color);
    padding: var(--space-2) var(--space-3);
  }
</style>