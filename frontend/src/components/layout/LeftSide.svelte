<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import ThemeToggle from '../common/ThemeToggle.svelte';
  import { useTheme } from '../../hooks/useTheme';
  import { useAuth } from '../../hooks/useAuth';
  import { isAuthenticated, getUserId } from '../../utils/auth';
  import { toastStore } from '../../stores/toastStore';
  import { getProfile } from '../../api/user';
  import { onMount } from 'svelte';

  // Import Feather icons
  import HomeIcon from 'svelte-feather-icons/src/icons/HomeIcon.svelte';
  import HashIcon from 'svelte-feather-icons/src/icons/HashIcon.svelte';
  import BellIcon from 'svelte-feather-icons/src/icons/BellIcon.svelte';
  import MailIcon from 'svelte-feather-icons/src/icons/MailIcon.svelte';
  import BookmarkIcon from 'svelte-feather-icons/src/icons/BookmarkIcon.svelte';
  import UsersIcon from 'svelte-feather-icons/src/icons/UsersIcon.svelte';
  import UserIcon from 'svelte-feather-icons/src/icons/UserIcon.svelte';
  import SettingsIcon from 'svelte-feather-icons/src/icons/SettingsIcon.svelte';
  import PlusIcon from 'svelte-feather-icons/src/icons/PlusIcon.svelte';
  import MoreHorizontalIcon from 'svelte-feather-icons/src/icons/MoreHorizontalIcon.svelte';
  import LogOutIcon from 'svelte-feather-icons/src/icons/LogOutIcon.svelte';
  import CalendarIcon from 'svelte-feather-icons/src/icons/CalendarIcon.svelte';
  import CheckCircleIcon from 'svelte-feather-icons/src/icons/CheckCircleIcon.svelte';
  import LogInIcon from 'svelte-feather-icons/src/icons/LogInIcon.svelte';

  // Props
  export let username = "";
  export let displayName = "";
  export let avatar = "https://secure.gravatar.com/avatar/0?d=mp";
  
  // Get theme from the store
  const { theme } = useTheme();
  $: isDarkMode = $theme === 'dark';
  
  // Get auth state
  const { getAuthState, logout, getAuthToken } = useAuth();
  let authState = getAuthState();
  
  // Add debug flag
  let debugging = false;
  let apiResponse = null;
  
  // User profile data
  let userDetails = {
    username: username || 'guest',
    displayName: displayName || 'Guest User',
    avatar: avatar,
    userId: getUserId() || '',
    email: '',
    isVerified: false,
    joinDate: ''
  };
  
  // Fetch user profile
  async function fetchUserProfile() {
    if (!isAuthenticated()) {
      console.log('User not authenticated, skipping profile fetch');
      return;
    }
    
    console.log('Fetching user profile...');
    try {
      const response = await getProfile();
      apiResponse = response; // Store for debugging
      
      // Check for both possible response structures (direct or nested)
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
        
        // Update the component props in case they're used elsewhere
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
  
  // Toggle debug info
  function toggleDebug() {
    debugging = !debugging;
  }
  
  // Handle logout
  async function handleLogout() {
    try {
      await logout();
      window.location.href = '/login';
    } catch (err) {
      console.error('Error during logout:', err);
      toastStore.showToast('Logout failed. Please try again.', 'error');
    }
  }
  
  // Event dispatcher
  const dispatch = createEventDispatcher();
  
  // Navigation items - updated according to available pages
  const navigationItems = [
    { label: "Feed", path: "/feed", icon: "home" },
    { label: "Explore", path: "/explore", icon: "hash" },
    { label: "Notifications", path: "/notifications", icon: "bell" },
    { label: "Messages", path: "/messages", icon: "mail" },
    { label: "Bookmarks", path: "/bookmarks", icon: "bookmark" },
    { label: "Communities", path: "/communities", icon: "users" },
    { label: "Profile", path: "/profile", icon: "user" },
  ];
  
  // Toggle user menu
  let showUserMenu = false;
  function toggleUserMenu() {
    showUserMenu = !showUserMenu;
    
    // If we're showing the menu and authenticated, update user details
    if (showUserMenu && isAuthenticated()) {
      fetchUserProfile();
    }
  }
  
  // Handle compose tweet modal action
  function handleToggleComposeModal() {
    dispatch('toggleComposeModal');
  }
  
  // Get current path for active state
  let currentPath = window.location.pathname;
  
  onMount(() => {
    // Try to fetch user profile
    if (isAuthenticated()) {
      console.log('User is authenticated, fetching profile on mount');
      fetchUserProfile();
    } else {
      console.log('User is not authenticated on mount');
    }
    
    // Set up polling to check authentication and refresh profile
    const intervalId = setInterval(() => {
      if (isAuthenticated() && userDetails.username === 'guest') {
        console.log('User is authenticated but still shows as guest, refreshing profile');
        fetchUserProfile();
      }
    }, 5000); // Check every 5 seconds
    
    // Return cleanup function
    return () => {
      clearInterval(intervalId);
    };
  });
</script>

<div class="flex flex-col h-full py-2 px-2 {isDarkMode ? 'text-white' : 'text-black'}">
  <!-- Logo -->
  <div class="px-3 mb-4">
    <a href="/" class="flex items-center justify-center md:justify-start p-3 rounded-full hover:bg-gray-200 dark:hover:bg-gray-800">
      <div class="text-3xl font-bold text-blue-500">AY</div>
    </a>
  </div>

  <!-- Navigation Menu -->
  <nav class="flex-1">
    <ul class="space-y-1">
      {#each navigationItems as item}
        <li>
          <a 
            href={item.path} 
            class="flex items-center px-4 py-3 rounded-full {currentPath === item.path ? 'font-bold' : 'font-normal'} {isDarkMode ? 'text-white hover:bg-gray-800' : 'text-black hover:bg-gray-200'}"
          >
            <!-- Icon -->
            <div class="flex items-center justify-center w-6 h-6">
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
              {:else if item.icon === 'user'}
                <UserIcon size="24" />
              {:else if item.icon === 'settings'}
                <SettingsIcon size="24" />
              {/if}
            </div>
            <!-- Label -->
            <span class="hidden md:block text-xl ml-4">{item.label}</span>
          </a>
        </li>
      {/each}
    </ul>

    <!-- Post Button -->
    <div class="mt-4 px-3">
      <button 
        class="w-full py-3 bg-blue-500 text-white rounded-full font-bold hover:bg-blue-600"
        on:click={handleToggleComposeModal}
      >
        <span class="md:hidden">
          <PlusIcon size="24" />
        </span>
        <span class="hidden md:block text-lg">Post</span>
      </button>
    </div>

    <div class="mt-4 px-3">
      <div class="flex justify-center md:justify-start p-2">
        <ThemeToggle size="md" />
      </div>
    </div>
  </nav>

  <div class="mt-4 px-3 mb-4 relative">
    <!-- User profile button with auth status indicator -->
    <button 
      class="flex items-center w-full p-3 rounded-full relative {isDarkMode ? 'bg-gray-800 hover:bg-gray-800' : 'hover:bg-gray-200'}"
      on:click={toggleUserMenu}
      on:dblclick={toggleDebug}
    >
      <!-- Auth indicator dot -->
      <div class="absolute -top-1 -right-1 w-4 h-4 rounded-full {isAuthenticated() ? 'bg-green-500' : 'bg-gray-500'} border-2 {isDarkMode ? 'border-gray-900' : 'border-white'}"></div>
      
      <div class="w-10 h-10 rounded-full {isDarkMode ? 'bg-gray-700' : 'bg-gray-300'} flex items-center justify-center overflow-hidden">
        {#if typeof userDetails.avatar === 'string' && userDetails.avatar.startsWith('http')}
          <img src={userDetails.avatar} alt={userDetails.username} class="w-full h-full object-cover" />
        {:else}
          <span class="text-lg">{userDetails.avatar}</span>
        {/if}
      </div>
      <div class="hidden md:block ml-3 flex-1 text-left">
        <p class="font-bold text-sm {isDarkMode ? 'text-white' : 'text-black'}">{userDetails.displayName}</p>
        <p class="text-sm {isDarkMode ? 'text-gray-300' : 'text-gray-700'}">@{userDetails.username}</p>
      </div>
      <div class="hidden md:flex">
        <MoreHorizontalIcon size="20" />
      </div>
    </button>
    
    <!-- User Menu Dropdown -->
    {#if showUserMenu}
      <div 
        class="absolute bottom-20 left-2 w-72 rounded-lg shadow border {isDarkMode ? 'bg-gray-900 border-gray-800' : 'bg-white border-gray-200'} z-50"
      >
        <div class="py-3 px-4 border-b {isDarkMode ? 'border-gray-800' : 'border-gray-200'}">
          <div class="flex items-center mb-2">
            <div class="w-12 h-12 rounded-full {isDarkMode ? 'bg-gray-700' : 'bg-gray-300'} flex items-center justify-center overflow-hidden mr-3">
              {#if typeof userDetails.avatar === 'string' && userDetails.avatar.startsWith('http')}
                <img src={userDetails.avatar} alt={userDetails.username} class="w-full h-full object-cover" />
              {:else}
                <span class="text-lg">{userDetails.avatar}</span>
              {/if}
            </div>
            
            <div>
              <p class="font-bold {isDarkMode ? 'text-white' : 'text-black'}">{userDetails.displayName}</p>
              <p class="text-sm {isDarkMode ? 'text-gray-300' : 'text-gray-700'}">@{userDetails.username}</p>
            </div>
          </div>
          
          {#if isAuthenticated()}
            <div class="text-xs {isDarkMode ? 'text-gray-400' : 'text-gray-600'} mt-2">
              <div class="flex items-center mb-1">
                <MailIcon size="16" class="mr-1" />
                {userDetails.email}
              </div>
              
              {#if userDetails.isVerified}
                <div class="flex items-center mb-1">
                  <CheckCircleIcon size="16" class="mr-1 text-green-500" />
                  Verified account
                </div>
              {/if}
              
              {#if userDetails.joinDate}
                <div class="flex items-center">
                  <CalendarIcon size="16" class="mr-1" />
                  Joined {userDetails.joinDate}
                </div>
              {/if}
            </div>
          {/if}
          
          <!-- Debug info (double-click user profile to toggle) -->
          {#if debugging}
            <div class="mt-2 p-2 bg-gray-100 dark:bg-gray-800 rounded text-xs overflow-auto max-h-36">
              <p class="font-bold">Debug Info:</p>
              <p>Auth state: {isAuthenticated() ? 'Authenticated' : 'Not authenticated'}</p>
              <p>User ID: {getUserId() || 'Not found'}</p>
              <p>Has token: {!!getAuthToken()}</p>
              <p>API response:</p>
              <pre>{JSON.stringify(apiResponse, null, 2)}</pre>
              <button 
                class="mt-1 bg-blue-500 text-white px-2 py-1 rounded text-xs"
                on:click|stopPropagation={() => fetchUserProfile()}
              >
                Refresh Profile
              </button>
            </div>
          {/if}
        </div>
        
        <div class="py-2">
          {#if isAuthenticated()}
            <button
              class="flex items-center w-full px-4 py-3 {isDarkMode ? 'text-white hover:bg-gray-800' : 'text-black hover:bg-gray-100'}"
              on:click={handleLogout}
            >
              <LogOutIcon size="20" class="mr-3" />
              Log out
            </button>
          {:else}
            <a
              href="/login"
              class="flex items-center w-full px-4 py-3 {isDarkMode ? 'text-white hover:bg-gray-800' : 'text-black hover:bg-gray-100'}"
            >
              <LogInIcon size="20" class="mr-3" />
              Log in
            </a>
          {/if}
        </div>
      </div>
    {/if}
  </div>
</div>