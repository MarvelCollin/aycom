<script lang="ts">
  import { onMount } from 'svelte';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useAuth } from '../hooks/useAuth';
  import { useTheme } from '../hooks/useTheme';
  import type { IAuthStore } from '../interfaces/IAuth';
  import { createLoggerWithPrefix } from '../utils/logger';
  import { toastStore } from '../stores/toastStore';
  
  const logger = createLoggerWithPrefix('CreateCommunity');
  
  // Auth and theme
  const { getAuthState } = useAuth();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: authState = getAuthState ? (getAuthState() as IAuthStore) : { userId: null, isAuthenticated: false, accessToken: null, refreshToken: null };
  $: isDarkMode = $theme === 'dark';
  $: sidebarUsername = authState?.userId ? `User_${authState.userId.substring(0, 4)}` : '';
  $: sidebarDisplayName = authState?.userId ? `User ${authState.userId.substring(0, 4)}` : '';
  $: sidebarAvatar = 'https://secure.gravatar.com/avatar/0?d=mp'; // Default avatar with proper image URL
  
  // Available categories for selection
  const availableCategories = [
    'Gaming', 'Sports', 'Food', 'Technology', 'Art', 'Music', 
    'Movies', 'Books', 'Fitness', 'Travel', 'Fashion', 'Education',
    'Science', 'Health', 'Business', 'Politics', 'News', 'Photography',
    'Lifestyle', 'Entertainment', 'Pets', 'Environment', 'DIY', 'Finance'
  ];
  
  // Form data
  let communityName = '';
  let description = '';
  let icon: File | null = null;
  let iconPreview: string | null = null;
  let selectedCategories: string[] = [];
  let banner: File | null = null;
  let bannerPreview: string | null = null;
  let rules = '';
  
  // Form state
  let isSubmitting = false;
  let isSuccess = false;
  let errors: Record<string, string> = {};
  
  // Authentication check
  function checkAuth() {
    if (!authState.isAuthenticated) {
      toastStore.showToast('You need to log in to create a community', 'warning');
      window.location.href = '/login';
      return false;
    }
    return true;
  }
  
  // Handle icon file selection
  function handleIconChange(event: Event) {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      icon = input.files[0];
      
      // Create a preview
      const reader = new FileReader();
      reader.onload = e => {
        iconPreview = e.target?.result as string;
      };
      reader.readAsDataURL(icon);
    }
  }
  
  // Handle banner file selection
  function handleBannerChange(event: Event) {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      banner = input.files[0];
      
      // Create a preview
      const reader = new FileReader();
      reader.onload = e => {
        bannerPreview = e.target?.result as string;
      };
      reader.readAsDataURL(banner);
    }
  }
  
  // Toggle category selection
  function toggleCategory(category: string) {
    if (selectedCategories.includes(category)) {
      selectedCategories = selectedCategories.filter(c => c !== category);
    } else {
      if (selectedCategories.length < 5) { // Limit to 5 categories
        selectedCategories = [...selectedCategories, category];
      } else {
        toastStore.showToast('You can select up to 5 categories', 'warning');
      }
    }
  }
  
  // Validate form
  function validateForm(): boolean {
    errors = {};
    
    // Community name validation
    if (!communityName.trim()) {
      errors.communityName = 'Community name is required';
    } else if (communityName.length < 3) {
      errors.communityName = 'Community name must be at least 3 characters';
    } else if (communityName.length > 50) {
      errors.communityName = 'Community name cannot exceed 50 characters';
    }
    
    // Description validation
    if (!description.trim()) {
      errors.description = 'Description is required';
    } else if (description.length < 30) {
      errors.description = 'Description must be at least 30 characters';
    } else if (description.length > 500) {
      errors.description = 'Description cannot exceed 500 characters';
    }
    
    // Icon validation
    if (!icon) {
      errors.icon = 'Community icon is required';
    }
    
    // Categories validation
    if (selectedCategories.length === 0) {
      errors.categories = 'At least one category is required';
    }
    
    // Banner validation
    if (!banner) {
      errors.banner = 'Community banner is required';
    }
    
    // Rules validation
    if (!rules.trim()) {
      errors.rules = 'Community rules are required';
    } else if (rules.length < 50) {
      errors.rules = 'Rules must be at least 50 characters';
    }
    
    return Object.keys(errors).length === 0;
  }
  
  // Submit form
  async function handleSubmit() {
    if (!validateForm()) {
      // Scroll to first error
      const firstErrorField = Object.keys(errors)[0];
      const element = document.getElementById(firstErrorField);
      if (element) {
        element.scrollIntoView({ behavior: 'smooth', block: 'center' });
      }
      return;
    }
    
    isSubmitting = true;
    
    try {
      // In a real implementation, this would be an API call to create a community
      // For example:
      // const formData = new FormData();
      // formData.append('name', communityName);
      // formData.append('description', description);
      // formData.append('icon', icon);
      // formData.append('categories', JSON.stringify(selectedCategories));
      // formData.append('banner', banner);
      // formData.append('rules', rules);
      // const response = await fetch('/api/communities', {
      //   method: 'POST',
      //   body: formData
      // });
      
      // Simulate API response with a delay
      await new Promise(resolve => setTimeout(resolve, 1500));
      
      logger.debug('Community creation request submitted', {
        name: communityName,
        categories: selectedCategories
      });
      
      // Success
      isSuccess = true;
      toastStore.showToast('Community creation request submitted for approval', 'success');
      
    } catch (error) {
      console.error('Error creating community:', error);
      toastStore.showToast('Failed to create community. Please try again.', 'error');
    } finally {
      isSubmitting = false;
    }
  }
  
  // Navigate back to communities
  function navigateToCommunities() {
    window.location.href = '/communities';
  }
  
  onMount(() => {
    logger.debug('Create Community page mounted', { authState });
    checkAuth();
  });
</script>

<MainLayout
  username={sidebarUsername}
  displayName={sidebarDisplayName}
  avatar={sidebarAvatar}
  on:toggleComposeModal={() => {}}
>
  <div class="min-h-screen border-x border-gray-200 dark:border-gray-800">
    <!-- Header -->
    <div class="sticky top-0 z-10 bg-white/80 dark:bg-black/80 backdrop-blur-md border-b border-gray-200 dark:border-gray-800 px-4 py-3">
      <div class="flex items-center">
        <button
          class="mr-4 p-2 rounded-full hover:bg-gray-200 dark:hover:bg-gray-800"
          on:click={navigateToCommunities}
          aria-label="Back to Communities"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
        </button>
        <h1 class="text-xl font-bold">Create Community</h1>
      </div>
    </div>
    
    <!-- Content -->
    <div class="p-4">
      {#if isSuccess}
        <div class="bg-green-50 dark:bg-green-900/20 rounded-lg p-6 text-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 mx-auto text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <h2 class="text-2xl font-bold mt-4 mb-2">Request Submitted!</h2>
          <p class="text-gray-600 dark:text-gray-300 mb-6 max-w-md mx-auto">
            Your community creation request has been submitted and is pending admin approval. 
            We'll notify you once it's approved.
          </p>
          <button
            class="bg-blue-500 hover:bg-blue-600 text-white px-6 py-2 rounded-full font-medium"
            on:click={navigateToCommunities}
          >
            Back to Communities
          </button>
        </div>
      {:else}
        <div class="max-w-3xl mx-auto">
          <p class="text-gray-600 dark:text-gray-300 mb-6">
            Create your own community to connect with people who share your interests. 
            All communities are subject to admin approval before they are created.
          </p>
          
          <form on:submit|preventDefault={handleSubmit} class="space-y-6">
            <!-- Community Name -->
            <div>
              <label for="communityName" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Community Name <span class="text-red-500">*</span>
              </label>
              <input
                id="communityName"
                type="text"
                bind:value={communityName}
                class="w-full px-4 py-2 rounded-lg border {errors.communityName ? 'border-red-500' : 'border-gray-300 dark:border-gray-600'} bg-white dark:bg-gray-800"
                placeholder="Enter community name"
              />
              {#if errors.communityName}
                <p class="mt-1 text-sm text-red-500">{errors.communityName}</p>
              {/if}
            </div>
            
            <!-- Description -->
            <div>
              <label for="description" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Description <span class="text-red-500">*</span>
              </label>
              <textarea
                id="description"
                bind:value={description}
                rows="4"
                class="w-full px-4 py-2 rounded-lg border {errors.description ? 'border-red-500' : 'border-gray-300 dark:border-gray-600'} bg-white dark:bg-gray-800"
                placeholder="Describe what your community is about"
              ></textarea>
              {#if errors.description}
                <p class="mt-1 text-sm text-red-500">{errors.description}</p>
              {:else}
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{description.length}/500 characters</p>
              {/if}
            </div>
            
            <!-- Icon Upload -->
            <div>
              <label for="icon" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Community Icon <span class="text-red-500">*</span>
              </label>
              <div class="flex items-center space-x-4">
                <div class="w-24 h-24 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center overflow-hidden border-2 {errors.icon ? 'border-red-500' : 'border-gray-300 dark:border-gray-600'}">
                  {#if iconPreview}
                    <img src={iconPreview} alt="Icon preview" class="w-full h-full object-cover" />
                  {:else}
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                    </svg>
                  {/if}
                </div>
                <div class="flex-1">
                  <input
                    id="icon"
                    type="file"
                    accept="image/*"
                    on:change={handleIconChange}
                    class="hidden"
                  />
                  <label
                    for="icon"
                    class="inline-block px-4 py-2 bg-gray-200 dark:bg-gray-700 rounded-lg cursor-pointer hover:bg-gray-300 dark:hover:bg-gray-600 transition"
                  >
                    Choose Icon
                  </label>
                  <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">PNG, JPG, or GIF. Recommended size 400x400 pixels.</p>
                  {#if errors.icon}
                    <p class="mt-1 text-sm text-red-500">{errors.icon}</p>
                  {/if}
                </div>
              </div>
            </div>
            
            <!-- Categories -->
            <div>
              <label id="categories-label" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Categories <span class="text-red-500">*</span> <span class="text-sm font-normal text-gray-500">(Select up to 5)</span>
              </label>
              <div class="flex flex-wrap gap-2 mt-2" role="group" aria-labelledby="categories-label">
                {#each availableCategories as category}
                  <button
                    type="button"
                    class="px-3 py-1 rounded-full text-sm {selectedCategories.includes(category) ? 'bg-blue-500 text-white' : 'bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200'}"
                    on:click={() => toggleCategory(category)}
                    aria-pressed={selectedCategories.includes(category)}
                  >
                    {category}
                  </button>
                {/each}
              </div>
              {#if errors.categories}
                <p class="mt-1 text-sm text-red-500">{errors.categories}</p>
              {/if}
            </div>
            
            <!-- Banner Upload -->
            <div>
              <label for="banner" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Community Banner <span class="text-red-500">*</span>
              </label>
              <div class="border-2 rounded-lg {errors.banner ? 'border-red-500' : 'border-gray-300 dark:border-gray-600'} overflow-hidden">
                <div class="aspect-[3/1] bg-gray-200 dark:bg-gray-700 flex items-center justify-center">
                  {#if bannerPreview}
                    <img src={bannerPreview} alt="Banner preview" class="w-full h-full object-cover" />
                  {:else}
                    <div class="text-center p-6">
                      <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 mx-auto text-gray-400 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                      </svg>
                      <p class="text-gray-500 dark:text-gray-400">Recommended size: 1200x400 pixels</p>
                    </div>
                  {/if}
                </div>
                <div class="p-3 bg-gray-100 dark:bg-gray-800">
                  <input
                    id="banner"
                    type="file"
                    accept="image/*"
                    on:change={handleBannerChange}
                    class="hidden"
                  />
                  <label
                    for="banner"
                    class="inline-block px-4 py-2 bg-gray-200 dark:bg-gray-700 rounded-lg cursor-pointer hover:bg-gray-300 dark:hover:bg-gray-600 transition"
                  >
                    Choose Banner
                  </label>
                  {#if errors.banner}
                    <p class="mt-1 text-sm text-red-500">{errors.banner}</p>
                  {/if}
                </div>
              </div>
            </div>
            
            <!-- Rules -->
            <div>
              <label for="rules" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Community Rules <span class="text-red-500">*</span>
              </label>
              <textarea
                id="rules"
                bind:value={rules}
                rows="6"
                class="w-full px-4 py-2 rounded-lg border {errors.rules ? 'border-red-500' : 'border-gray-300 dark:border-gray-600'} bg-white dark:bg-gray-800"
                placeholder="Enter the rules for your community. These help maintain a positive environment and let members know what's expected."
              ></textarea>
              {#if errors.rules}
                <p class="mt-1 text-sm text-red-500">{errors.rules}</p>
              {/if}
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                Include guidelines for posting, commenting, and interacting within the community.
              </p>
            </div>
            
            <!-- Submit Button -->
            <div class="flex justify-end pt-4">
              <button
                type="button"
                class="px-6 py-2 mr-3 border border-gray-300 dark:border-gray-600 rounded-full font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition"
                on:click={navigateToCommunities}
                disabled={isSubmitting}
              >
                Cancel
              </button>
              <button
                type="submit"
                class="px-6 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-full font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                disabled={isSubmitting}
              >
                {isSubmitting ? 'Submitting...' : 'Submit for Approval'}
              </button>
            </div>
          </form>
        </div>
      {/if}
    </div>
  </div>
</MainLayout>

<style>
  /* Prevent text overflow */
  textarea {
    resize: vertical;
  }
</style>
