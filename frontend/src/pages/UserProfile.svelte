<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '../stores/routeStore';
  import OwnProfile from './OwnProfile.svelte';
  import LoadingSkeleton from '../components/common/LoadingSkeleton.svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { toastStore } from '../stores/toastStore';
  import { getUserById } from '../api/user';

  // Extract userId from the URL path
  let userId: string = '';
  let isLoading = true;
  let error: string | null = null;
  let username = '';
  let displayName = '';
  let avatar = '';

  // Subscribe to the page store to get URL parameters
  const unsubscribe = page.subscribe(($page) => {
    // Check if we have a userId from the route params
    if ($page.params.userId) {
      userId = $page.params.userId;
      
      // Only load user data once when userId changes
      if (isLoading) {
        loadUserBasicInfo(userId);
      }
    }
  });

  // Load basic user info for the layout
  async function loadUserBasicInfo(id: string) {
    if (!id) {
      error = 'Invalid user ID';
      isLoading = false;
      return;
    }
    
    isLoading = true;
    error = null;
    
    try {
      const userData = await getUserById(id);
      if (userData && userData.user) {
        username = userData.user.username || '';
        displayName = userData.user.display_name || '';
        avatar = userData.user.profile_picture_url || '';
      } else {
        error = 'User not found';
      }
    } catch (err) {
      console.error('Error loading user:', err);
      error = 'Failed to load user profile';
      toastStore.showToast('Failed to load user profile. Please try again.', 'error');
    } finally {
      isLoading = false;
    }
  }

  onMount(() => {
    // Fallback to direct URL parsing if routeStore didn't provide userId
    if (!userId) {
      const pathParts = window.location.pathname.split('/');
      const userIndex = pathParts.indexOf('user');
      if (userIndex >= 0 && userIndex + 1 < pathParts.length) {
        userId = pathParts[userIndex + 1];
        loadUserBasicInfo(userId);
      } else {
        error = 'Invalid user ID';
        isLoading = false;
      }
    }
  });

  // Clean up subscription when component is destroyed
  onMount(() => {
    return () => {
      unsubscribe();
    };
  });
</script>

{#if isLoading}
  <MainLayout>
    <LoadingSkeleton type="profile" />
  </MainLayout>
{:else if error}
  <MainLayout>
    <div class="flex flex-col items-center justify-center p-8">
      <h2 class="text-xl font-bold text-gray-700 dark:text-gray-300 mb-4">{error}</h2>
      <a href="/" class="text-blue-500 hover:underline">Return to home</a>
    </div>
  </MainLayout>
{:else}
  <OwnProfile {userId} />
{/if} 