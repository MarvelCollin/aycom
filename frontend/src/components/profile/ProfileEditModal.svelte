<!-- ProfileEditModal.svelte -->
<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { uploadProfilePicture, uploadBanner } from '../../api/user';
  import { toastStore } from '../../stores/toastStore';
  import type { IUserProfile } from '../../interfaces/IUser';
  import { useTheme } from '../../hooks/useTheme';
  
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  export let profile: IUserProfile | null = null;
  export const isOpen = true;
  
  // Local form data that we can safely bind to
  let formData = {
    displayName: '',
    bio: '',
    email: '',
    dateOfBirth: '',
    gender: ''
  };
  
  let profilePictureFile: File | null = null;
  let bannerFile: File | null = null;
  let profilePicturePreview: string | null = null;
  let bannerPreview: string | null = null;
  let isUploading = false;
  let errorMessage = '';
  
  // Initialize form data when profile changes
  $: if (profile) {
    formData = {
      displayName: profile.displayName || '',
      bio: profile.bio || '',
      email: profile.email || '',
      dateOfBirth: profile.dateOfBirth || '',
      gender: profile.gender || ''
    };
  }
  
  onMount(() => {
    if (profile) {
      profilePicturePreview = profile.profile_picture || null;
      bannerPreview = profile.banner || null;
    }
  });
  
  function handleProfilePictureChange(e: Event) {
    const input = e.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      profilePictureFile = input.files[0];
      
      // Create preview
      const reader = new FileReader();
      reader.onload = (e) => {
        profilePicturePreview = e.target?.result as string;
      };
      reader.readAsDataURL(profilePictureFile);
    }
  }
  
  function handleBannerChange(e: Event) {
    const input = e.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      bannerFile = input.files[0];
      
      // Create preview
      const reader = new FileReader();
      reader.onload = (e) => {
        bannerPreview = e.target?.result as string;
      };
      reader.readAsDataURL(bannerFile);
    }
  }
  
  function removeProfilePicture() {
    profilePictureFile = null;
    profilePicturePreview = null;
  }
  
  function removeBanner() {
    bannerFile = null;
    bannerPreview = null;
  }
  
  async function handleSave() {
    // If we have form changes, update those as well
    if (profile) {
      // Update profile with form data
      if (formData.displayName !== profile.displayName || 
          formData.bio !== profile.bio ||
          formData.email !== profile.email ||
          formData.dateOfBirth !== profile.dateOfBirth ||
          formData.gender !== profile.gender) {
        // Dispatch event to parent to handle the profile update
        dispatch('updateProfile', formData);
      }
    }
    
    if (!profilePictureFile && !bannerFile) {
      dispatch('close');
      return;
    }
    
    isUploading = true;
    errorMessage = '';
    
    try {
      // Upload profile picture if selected
      if (profilePictureFile) {
        const profileResult = await uploadProfilePicture(profilePictureFile);
        if (profileResult && profileResult.success) {
          toastStore.showToast('Profile picture updated successfully', 'success');
          dispatch('profilePictureUpdated', { url: profileResult.url });
        } else {
          throw new Error('Failed to upload profile picture');
        }
      }
      
      // Upload banner if selected
      if (bannerFile) {
        const bannerResult = await uploadBanner(bannerFile);
        if (bannerResult && bannerResult.success) {
          toastStore.showToast('Banner updated successfully', 'success');
          dispatch('bannerUpdated', { url: bannerResult.url });
        } else {
          throw new Error('Failed to upload banner');
        }
      }
      
      // Close the modal after successful upload
      dispatch('close');
      
    } catch (err) {
      console.error('Error updating profile media:', err);
      errorMessage = err instanceof Error ? err.message : 'Failed to update profile media';
      toastStore.showToast(errorMessage, 'error');
    } finally {
      isUploading = false;
    }
  }
  
  function handleClose() {
    dispatch('close');
  }
  
  // Keyboard event handler for modal
  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      handleClose();
    }
  }
</script>

<div 
  class="fixed inset-0 bg-black/60 flex items-center justify-center z-50" 
  on:click={handleClose}
  on:keydown={handleKeyDown}
  role="dialog" 
  aria-modal="true"
  aria-labelledby="modal-title"
  tabindex="0"
>
  <div 
    class="bg-white dark:bg-gray-900 rounded-xl w-full max-w-xl mx-4 max-h-[90vh] overflow-y-auto"
    on:click|stopPropagation
    on:keydown|stopPropagation
    role="document"
  >
    <!-- Header -->
    <div class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
      <div class="flex items-center gap-4">
        <button 
          on:click={handleClose} 
          class="text-gray-600 dark:text-gray-300 hover:text-gray-800 dark:hover:text-gray-100"
          aria-label="Close modal"
        >
          <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
        <h2 id="modal-title" class="text-xl font-bold dark:text-white">Edit profile</h2>
      </div>
      
      <button 
        class="py-1.5 px-4 bg-blue-500 text-white font-bold rounded-full hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        disabled={isUploading}
        on:click={handleSave}
        aria-label="Save profile changes"
      >
        {#if isUploading}
          <span class="flex items-center gap-2">
            <svg class="animate-spin h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Saving
          </span>
        {:else}
          Save
        {/if}
      </button>
    </div>
    
    <!-- Banner image -->
    <div class="relative h-48">
      <div class="absolute inset-0 overflow-hidden">
        {#if bannerPreview}
          <img 
            src={bannerPreview} 
            alt="Banner preview" 
            class="w-full h-full object-cover"
          />
        {:else}
          <div class="w-full h-full bg-blue-500"></div>
        {/if}
      </div>
      
      <div class="absolute inset-0 flex items-center justify-center bg-black/30">
        <label class="flex flex-col items-center justify-center cursor-pointer">
          <svg class="w-8 h-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          <span class="text-white font-medium mt-1">Add banner photo</span>
          <input 
            type="file" 
            accept="image/*" 
            class="hidden" 
            on:change={handleBannerChange}
          />
        </label>
      </div>
    </div>
    
    <!-- Profile picture -->
    <div class="px-4 -mt-16 mb-4 relative z-10">
      <div class="relative inline-block">
        <div class="border-4 border-white dark:border-gray-900 rounded-full overflow-hidden">
          {#if profilePicturePreview}
            <img 
              src={profilePicturePreview} 
              alt="Profile preview" 
              class="w-32 h-32 object-cover"
            />
          {:else}
            <div class="w-32 h-32 flex items-center justify-center bg-blue-200 dark:bg-blue-700 text-4xl font-bold">
              {formData.displayName.charAt(0).toUpperCase()}
            </div>
          {/if}
        </div>
        
        <label class="absolute inset-0 flex items-center justify-center rounded-full bg-black/30 cursor-pointer">
          <svg class="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          <input 
            type="file" 
            accept="image/*" 
            class="hidden" 
            on:change={handleProfilePictureChange}
          />
        </label>
      </div>
    </div>
    
    <!-- Form fields -->
    <form class="p-4 space-y-4" on:submit|preventDefault={handleSave}>
      <!-- Display name -->
      <div>
        <label for="displayName" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Display name
        </label>
        <input 
          type="text" 
          id="displayName"
          bind:value={formData.displayName}
          maxlength="50"
          class="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
          placeholder="Your display name"
        />
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400 text-right">
          {formData.displayName.length}/50
        </p>
      </div>
      
      <!-- Bio -->
      <div>
        <label for="bio" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Bio
        </label>
        <textarea 
          id="bio"
          bind:value={formData.bio}
          maxlength="160"
          rows="3"
          class="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-gray-800 text-gray-900 dark:text-white resize-none"
          placeholder="Tell us about yourself"
        ></textarea>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400 text-right">
          {formData.bio.length}/160
        </p>
      </div>
      
      <!-- Email -->
      <div>
        <label for="email" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Email
        </label>
        <input 
          type="email" 
          id="email"
          bind:value={formData.email}
          class="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
          placeholder="Your email address"
        />
      </div>
      
      <!-- Date of birth -->
      <div>
        <label for="dateOfBirth" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Date of birth
        </label>
        <input 
          type="date" 
          id="dateOfBirth"
          bind:value={formData.dateOfBirth}
          class="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
        />
      </div>
      
      <!-- Gender -->
      <div>
        <label for="gender" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Gender
        </label>
        <select 
          id="gender"
          bind:value={formData.gender}
          class="w-full px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
        >
          <option value="">Prefer not to say</option>
          <option value="male">Male</option>
          <option value="female">Female</option>
          <option value="other">Other</option>
        </select>
      </div>
    </form>
  </div>
</div>

<style>
  /* Only native CSS for backgrounds as requested */
  :global(:root) {
    --bg-color: #ffffff;
    --bg-secondary: #f7f9fa;
  }

  :global([data-theme="dark"]) {
    --bg-color: #000000;
    --bg-secondary: #16181c;
  }
</style> 