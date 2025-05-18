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

  export let username = "";
  export let displayName = "";
  export let avatar = "https://secure.gravatar.com/avatar/0?d=mp";
  
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
  
  async function fetchUserProfile() {
    if (!isAuthenticated()) {
      console.log('User not authenticated, skipping profile fetch');
      return;
    }
    
    console.log('Fetching user profile...');
    try {
      const response = await getProfile();
      apiResponse = response;
      
      const userData = response.user || (response.data && response.data.user);
      
      if (userData) {
        userDetails = {
          username: userData.username || username,
          displayName: userData.name || userData.display_name || displayName,
          avatar: userData.profile_picture_url || avatar,
          userId: userData.id || getUserId() || '',
          email: userData.email || '',
          isVerified: userData.is_verified || false,
          joinDate: userData.created_at ? new Date(userData.created_at).toLocaleDateString() : ''
        };
        
        isAdmin = isAuthenticated();
        
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
  
  const navigationItems = [
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
  
  let currentPath = window.location.pathname;
  
  onMount(() => {
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
    
    return () => {
      clearInterval(intervalId);
    };
  });
</script>

<div class="sidebar {isDarkMode ? 'sidebar-dark' : ''}">
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