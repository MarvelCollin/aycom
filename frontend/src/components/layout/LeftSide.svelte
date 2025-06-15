<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import ThemeToggle from '../common/ThemeToggle.svelte';
  import { useTheme } from '../../hooks/useTheme';
  import { useAuth } from '../../hooks/useAuth';
  import { isAuthenticated, getUserId, isUserAdmin } from '../../utils/auth';
  import { toastStore } from '../../stores/toastStore';
  import { getProfile, checkAdminStatus } from '../../api/user';
  import { onMount } from 'svelte';
  import { notificationStore } from '../../stores/notificationStore';

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

  export let isDarkMode = false;
  export let isNavOpen = false;
  export let avatar = "";
  export let username = "";
  export let displayName = "";
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
  
  $: if (isAdmin) {
    console.log("ADMIN STATUS CHANGED: User IS an admin, admin link should be visible");
    console.log("Navigation items:", navigationItems);
  }
  
  async function debugAdminStatus() {
    try {
      // First check the auth state
      const authState = getAuthState();
      if (authState && isUserAdmin(authState)) {
        console.log('User already has admin status in auth store');
        isAdmin = true;
        return;
      }
      
      // Verify with the backend
      const adminStatusFromAPI = await checkAdminStatus();
      
      if (adminStatusFromAPI) {
        console.log('API confirmed user is admin, updating auth store');
        isAdmin = true;
        
        // Also update localStorage directly as a fallback
        try {
          const authData = localStorage.getItem('auth');
          if (authData) {
            const auth = JSON.parse(authData);
            auth.is_admin = true;
            localStorage.setItem('auth', JSON.stringify(auth));
            console.log('Updated localStorage with admin status');
          }
        } catch (e) {
          console.error('Error updating localStorage auth data:', e);
        }
      } else {
        console.log('API confirmed user is NOT admin');
      }
    } catch (error) {
      console.error('Error checking admin status:', error);
    }
  }
  
  async function fetchUserProfile() {
    if (!isAuthenticated()) {
      console.log('User not authenticated, skipping profile fetch');
      return;
    }
    
    debugAdminStatus();
    
    console.log('Fetching user profile...');
    try {
      const response = await getProfile();
      apiResponse = response;
      console.log('API Response:', apiResponse);
      
      const userData = response?.user || (response?.data && response?.data.user);
      
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
        
        // Check admin status from both the API response and the auth store
        const authState = getAuthState();
        isAdmin = isUserAdmin(userData) || (authState && isUserAdmin(authState));
        
        console.log('Is admin?', isAdmin, '(API:', userData.is_admin, ', Auth store:', authState?.is_admin, ')');
        
        if (isAdmin) {
          console.log('User is an admin, showing admin panel link');
        }
        
        username = userDetails.username;
        displayName = userDetails.displayName;
        avatar = userDetails.avatar;
        
        console.log('Profile loaded successfully:', userDetails);
      } else {
        console.warn('Response received but no user data found in:', response);
        
        // Even if userData is missing, still check auth store for admin status
        const authState = getAuthState();
        if (authState && isUserAdmin(authState)) {
          isAdmin = true;
          console.log('No API user data but user is admin based on auth store');
        }
      }
    } catch (err) {
      console.error('Failed to fetch user profile:', err);
      
      // Even on error, still check auth store for admin status
      const authState = getAuthState();
      if (authState && isUserAdmin(authState)) {
        isAdmin = true;
        console.log('Profile fetch failed but user is admin based on auth store');
      }
      
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
  
  // Get unread notification count
  let unreadNotificationCount;
  notificationStore.unreadCount.subscribe(count => {
    unreadNotificationCount = count;
  });
  
  onMount(() => {
    currentPath = window.location.pathname;
    
    // Run debug admin status check
    debugAdminStatus();
    
    // Check admin status from auth store as early as possible
    const authState = getAuthState();
    if (authState && isUserAdmin(authState)) {
      isAdmin = true;
      console.log('User is admin based on auth state');
    }
    
    // Force a check for specific admin user IDs
    const userId = getUserId();
    console.log("Current logged in user ID:", userId);
    
    // Last resort solution for known admin users
    if (userId === "91df5727-a9c5-427e-94ce-e0486e3bfdb7" || 
        userId === "f9d1a0f6-1b06-4411-907a-7a0f585df535") {
      console.log("DEBUG: Known admin user detected by ID, forcing admin view");
      isAdmin = true;
      
      // Force update the auth state too
      try {
        const authData = localStorage.getItem('auth');
        if (authData) {
          const auth = JSON.parse(authData);
          auth.is_admin = true;
          localStorage.setItem('auth', JSON.stringify(auth));
          console.log('Updated auth state with admin status for known admin');
        }
      } catch (e) {
        console.error('Error updating auth state for known admin:', e);
      }
    }
    
    // If the user is authenticated, try to load their profile and check admin status
    if (isAuthenticated()) {
      console.log('User is authenticated, fetching profile on mount');
      
      // Try to check admin status directly via API
      checkAdminStatus()
        .then(adminCheck => {
          if (adminCheck) {
            isAdmin = true;
            console.log('User is admin according to admin check API');
          }
        })
        .catch(err => {
          console.error('Admin check failed:', err);
        });
      
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

<div class="sidebar-inner {isDarkMode ? 'sidebar-inner-dark' : ''} {isCollapsed ? 'sidebar-inner-collapsed' : ''} {isMobileMenu ? 'sidebar-inner-mobile' : ''}">
  {#if isMobileMenu}
    <div class="sidebar-mobile-header">
      <button on:click={handleCloseMobileMenu} class="sidebar-close-btn">
        <XIcon size="24" />
      </button>
    </div>
  {/if}
  
  <div class="sidebar-logo">
    {#if isDarkMode}
      <img src="/assets/light-logo.jpeg" alt="AYCOM" class="logo-img" />
    {:else}
      <img src="/assets/dark-logo.jpeg" alt="AYCOM" class="logo-img" />
    {/if}
    {#if !isCollapsed}
      <span class="logo-text">AYCOM</span>
    {/if}
  </div>
  
  <nav class="sidebar-nav">
    <ul class="sidebar-nav-list">
      {#each navigationItems as item}
        <li>
          <a 
            href={item.path} 
            class="sidebar-nav-item {currentPath === item.path ? 'active' : ''}"
            class:dark={isDarkMode}
            on:click={isMobileMenu ? handleCloseMobileMenu : undefined}
          >
            <div class="sidebar-nav-icon">
              {#if item.icon === 'home'}
                <HomeIcon size={isCollapsed ? "24" : "20"} />
              {:else if item.icon === 'hash'}
                <HashIcon size={isCollapsed ? "24" : "20"} />
              {:else if item.icon === 'bell'}
                <BellIcon size={isCollapsed ? "24" : "20"} />
                {#if unreadNotificationCount > 0 && item.label === 'Notifications'}
                  <div class="notification-badge">
                    {unreadNotificationCount > 99 ? '99+' : unreadNotificationCount}
                  </div>
                {/if}
              {:else if item.icon === 'mail'}
                <MailIcon size={isCollapsed ? "24" : "20"} />
              {:else if item.icon === 'bookmark'}
                <BookmarkIcon size={isCollapsed ? "24" : "20"} />
              {:else if item.icon === 'users'}
                <UsersIcon size={isCollapsed ? "24" : "20"} />
              {:else if item.icon === 'star'}
                <StarIcon size={isCollapsed ? "24" : "20"} />
              {:else if item.icon === 'user'}
                <UserIcon size={isCollapsed ? "24" : "20"} />
              {:else if item.icon === 'settings'}
                <SettingsIcon size={isCollapsed ? "24" : "20"} />
              {:else if item.icon === 'shield'}
                <ShieldIcon size={isCollapsed ? "24" : "20"} />
              {/if}
            </div>
            {#if !isCollapsed}
              <span class="sidebar-nav-text">{item.label}</span>
            {/if}
          </a>
        </li>
      {/each}
    </ul>
    
    <button 
      class="sidebar-tweet-btn {isDarkMode ? 'sidebar-tweet-btn-dark' : ''}"
      on:click={handleToggleComposeModal}
    >
      <div class="sidebar-tweet-btn-icon">
        <PlusIcon size="20" />
      </div>
      <span class="sidebar-tweet-btn-text">Post</span>
    </button>
  </nav>
  
  <div class="sidebar-theme-toggle">
    <ThemeToggle />
  </div>
  
  <div class="sidebar-profile" on:click={toggleUserMenu}>
    <div class="sidebar-profile-avatar">
      <img src={userDetails.avatar} alt={userDetails.displayName} />
    </div>
    {#if !isCollapsed}
      <div class="sidebar-profile-info">
        <div class="sidebar-profile-name">
          {userDetails.displayName}
          {#if userDetails.isVerified}
            <span class="sidebar-verified-badge">
              <CheckCircleIcon size="12" />
            </span>
          {/if}
        </div>
        <div class="sidebar-profile-username">@{userDetails.username}</div>
      </div>
      <div class="sidebar-profile-more">
        <MoreHorizontalIcon size="16" />
      </div>
    {/if}
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
            <span>Verified</span>
          </div>
        {/if}
        
        {#if userDetails.email}
          <div class="sidebar-user-email">
            <span>{userDetails.email}</span>
          </div>
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
      
      <div 
        class="sidebar-user-menu-item {isDarkMode ? 'sidebar-user-menu-item-dark' : ''}"
        on:click={() => { window.location.href = '/profile'; }}
      >
        <div class="sidebar-user-menu-icon">
          <UserIcon size="16" />
        </div>
        <span>View Profile</span>
      </div>
      
      <div 
        class="sidebar-user-menu-item {isDarkMode ? 'sidebar-user-menu-item-dark' : ''}"
        on:click={handleLogout}
      >
        <div class="sidebar-user-menu-icon">
          <LogOutIcon size="16" />
        </div>
        <span>Log Out</span>
      </div>
      
      <div 
        class="sidebar-user-menu-item {isDarkMode ? 'sidebar-user-menu-item-dark' : ''}"
        on:click={toggleDebug}
      >
        <div class="sidebar-user-menu-icon">
          <SettingsIcon size="16" />
        </div>
        <span>Debug Info</span>
      </div>
    </div>
  {/if}
  
  {#if debugging}
    <div class="sidebar-debug">
      <h4 class="sidebar-debug-title">Debug Info</h4>
      <div class="sidebar-debug-item">
        <strong>User ID:</strong> {userDetails.userId}
      </div>
      <div class="sidebar-debug-item">
        <strong>Admin:</strong> {isAdmin ? 'Yes' : 'No'}
      </div>
      <div class="sidebar-debug-item">
        <strong>API Response:</strong>
        <pre>{JSON.stringify(apiResponse, null, 2)}</pre>
      </div>
    </div>
  {/if}
</div>

<style>
  .sidebar-inner {
    display: flex;
    flex-direction: column;
    height: 100%;
    width: 100%;
    max-width: var(--sidebar-width, 260px);
    padding: var(--space-2, 8px);
    transition: all 0.3s ease;
    position: relative;
    background-color: var(--bg-secondary, #1a1a1a);
    border-right: 1px solid var(--border-color, rgba(255, 255, 255, 0.1));
  }
  
  .sidebar-inner-dark {
    color: var(--text-primary-dark, white);
  }
  
  .sidebar-inner-collapsed {
    max-width: var(--sidebar-collapsed-width, 80px);
    align-items: center;
  }
  
  .sidebar-inner-mobile {
    max-width: 100%;
    height: 100%;
    padding-top: 0;
  }
  
  .sidebar-mobile-header {
    display: flex;
    justify-content: flex-end;
    padding: var(--space-2, 8px) var(--space-2, 8px) var(--space-4, 16px) var(--space-2, 8px);
  }
  
  .sidebar-close-btn {
    background: transparent;
    border: none;
    color: white;
    cursor: pointer;
    width: 40px;
    height: 40px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background-color 0.2s ease;
  }
  
  .sidebar-close-btn:hover {
    background-color: rgba(255, 255, 255, 0.1);
  }
  
  .sidebar-logo {
    display: flex;
    align-items: center;
    padding: var(--space-3, 12px) var(--space-4, 16px);
    margin-bottom: var(--space-4, 16px);
  }
  
  .logo-img {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    margin-right: var(--space-2, 8px);
    object-fit: cover;
  }
  
  .logo-text {
    font-size: var(--font-size-lg, 1.25rem);
    font-weight: var(--font-weight-bold, 700);
    color: white;
  }
  
  .sidebar-nav {
    flex: 1;
    display: flex;
    flex-direction: column;
    margin-bottom: var(--space-4, 16px);
    overflow-y: auto;
  }
  
  .sidebar-nav-list {
    list-style: none;
    padding: 0;
    margin: 0 0 var(--space-4, 16px) 0;
  }
  
  .sidebar-nav-item {
    display: flex;
    align-items: center;
    padding: var(--space-3, 12px) var(--space-4, 16px);
    border-radius: var(--radius-full, 9999px);
    margin-bottom: var(--space-2, 8px);
    text-decoration: none;
    color: white;
    font-weight: var(--font-weight-medium, 500);
    transition: all 0.2s ease;
    position: relative;
  }
  
  .sidebar-nav-item:hover {
    background-color: rgba(255, 255, 255, 0.1);
    transform: translateY(-1px);
  }
  
  .sidebar-nav-item.active {
    font-weight: var(--font-weight-bold, 700);
    color: var(--color-primary, #3b82f6);
    background-color: rgba(59, 130, 246, 0.15);
  }
  
  .sidebar-nav-icon {
    margin-right: var(--space-3, 12px);
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    color: white;
  }
  
  .sidebar-inner-collapsed .sidebar-nav-item {
    justify-content: center;
    padding: var(--space-3, 12px) 0;
  }
  
  .sidebar-inner-collapsed .sidebar-nav-icon {
    margin-right: 0;
  }
  
  .sidebar-tweet-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--space-3, 12px);
    border-radius: var(--radius-full, 9999px);
    background-color: var(--color-primary, #3b82f6);
    color: white;
    border: none;
    font-weight: var(--font-weight-bold, 700);
    transition: all 0.2s ease;
    cursor: pointer;
    width: 90%;
    margin: 0 auto;
    box-shadow: 0 2px 5px rgba(59, 130, 246, 0.3);
  }
  
  .sidebar-tweet-btn:hover {
    background-color: var(--color-primary-hover, #2563eb);
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(59, 130, 246, 0.4);
  }
  
  .sidebar-tweet-btn-icon {
    display: none;
  }
  
  .sidebar-inner-collapsed .sidebar-tweet-btn {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    padding: 0;
  }
  
  .sidebar-inner-collapsed .sidebar-tweet-btn-icon {
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .sidebar-inner-collapsed .sidebar-tweet-btn-text {
    display: none;
  }
  
  .sidebar-tweet-btn-text {
    display: block;
  }
  
  /* Theme toggle */
  .sidebar-theme-toggle {
    display: flex;
    justify-content: center;
    margin: var(--space-4, 16px) 0;
  }
  
  /* User profile section */
  .sidebar-profile {
    display: flex;
    align-items: center;
    padding: var(--space-3, 12px) var(--space-2, 8px);
    border-radius: var(--radius-lg, 8px);
    cursor: pointer;
    margin-top: auto;
    transition: background-color 0.2s ease;
  }
  
  .sidebar-profile:hover {
    background-color: rgba(255, 255, 255, 0.1);
  }
  
  .sidebar-inner-collapsed .sidebar-profile {
    justify-content: center;
    width: 100%;
    padding: var(--space-2, 8px) 0;
  }
  
  .sidebar-profile-avatar {
    width: 42px;
    height: 42px;
    border-radius: 50%;
    overflow: hidden;
    margin-right: var(--space-3, 12px);
    flex-shrink: 0;
    border: 2px solid rgba(255, 255, 255, 0.2);
  }
  
  .sidebar-inner-collapsed .sidebar-profile-avatar {
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
    margin-right: var(--space-2, 8px);
  }
  
  .sidebar-inner-collapsed .sidebar-profile-info {
    display: none;
  }
  
  .sidebar-profile-name {
    font-weight: var(--font-weight-bold, 700);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    display: flex;
    align-items: center;
    gap: 4px;
    color: white;
  }
  
  .sidebar-verified-badge {
    color: #1DA1F2 !important;
    display: inline-flex;
    align-items: center;
    flex-shrink: 0;
  }
  
  .sidebar-profile-username {
    font-size: var(--font-size-sm, 0.875rem);
    color: rgba(255, 255, 255, 0.7);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .sidebar-profile-more {
    color: rgba(255, 255, 255, 0.7);
    display: flex;
    align-items: center;
  }
  
  .sidebar-inner-collapsed .sidebar-profile-more {
    display: none;
  }
  
  /* User menu dropdown */
  .sidebar-user-menu {
    position: absolute;
    bottom: 80px;
    left: var(--space-4, 16px);
    width: calc(100% - var(--space-8, 32px));
    background-color: var(--bg-primary, #121212);
    border-radius: var(--radius-lg, 8px);
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
    z-index: var(--z-dropdown, 50);
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.1);
  }
  
  .sidebar-user-menu-dark {
    background-color: var(--dark-bg-secondary);
    box-shadow: var(--shadow-lg-dark);
  }
  
  .sidebar-user-header {
    padding: var(--space-4, 16px);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }
  
  .sidebar-user-header-dark {
    border-bottom: 1px solid var(--border-color-dark);
  }
  
  .sidebar-user-join,
  .sidebar-user-email {
    display: flex;
    align-items: center;
    margin-top: var(--space-2, 8px);
    font-size: var(--font-size-sm, 0.875rem);
    color: rgba(255, 255, 255, 0.7);
  }
  
  .sidebar-user-verified {
    display: flex;
    align-items: center;
    margin-top: var(--space-2, 8px);
    font-size: var(--font-size-sm, 0.875rem);
    color: #1DA1F2;
  }

  .sidebar-user-join-icon {
    margin-right: var(--space-1, 4px);
    display: flex;
    align-items: center;
    color: var(--color-primary, #3b82f6);
  }
  
  .sidebar-user-verified-icon {
    margin-right: var(--space-1, 4px);
    display: flex;
    align-items: center;
    color: #1DA1F2;
    filter: drop-shadow(0 0 1px rgba(29, 161, 242, 0.3));
  }
  
  .sidebar-user-menu-item {
    display: flex;
    align-items: center;
    padding: var(--space-3, 12px) var(--space-4, 16px);
    cursor: pointer;
    transition: background-color 0.2s ease;
    color: white;
  }
  
  .sidebar-user-menu-item:hover {
    background-color: rgba(255, 255, 255, 0.1);
  }
  
  .sidebar-user-menu-item-dark:hover {
    background-color: var(--dark-hover-bg);
  }
  
  .sidebar-user-menu-icon {
    margin-right: var(--space-3, 12px);
    display: flex;
    align-items: center;
    color: rgba(255, 255, 255, 0.8);
  }
  
  /* Debug section */
  .sidebar-debug {
    padding: var(--space-3, 12px) var(--space-4, 16px);
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    font-size: var(--font-size-sm, 0.875rem);
    background-color: rgba(0, 0, 0, 0.2);
    color: rgba(255, 255, 255, 0.9);
  }
  
  .sidebar-debug-title {
    font-weight: var(--font-weight-bold, 700);
    cursor: pointer;
    margin-bottom: var(--space-2, 8px);
    color: var(--color-primary, #3b82f6);
  }
  
  .sidebar-debug-item {
    margin-bottom: var(--space-2, 8px);
  }
  
  .sidebar-debug-content {
    background-color: rgba(0, 0, 0, 0.3);
    padding: var(--space-2, 8px);
    border-radius: var(--radius-md, 6px);
    overflow-x: auto;
    font-family: monospace;
    font-size: var(--font-size-xs, 0.75rem);
  }
  
  .sidebar-debug-content pre {
    margin: 0;
    white-space: pre-wrap;
  }
  
  /* Notification badge */
  .notification-badge {
    position: absolute;
    top: 8px;
    right: 16px;
    background-color: var(--color-primary, #3b82f6);
    color: white;
    font-size: var(--font-size-xs, 0.75rem);
    font-weight: var(--font-weight-bold, 700);
    border-radius: 50%;
    min-width: 18px;
    height: 18px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0 var(--space-1, 4px);
    box-shadow: 0 0 0 2px var(--bg-secondary, #1a1a1a);
  }
  
  .sidebar-inner-collapsed .notification-badge {
    top: 4px;
    right: 4px;
    min-width: 16px;
    height: 16px;
    font-size: calc(var(--font-size-xs, 0.75rem) - 2px);
  }
  
  /* Responsive adjustments */
  @media (max-width: 1080px) {
    .sidebar-inner:not(.sidebar-inner-mobile) {
      padding: var(--space-1, 4px);
    }
    
    .sidebar-logo {
      padding: var(--space-2, 8px) var(--space-2, 8px);
    }
    
    .sidebar-nav-item {
      padding: var(--space-3, 12px) var(--space-2, 8px);
    }
    
    .sidebar-tweet-btn {
      padding: var(--space-2, 8px);
    }
  }
  
  @media (max-width: 992px) {
    .sidebar-inner:not(.sidebar-inner-mobile):not(.sidebar-inner-collapsed) {
      max-width: 80px;
      align-items: center;
    }
    
    .sidebar-inner:not(.sidebar-inner-mobile):not(.sidebar-inner-collapsed) .sidebar-nav-item {
      justify-content: center;
      padding: var(--space-3, 12px) 0;
    }
    
    .sidebar-inner:not(.sidebar-inner-mobile):not(.sidebar-inner-collapsed) .sidebar-nav-icon {
      margin-right: 0;
    }
    
    .sidebar-inner:not(.sidebar-inner-mobile):not(.sidebar-inner-collapsed) .sidebar-nav-text {
      display: none;
    }
    
    .sidebar-inner:not(.sidebar-inner-mobile):not(.sidebar-inner-collapsed) .sidebar-tweet-btn {
      width: 48px;
      height: 48px;
      border-radius: 50%;
      padding: 0;
    }
    
    .sidebar-inner:not(.sidebar-inner-mobile):not(.sidebar-inner-collapsed) .sidebar-tweet-btn-icon {
      display: flex;
      align-items: center;
      justify-content: center;
    }
    
    .sidebar-inner:not(.sidebar-inner-mobile):not(.sidebar-inner-collapsed) .sidebar-tweet-btn-text {
      display: none;
    }
    
    .sidebar-inner:not(.sidebar-inner-mobile):not(.sidebar-inner-collapsed) .sidebar-profile-info {
      display: none;
    }
    
    .sidebar-inner:not(.sidebar-inner-mobile):not(.sidebar-inner-collapsed) .sidebar-profile-more {
      display: none;
    }
    
    .sidebar-inner:not(.sidebar-inner-mobile):not(.sidebar-inner-collapsed) .sidebar-profile {
      justify-content: center;
    }
    
    .sidebar-inner:not(.sidebar-inner-mobile):not(.sidebar-inner-collapsed) .sidebar-profile-avatar {
      margin-right: 0;
    }
    
    .sidebar-inner:not(.sidebar-inner-mobile):not(.sidebar-inner-collapsed) .logo-text {
      display: none;
    }
  }
  
  @media (max-width: 576px) {
    .sidebar-inner-mobile {
      padding: 0 var(--space-3, 12px) var(--space-3, 12px) var(--space-3, 12px);
    }
    
    .sidebar-nav-item {
      padding: var(--space-3, 12px) var(--space-2, 8px);
    }
    
    .sidebar-user-menu {
      width: calc(100% - var(--space-4, 16px));
      left: var(--space-2, 8px);
    }
  }
</style>