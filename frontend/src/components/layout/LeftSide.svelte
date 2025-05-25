<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import ThemeToggle from '../common/ThemeToggle.svelte';
  import { useTheme } from '../../hooks/useTheme';
  import { useAuth } from '../../hooks/useAuth';
  import { isAuthenticated, getUserId } from '../../utils/auth';
  import { toastStore } from '../../stores/toastStore';
  import { getProfile } from '../../api/user';
  import { onMount } from 'svelte';

  import HomeIcon from 'svelte-feather-icons/src/icons/HomeIcon.svelte';
  import HashIcon from 'svelte-feather-icons/src/icons/HashIcon.svelte';
  import BellIcon from 'svelte-feather-icons/src/icons/BellIcon.svelte';
  import MailIcon from 'svelte-feather-icons/src/icons/MailIcon.svelte';
  import BookmarkIcon from 'svelte-feather-icons/src/icons/BookmarkIcon.svelte';
  import UsersIcon from 'svelte-feather-icons/src/icons/UsersIcon.svelte';
  import UserIcon from 'svelte-feather-icons/src/icons/UserIcon.svelte';
  import SettingsIcon from 'svelte-feather-icons/src/icons/SettingsIcon.svelte';
  import PlusIcon from 'svelte-feather-icons/src/icons/PlusIcon.svelte';
  import StarIcon from 'svelte-feather-icons/src/icons/StarIcon.svelte';
  import MoreHorizontalIcon from 'svelte-feather-icons/src/icons/MoreHorizontalIcon.svelte';
  import LogOutIcon from 'svelte-feather-icons/src/icons/LogOutIcon.svelte';
  import CalendarIcon from 'svelte-feather-icons/src/icons/CalendarIcon.svelte';
  import CheckCircleIcon from 'svelte-feather-icons/src/icons/CheckCircleIcon.svelte';
  import LogInIcon from 'svelte-feather-icons/src/icons/LogInIcon.svelte';
  import ShieldIcon from 'svelte-feather-icons/src/icons/ShieldIcon.svelte';
  import XIcon from 'svelte-feather-icons/src/icons/XIcon.svelte';

  export let username = "";
  export let displayName = "";
  export let avatar = "https://secure.gravatar.com/avatar/0?d=mp";
  export let isCollapsed = false;
  export let isMobileMenu = false;
  
  const { theme } = useTheme();
  $: isDarkMode = $theme === 'dark';
  
  const { getAuthState, logout, getAuthToken } = useAuth();
  let authState = getAuthState();
  
  let debugging = false;
  let apiResponse = null;
  
  let userDetails = {
    username: username || 'guest',
    displayName: displayName || 'Guest User',
    avatar: avatar,
    userId: getUserId() || '',
    email: '',
    isVerified: false,
    joinDate: ''
  };
  
  let isAdmin = false;
  let windowWidth = 0;
  
  async function fetchUserProfile() {
    if (!isAuthenticated()) {
      console.log('User not authenticated, skipping profile fetch');
      return;
    }
    
    console.log('Fetching user profile...');
    try {
      const response = await getProfile();
      apiResponse = response;
      console.log('API Response:', apiResponse);
      
      const userData = response.user || (response.data && response.data.user);
      
      if (userData) {
        console.log('User data:', userData);
        userDetails = {
          username: userData.username || username,
          displayName: userData.name || userData.display_name || displayName,
          avatar: userData.profile_picture_url || avatar,
          userId: userData.id || getUserId() || '',
          email: userData.email || '',
          isVerified: userData.is_verified || false,
          joinDate: userData.created_at ? new Date(userData.created_at).toLocaleDateString() : ''
        };
        
        isAdmin = userData.is_admin || false;
        console.log('Is admin?', userData.is_admin, isAdmin);
        
        if (isAdmin) {
          console.log('User is an admin, showing admin panel link');
        }
        
        username = userDetails.username;
        displayName = userDetails.displayName;
        avatar = userDetails.avatar;
        
        console.log('Profile loaded successfully:', userDetails);
      } else {
        console.warn('Response received but no user data found in:', response);
      }
    } catch (err) {
      console.error('Failed to fetch user profile:', err);
      toastStore.showToast('Failed to load user profile. Please try again.', 'error');
    }
  }
  
  function toggleDebug() {
    debugging = !debugging;
  }
  
  async function handleLogout() {
    try {
      await logout();
      window.location.href = '/login';
    } catch (err) {
      console.error('Error during logout:', err);
      toastStore.showToast('Logout failed. Please try again.', 'error');
    }
  }
  
  const dispatch = createEventDispatcher();
  
  $: navigationItems = [
    { label: "Feed", path: "/feed", icon: "home" },
    { label: "Explore", path: "/explore", icon: "hash" },
    { label: "Notifications", path: "/notifications", icon: "bell" },
    { label: "Messages", path: "/messages", icon: "mail" },
    { label: "Bookmarks", path: "/bookmarks", icon: "bookmark" },
    { label: "Communities", path: "/communities", icon: "users" },
    { label: "Premium", path: "/premium", icon: "star" },
    { label: "Profile", path: "/profile", icon: "user" },
    { label: "Settings", path: "/settings", icon: "settings" },
    ...(isAdmin ? [{ label: "Admin", path: "/admin", icon: "shield" }] : []),
  ];
  
  let showUserMenu = false;
  function toggleUserMenu() {
    showUserMenu = !showUserMenu;
    
    if (showUserMenu && isAuthenticated()) {
      fetchUserProfile();
    }
  }
  
  function handleToggleComposeModal() {
    console.log('LeftSide: Dispatching toggleComposeModal event');
    dispatch('toggleComposeModal');
  }
  
  function handleCloseMobileMenu() {
    if (isMobileMenu) {
      dispatch('closeMobileMenu');
    }
  }
  
  let currentPath = '';
  
  onMount(() => {
    currentPath = window.location.pathname;
    
    if (isAuthenticated()) {
      console.log('User is authenticated, fetching profile on mount');
      fetchUserProfile();
    } else {
      console.log('User is not authenticated on mount');
    }
    
    const intervalId = setInterval(() => {
      if (isAuthenticated() && userDetails.username === 'guest') {
        console.log('User is authenticated but still shows as guest, refreshing profile');
        fetchUserProfile();
      }
    }, 5000);
    
    // Check window width to determine collapsed state
    const checkWidth = () => {
      windowWidth = window.innerWidth;
    };
    
    checkWidth();
    window.addEventListener('resize', checkWidth);
    
    return () => {
      clearInterval(intervalId);
      window.removeEventListener('resize', checkWidth);
    };
  });
</script>

<div class="sidebar {isDarkMode ? 'sidebar-dark' : ''} {isCollapsed ? 'sidebar-collapsed' : ''} {isMobileMenu ? 'sidebar-mobile' : ''}">
  {#if isMobileMenu}
    <div class="sidebar-mobile-header">
      <button 
        class="sidebar-close-btn"
        on:click={handleCloseMobileMenu}
        aria-label="Close menu"
      >
        <XIcon size="24" />
      </button>
    </div>
  {/if}

  <div class="sidebar-logo">
    <a href="/" aria-label="Home">
      {#if isDarkMode}
        <img src="/assets/light-logo.jpeg" alt="AYCOM" class="logo-img" />
      {:else}
        <img src="/assets/dark-logo.jpeg" alt="AYCOM" class="logo-img" />
      {/if}
    </a>
  </div>

  <nav class="sidebar-nav">
    <ul>
      {#each navigationItems as item}
        <li>
          <a 
            href={item.path} 
            class="sidebar-nav-item {currentPath === item.path ? 'active' : ''} {isDarkMode ? 'sidebar-nav-item-dark' : ''}"
            aria-label={isCollapsed ? item.label : undefined}
            on:click={() => isMobileMenu && handleCloseMobileMenu()}
          >
            <div class="sidebar-nav-icon">
              {#if item.icon === 'home'}
                <HomeIcon size="24" />
              {:else if item.icon === 'hash'}
                <HashIcon size="24" />
              {:else if item.icon === 'bell'}
                <BellIcon size="24" />
              {:else if item.icon === 'mail'}
                <MailIcon size="24" />
              {:else if item.icon === 'bookmark'}
                <BookmarkIcon size="24" />
              {:else if item.icon === 'users'}
                <UsersIcon size="24" />
              {:else if item.icon === 'star'}
                <StarIcon size="24" />
              {:else if item.icon === 'user'}
                <UserIcon size="24" />
              {:else if item.icon === 'settings'}
                <SettingsIcon size="24" />
              {:else if item.icon === 'shield'}
                <ShieldIcon size="24" />
              {/if}
            </div>
            <span class="sidebar-nav-text">{item.label}</span>
          </a>
        </li>
      {/each}
    </ul>

    <button 
      class="sidebar-tweet-btn {isDarkMode ? 'sidebar-tweet-btn-dark' : ''}"
      on:click={handleToggleComposeModal}
      aria-label={isCollapsed ? "Create new post" : undefined}
    >
      <span class="sidebar-tweet-btn-icon">
        <PlusIcon size="24" />
      </span>
      <span class="sidebar-tweet-btn-text">Post</span>
    </button>

    <div class="sidebar-theme-toggle">
      <ThemeToggle size="md" />
    </div>
  </nav>

  <div 
    class="sidebar-profile"
    on:click={toggleUserMenu}
    on:keydown={(e) => e.key === 'Enter' && toggleUserMenu()}
    role="button"
    tabindex="0"
  >
    <div class="sidebar-profile-avatar">
      <img 
        src={userDetails.avatar || "https://secure.gravatar.com/avatar/0?d=mp"} 
        alt={userDetails.displayName}
      />
    </div>
    <div class="sidebar-profile-info">
      <div class="sidebar-profile-name">{userDetails.displayName}</div>
      <div class="sidebar-profile-username">@{userDetails.username}</div>
    </div>
    <div class="sidebar-profile-more">
      <MoreHorizontalIcon size="20" />
    </div>
  </div>

  {#if showUserMenu}
    <div class="sidebar-user-menu {isDarkMode ? 'sidebar-user-menu-dark' : ''}">
      <div class="sidebar-user-header {isDarkMode ? 'sidebar-user-header-dark' : ''}">
        <div class="sidebar-profile-name">{userDetails.displayName}</div>
        <div class="sidebar-profile-username">@{userDetails.username}</div>
        
        {#if userDetails.isVerified}
          <div class="sidebar-user-verified">
            <div class="sidebar-user-verified-icon">
              <CheckCircleIcon size="14" />
            </div>
            <span>Verified Account</span>
          </div>
        {/if}
        
        {#if userDetails.email}
          <div class="sidebar-user-email">{userDetails.email}</div>
        {/if}
        
        {#if userDetails.joinDate}
          <div class="sidebar-user-join">
            <div class="sidebar-user-join-icon">
              <CalendarIcon size="14" />
            </div>
            <span>Joined {userDetails.joinDate}</span>
          </div>
        {/if}
      </div>
      
      {#if userDetails.username !== 'guest'}
        <div 
          class="sidebar-user-menu-item {isDarkMode ? 'sidebar-user-menu-item-dark' : ''}"
          role="button"
          tabindex="0"
          on:click={handleLogout}
          on:keydown={(e) => e.key === 'Enter' && handleLogout()}
        >
          <div class="sidebar-user-menu-icon">
            <LogOutIcon size="16" />
          </div>
          <span>Log out @{userDetails.username}</span>
        </div>
      {:else}
        <div 
          class="sidebar-user-menu-item {isDarkMode ? 'sidebar-user-menu-item-dark' : ''}"
          role="button"
          tabindex="0"
          on:click={() => window.location.href = '/login'}
          on:keydown={(e) => e.key === 'Enter' && (window.location.href = '/login')}
        >
          <div class="sidebar-user-menu-icon">
            <LogInIcon size="16" />
          </div>
          <span>Log in</span>
        </div>
      {/if}
      
      {#if import.meta.env.DEV}
        <div class="sidebar-debug">
          <div 
            class="sidebar-debug-title"
            role="button"
            tabindex="0"
            on:click={toggleDebug}
            on:keydown={(e) => e.key === 'Enter' && toggleDebug()}
          >
            Debug Info {debugging ? '▲' : '▼'}
          </div>
          {#if debugging}
            <div class="sidebar-debug-content">
              <pre>Auth: {JSON.stringify(authState, null, 2)}</pre>
              <pre>User: {JSON.stringify(userDetails, null, 2)}</pre>
              <pre>API: {JSON.stringify(apiResponse, null, 2)}</pre>
            </div>
          {/if}
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .sidebar {
    display: flex;
    flex-direction: column;
    height: 100vh;
    padding: var(--space-2);
    position: sticky;
    top: 0;
    z-index: var(--z-sidebar);
    transition: width 0.3s ease;
    width: var(--sidebar-width);
    background-color: var(--bg-primary);
  }
  
  .sidebar-collapsed {
    width: var(--sidebar-collapsed-width);
  }
  
  .sidebar-mobile {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    width: var(--sidebar-width);
    z-index: var(--z-sidebar);
    box-shadow: var(--shadow-lg);
  }
  
  .sidebar-mobile-header {
    display: flex;
    justify-content: flex-end;
    padding: var(--space-2);
  }
  
  .sidebar-close-btn {
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
  }
  
  .sidebar-close-btn:hover {
    background-color: var(--bg-hover);
  }
  
  .sidebar-dark {
    color: var(--text-primary-dark);
    background-color: var(--dark-bg-primary);
  }
  
  .sidebar-logo {
    padding: var(--space-3) var(--space-2);
    display: flex;
    align-items: center;
    justify-content: flex-start;
  }
  
  .logo-img {
    width: 40px;
    height: 40px;
    object-fit: contain;
  }
  
  .sidebar-nav {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow-y: auto;
  }
  
  .sidebar-nav ul {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  
  .sidebar-nav-item {
    display: flex;
    align-items: center;
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-full);
    color: var(--text-primary);
    font-weight: var(--font-weight-medium);
    transition: background-color var(--transition-fast);
    text-decoration: none;
    margin-bottom: var(--space-1);
  }
  
  .sidebar-nav-item:hover {
    background-color: var(--bg-hover);
  }
  
  .sidebar-nav-item.active {
    font-weight: var(--font-weight-bold);
    color: var(--color-primary);
  }
  
  .sidebar-nav-item-dark:hover {
    background-color: var(--dark-hover-bg);
  }
  
  .sidebar-nav-icon {
    margin-right: var(--space-4);
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .sidebar-collapsed .sidebar-nav-icon {
    margin-right: 0;
  }
  
  .sidebar-nav-text {
    display: inline-block;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .sidebar-collapsed .sidebar-nav-text {
    display: none;
  }
  
  /* Tweet button in sidebar */
  .sidebar-tweet-btn {
    margin-top: var(--space-3);
    background-color: var(--color-primary);
    color: white;
    font-weight: var(--font-weight-bold);
    border-radius: var(--radius-full);
    padding: var(--space-3) var(--space-4);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background-color var(--transition-fast), transform var(--transition-fast);
    border: none;
    cursor: pointer;
    width: 90%;
    margin-left: auto;
    margin-right: auto;
    box-shadow: 0 2px 5px rgba(var(--color-primary-rgb), 0.3);
  }
  
  .sidebar-tweet-btn-dark {
    background-color: var(--color-primary);
    color: white;
    box-shadow: 0 2px 5px rgba(var(--color-primary-rgb), 0.5);
  }
  
  .sidebar-tweet-btn:hover {
    background-color: var(--color-primary-hover);
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(var(--color-primary-rgb), 0.4);
  }
  
  .sidebar-tweet-btn-icon {
    display: none;
  }
  
  .sidebar-collapsed .sidebar-tweet-btn {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    padding: 0;
  }
  
  .sidebar-collapsed .sidebar-tweet-btn-icon {
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .sidebar-collapsed .sidebar-tweet-btn-text {
    display: none;
  }
  
  .sidebar-tweet-btn-text {
    display: block;
  }
  
  /* Theme toggle */
  .sidebar-theme-toggle {
    display: flex;
    justify-content: center;
    margin: var(--space-4) 0;
  }
  
  /* User profile section */
  .sidebar-profile {
    display: flex;
    align-items: center;
    padding: var(--space-3) var(--space-2);
    border-radius: var(--radius-lg);
    cursor: pointer;
    margin-top: auto;
    transition: background-color var(--transition-fast);
  }
  
  .sidebar-profile:hover {
    background-color: var(--bg-hover);
  }
  
  .sidebar-collapsed .sidebar-profile {
    justify-content: center;
  }
  
  .sidebar-profile-avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    overflow: hidden;
    margin-right: var(--space-3);
    flex-shrink: 0;
  }
  
  .sidebar-collapsed .sidebar-profile-avatar {
    margin-right: 0;
  }
  
  .sidebar-profile-avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  
  .sidebar-profile-info {
    flex: 1;
    min-width: 0;
    margin-right: var(--space-2);
  }
  
  .sidebar-collapsed .sidebar-profile-info {
    display: none;
  }
  
  .sidebar-profile-name {
    font-weight: var(--font-weight-bold);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .sidebar-profile-username {
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .sidebar-profile-more {
    color: var(--text-secondary);
    display: flex;
    align-items: center;
  }
  
  .sidebar-collapsed .sidebar-profile-more {
    display: none;
  }
  
  /* User menu dropdown */
  .sidebar-user-menu {
    position: absolute;
    bottom: 80px;
    left: var(--space-4);
    width: calc(100% - var(--space-8));
    background-color: var(--bg-primary);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
    z-index: var(--z-dropdown);
    overflow: hidden;
  }
  
  .sidebar-user-menu-dark {
    background-color: var(--dark-bg-secondary);
    box-shadow: var(--shadow-lg-dark);
  }
  
  .sidebar-user-header {
    padding: var(--space-4);
    border-bottom: 1px solid var(--border-color);
  }
  
  .sidebar-user-header-dark {
    border-bottom: 1px solid var(--border-color-dark);
  }
  
  .sidebar-user-verified,
  .sidebar-user-join,
  .sidebar-user-email {
    display: flex;
    align-items: center;
    margin-top: var(--space-2);
    font-size: var(--font-size-sm);
    color: var(--text-secondary);
  }
  
  .sidebar-user-verified-icon,
  .sidebar-user-join-icon {
    margin-right: var(--space-1);
    display: flex;
    align-items: center;
    color: var(--color-primary);
  }
  
  .sidebar-user-menu-item {
    display: flex;
    align-items: center;
    padding: var(--space-3) var(--space-4);
    cursor: pointer;
    transition: background-color var(--transition-fast);
  }
  
  .sidebar-user-menu-item:hover {
    background-color: var(--bg-hover);
  }
  
  .sidebar-user-menu-item-dark:hover {
    background-color: var(--dark-hover-bg);
  }
  
  .sidebar-user-menu-icon {
    margin-right: var(--space-3);
    display: flex;
    align-items: center;
  }
  
  /* Debug section */
  .sidebar-debug {
    padding: var(--space-3) var(--space-4);
    border-top: 1px solid var(--border-color);
    font-size: var(--font-size-sm);
  }
  
  .sidebar-debug-title {
    font-weight: var(--font-weight-bold);
    cursor: pointer;
    margin-bottom: var(--space-2);
  }
  
  .sidebar-debug-content {
    background-color: var(--bg-tertiary);
    padding: var(--space-2);
    border-radius: var(--radius-md);
    overflow-x: auto;
    font-family: monospace;
    font-size: var(--font-size-xs);
  }
  
  .sidebar-debug-content pre {
    margin: 0;
    white-space: pre-wrap;
  }
</style>