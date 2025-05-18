<!-- ProfileEditModal.svelte -->
<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { uploadProfilePicture, uploadBanner } from '../../api/user';
  import { toastStore } from '../../stores/toastStore';
  import type { IUserProfile } from '../../interfaces/IUser';
  import { useTheme } from '../../hooks/useTheme';
  
  // Define default image URL for fallback
  const DEFAULT_AVATAR = "https://secure.gravatar.com/avatar/0?d=mp";
  
  const dispatch = createEventDispatcher();
  const { theme } = useTheme();
  
  // Reactive declarations
  $: isDarkMode = $theme === 'dark';
  
  export let profile: IUserProfile | null = null;
  export let isOpen: boolean = false;
  
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
      // Handle profile picture URL - directly use the URL without formatting
      profilePicturePreview = profile.profile_picture_url || profile.profile_picture || profile.avatar || DEFAULT_AVATAR;
      
      // Handle banner URL - directly use the URL without formatting
      bannerPreview = profile.banner_url || profile.banner || profile.background_banner_url || '';
      
      console.log('[ProfileEditModal] Profile data:', {
        profilePicture: profilePicturePreview,
        banner: bannerPreview,
      });
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
        console.log('Uploading profile picture:', profilePictureFile.name);
        const profileResult = await uploadProfilePicture(profilePictureFile);
        if (profileResult && profileResult.success) {
          toastStore.showToast('Profile picture updated successfully', 'success');
          console.log('Profile picture updated successfully:', profileResult.url);
          dispatch('profilePictureUpdated', { url: profileResult.url });
        } else {
          throw new Error('Failed to upload profile picture');
        }
      }
      
      // Upload banner if selected
      if (bannerFile) {
        console.log('Uploading banner:', bannerFile.name);
        const bannerResult = await uploadBanner(bannerFile);
        if (bannerResult && bannerResult.success) {
          toastStore.showToast('Banner updated successfully', 'success');
          console.log('Banner updated successfully:', bannerResult.url);
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

{#if isOpen}
<div 
  class="modal-overlay"
  on:click={handleClose}
  on:keydown={handleKeyDown}
  role="dialog" 
  aria-modal="true"
  aria-labelledby="modal-title"
  tabindex="0"
>
  <div 
    class="modal-container"
    on:click|stopPropagation
    on:keydown|stopPropagation
    role="document"
  >
    <!-- Header -->
    <div class="modal-header">
      <div class="header-left">
        <button 
          on:click={handleClose} 
          class="close-button"
          aria-label="Close modal"
        >
          <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
        <h2 id="modal-title" class="modal-title">Edit profile</h2>
      </div>
      
      <button 
        class="save-button"
        disabled={isUploading}
        on:click={handleSave}
        aria-label="Save profile changes"
      >
        {#if isUploading}
          <span class="loading-indicator">
            <svg class="spinner" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
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
    <div class="banner-container">
      <div class="banner-wrapper">
        {#if bannerPreview}
          <img 
            src={bannerPreview} 
            alt="Banner preview" 
            class="banner-image"
          />
        {:else}
          <div class="banner-placeholder"></div>
        {/if}
      </div>
      
      <div class="banner-overlay">
        <label class="upload-label">
          <svg class="upload-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          <span class="upload-text">Add banner photo</span>
          <input 
            type="file" 
            accept="image/*" 
            class="file-input" 
            on:change={handleBannerChange}
          />
        </label>
      </div>
    </div>
    
    <!-- Profile picture -->
    <div class="profile-picture-container">
      <div class="profile-picture-wrapper">
        <div class="profile-picture-border">
          {#if profilePicturePreview}
            <img 
              src={profilePicturePreview} 
              alt="Profile preview" 
              class="profile-picture"
            />
          {:else}
            <div class="profile-picture-placeholder">
              {formData.displayName.charAt(0).toUpperCase()}
            </div>
          {/if}
        </div>
        
        <label class="profile-picture-overlay">
          <svg class="upload-icon-small" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          <input 
            type="file" 
            accept="image/*" 
            class="file-input" 
            on:change={handleProfilePictureChange}
          />
        </label>
      </div>
    </div>
    
    <!-- Form fields -->
    <form class="form-container" on:submit|preventDefault={handleSave}>
      <!-- Display name -->
      <div class="form-field">
        <label for="displayName" class="form-label">
          Display name
        </label>
        <input 
          type="text" 
          id="displayName"
          bind:value={formData.displayName}
          maxlength="50"
          class="form-input"
          placeholder="Your display name"
        />
        <p class="form-help-text">
          {formData.displayName.length}/50
        </p>
      </div>
      
      <!-- Bio -->
      <div class="form-field">
        <label for="bio" class="form-label">
          Bio
        </label>
        <textarea 
          id="bio"
          bind:value={formData.bio}
          maxlength="160"
          rows="3"
          class="form-textarea"
          placeholder="Tell us about yourself"
        ></textarea>
        <p class="form-help-text">
          {formData.bio.length}/160
        </p>
      </div>
      
      <!-- Email -->
      <div class="form-field">
        <label for="email" class="form-label">
          Email
        </label>
        <input 
          type="email" 
          id="email"
          bind:value={formData.email}
          class="form-input"
          placeholder="Your email address"
        />
      </div>
      
      <!-- Date of birth -->
      <div class="form-field">
        <label for="dateOfBirth" class="form-label">
          Date of birth
        </label>
        <input 
          type="date" 
          id="dateOfBirth"
          bind:value={formData.dateOfBirth}
          class="form-input"
        />
      </div>
      
      <!-- Gender -->
      <div class="form-field">
        <label for="gender" class="form-label">
          Gender
        </label>
        <select 
          id="gender"
          bind:value={formData.gender}
          class="form-select"
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
{/if}

<style>
  /* Base theme variables */
  :global(:root) {
    --bg-color: #ffffff;
    --bg-secondary: #f7f9fa;
    --text-primary: #0f1419;
    --text-secondary: #536471;
    --border-color: #eff3f4;
    --color-primary: #1da1f2;
    --color-primary-hover: #1a91da;
    --modal-overlay-bg: rgba(0, 0, 0, 0.6);
  }

  :global([data-theme="dark"]) {
    --bg-color: #000000;
    --bg-secondary: #16181c;
    --text-primary: #e7e9ea;
    --text-secondary: #71767b;
    --border-color: #2f3336;
  }

  /* Modal overlay */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background-color: var(--modal-overlay-bg);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 50;
  }

  /* Modal container */
  .modal-container {
    background-color: var(--bg-color);
    border-radius: 16px;
    width: 100%;
    max-width: 600px;
    margin: 0 16px;
    max-height: 90vh;
    overflow-y: auto;
  }

  /* Modal header */
  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-color);
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .close-button {
    background: none;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    display: flex;
    padding: 8px;
    border-radius: 50%;
    transition: background-color 0.2s;
  }

  .close-button:hover {
    background-color: rgba(0, 0, 0, 0.05);
  }

  .modal-title {
    font-size: 20px;
    font-weight: 700;
    color: var(--text-primary);
    margin: 0;
  }

  .save-button {
    padding: 6px 16px;
    background-color: var(--color-primary);
    color: white;
    font-weight: 700;
    border-radius: 9999px;
    border: none;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .save-button:hover:not(:disabled) {
    background-color: var(--color-primary-hover);
  }

  .save-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .loading-indicator {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .spinner {
    animation: spin 1s linear infinite;
    height: 16px;
    width: 16px;
  }

  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  /* Banner styling */
  .banner-container {
    position: relative;
    width: 100%;
    height: 200px;
  }

  .banner-wrapper {
    position: absolute;
    inset: 0;
    overflow: hidden;
  }

  .banner-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .banner-placeholder {
    width: 100%;
    height: 100%;
    background-color: var(--color-primary);
  }

  .banner-overlay {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: rgba(0, 0, 0, 0.3);
  }

  /* Profile picture styling */
  .profile-picture-container {
    padding: 0 16px;
    margin-top: -64px;
    margin-bottom: 16px;
    position: relative;
    z-index: 1;
  }

  .profile-picture-wrapper {
    position: relative;
    display: inline-block;
  }

  .profile-picture-border {
    width: 112px;
    height: 112px;
    border-radius: 50%;
    border: 4px solid var(--bg-color);
    overflow: hidden;
  }

  .profile-picture {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .profile-picture-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: #333;
    color: white;
    font-size: 48px;
    font-weight: bold;
  }

  .profile-picture-overlay {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: rgba(0, 0, 0, 0.3);
    cursor: pointer;
  }

  /* Upload elements */
  .upload-label {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    cursor: pointer;
  }

  .upload-icon {
    width: 32px;
    height: 32px;
    color: white;
  }

  .upload-icon-small {
    width: 24px;
    height: 24px;
    color: white;
  }

  .upload-text {
    color: white;
    font-weight: 500;
    margin-top: 4px;
  }

  .file-input {
    display: none;
  }

  /* Form styling */
  .form-container {
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .form-field {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .form-label {
    font-size: 14px;
    font-weight: 500;
    color: var(--text-primary);
  }

  .form-input,
  .form-textarea,
  .form-select {
    padding: 10px 12px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background-color: var(--bg-color);
    color: var(--text-primary);
    font-size: 16px;
    width: 100%;
    transition: border-color 0.2s;
  }

  .form-input:focus,
  .form-textarea:focus,
  .form-select:focus {
    outline: none;
    border-color: var(--color-primary);
    box-shadow: 0 0 0 2px rgba(29, 161, 242, 0.2);
  }

  .form-textarea {
    resize: none;
  }

  .form-help-text {
    font-size: 12px;
    color: var(--text-secondary);
    text-align: right;
    margin: 2px 0 0 0;
  }
</style> 