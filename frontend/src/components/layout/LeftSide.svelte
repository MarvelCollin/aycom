<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import ThemeToggle from '../common/ThemeToggle.svelte';
  import { useTheme } from '../../hooks/useTheme';
  import { useAuth } from '../../hooks/useAuth';
  import { isAuthenticated, getUserId } from '../../utils/auth';
  import { toastStore } from '../../stores/toastStore';
  import { getProfile } from '../../api/user';
  import { onMount } from 'svelte';

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
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
                </svg>
              {:else if item.icon === 'hash'}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
                </svg>
              {:else if item.icon === 'bell'}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                </svg>
              {:else if item.icon === 'mail'}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>
              {:else if item.icon === 'bookmark'}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
                </svg>
              {:else if item.icon === 'users'}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                </svg>
              {:else if item.icon === 'user'}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
              {:else if item.icon === 'settings'}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
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
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 mx-auto" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
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
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 {isDarkMode ? 'text-gray-300' : 'text-gray-700'}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h.01M12 12h.01M19 12h.01M6 12a1 1 0 11-2 0 1 1 0 012 0zm7 0a1 1 0 11-2 0 1 1 0 012 0zm7 0a1 1 0 11-2 0 1 1 0 012 0z" />
        </svg>
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
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>
                {userDetails.email}
              </div>
              
              {#if userDetails.isVerified}
                <div class="flex items-center mb-1">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  Verified account
                </div>
              {/if}
              
              {#if userDetails.joinDate}
                <div class="flex items-center">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
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
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
              </svg>
              Log out
            </button>
          {:else}
            <a
              href="/login"
              class="flex items-center w-full px-4 py-3 {isDarkMode ? 'text-white hover:bg-gray-800' : 'text-black hover:bg-gray-100'}"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
              </svg>
              Sign in
            </a>
          {/if}
        </div>
      </div>
    {/if}
  </div>
</div>