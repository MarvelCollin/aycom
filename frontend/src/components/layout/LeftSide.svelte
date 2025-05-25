<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import ThemeToggle from '../common/ThemeToggle.svelte';
  import { useTheme } from '../../hooks/useTheme';
  import { useAuth } from '../../hooks/useAuth';
  import { isAuthenticated, getUserId } from '../../utils/auth';
  import { toastStore } from '../../stores/toastStore';
  import { getProfile, checkAdminStatus } from '../../api/user';
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
  
  function debugAdminStatus() {
    console.log('DEBUG: Checking admin status directly from localStorage');
    try {
      const authData = localStorage.getItem('auth');
      if (authData) {
        const auth = JSON.parse(authData);
        console.log('AUTH DATA:', auth);
        if (auth.is_admin === true) {
          console.log('DEBUG: User is admin according to localStorage');
          isAdmin = true;
        }
      }
    } catch (e) {
      console.error('DEBUG: Error checking admin status:', e);
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
        isAdmin = userData.is_admin === true || (authState && authState.is_admin === true);
        
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
        if (authState && authState.is_admin === true) {
          isAdmin = true;
          console.log('No API user data but user is admin based on auth store');
        }
      }
    } catch (err) {
      console.error('Failed to fetch user profile:', err);
      
      // Even on error, still check auth store for admin status
      const authState = getAuthState();
      if (authState && authState.is_admin === true) {
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
  
  onMount(() => {
    currentPath = window.location.pathname;
    
    // Run debug admin status check
    debugAdminStatus();
    
    // Check admin status from auth store as early as possible
    const authState = getAuthState();
    if (authState && authState.is_admin === true) {
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
        <div class="sidebar-profile-name">{userDetails.displayName}</div>
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
    max-width: var(--sidebar-width);
    padding: var(--space-2);
    transition: all 0.3s ease;
    position: relative;
  }
  
  .sidebar-inner-dark {
    color: var(--text-primary-dark);
  }
  
  .sidebar-inner-collapsed {
    max-width: var(--sidebar-collapsed-width);
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
    padding: var(--space-2) var(--space-2) var(--space-4) var(--space-2);
  }
  
  .sidebar-close-btn {
    background: transparent;
    border: none;
    color: var(--text-primary);
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
    background-color: var(--bg-hover);
  }
  
  .sidebar-logo {
    display: flex;
    align-items: center;
    padding: var(--space-2) var(--space-4);
    margin-bottom: var(--space-4);
  }
  
  .logo-img {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    margin-right: var(--space-2);
  }
  
  .logo-text {
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-bold);
  }
  
  .sidebar-nav {
    flex: 1;
    display: flex;
    flex-direction: column;
    margin-bottom: var(--space-4);
    overflow-y: auto;
  }
  
  .sidebar-nav-list {
    list-style: none;
    padding: 0;
    margin: 0 0 var(--space-4) 0;
  }
  
  .sidebar-nav-item {
    display: flex;
    align-items: center;
    padding: var(--space-3) var(--space-4);
    border-radius: var(--radius-full);
    margin-bottom: var(--space-2);
    text-decoration: none;
    color: var(--text-primary);
    font-weight: var(--font-weight-medium);
    transition: background-color 0.2s ease;
  }
  
  .sidebar-nav-item:hover {
    background-color: var(--bg-hover);
  }
  
  .sidebar-nav-item.active {
    font-weight: var(--font-weight-bold);
    color: var(--color-primary);
  }
  
  .sidebar-nav-item.dark {
    color: var(--text-primary-dark);
  }
  
  .sidebar-nav-item.active.dark {
    color: var(--color-primary);
  }
  
  .sidebar-nav-icon {
    margin-right: var(--space-3);
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
  }
  
  .sidebar-inner-collapsed .sidebar-nav-item {
    justify-content: center;
    padding: var(--space-3) 0;
  }
  
  .sidebar-inner-collapsed .sidebar-nav-icon {
    margin-right: 0;
  }
  
  .sidebar-tweet-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--space-3);
    border-radius: var(--radius-full);
    background-color: var(--color-primary);
    color: white;
    border: none;
    font-weight: var(--font-weight-bold);
    transition: all 0.2s ease;
    cursor: pointer;
    width: 90%;
    margin: 0 auto;
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
    transition: background-color 0.2s ease;
  }
  
  .sidebar-profile:hover {
    background-color: var(--bg-hover);
  }
  
  .sidebar-inner-collapsed .sidebar-profile {
    justify-content: center;
    width: 100%;
    padding: var(--space-2) 0;
  }
  
  .sidebar-profile-avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    overflow: hidden;
    margin-right: var(--space-3);
    flex-shrink: 0;
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
    margin-right: var(--space-2);
  }
  
  .sidebar-inner-collapsed .sidebar-profile-info {
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
  
  .sidebar-inner-collapsed .sidebar-profile-more {
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
  
  .sidebar-debug-item {
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
  
  /* Responsive adjustments */
  @media (max-width: 1080px) {
    .sidebar-inner:not(.sidebar-inner-mobile) {
      padding: var(--space-1);
    }
    
    .sidebar-logo {
      padding: var(--space-2) var(--space-2);
    }
    
    .sidebar-nav-item {
      padding: var(--space-3) var(--space-2);
    }
    
    .sidebar-tweet-btn {
      padding: var(--space-2);
    }
  }
  
  @media (max-width: 576px) {
    .sidebar-inner-mobile {
      padding: 0 var(--space-3) var(--space-3) var(--space-3);
    }
    
    .sidebar-nav-item {
      padding: var(--space-3) var(--space-2);
    }
    
    .sidebar-user-menu {
      width: calc(100% - var(--space-4));
      left: var(--space-2);
    }
  }
</style>