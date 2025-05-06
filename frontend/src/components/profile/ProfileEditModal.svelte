<!-- ProfileEditModal.svelte -->
<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { uploadProfilePicture, uploadBanner } from '../../api/user';
  import { toastStore } from '../../stores/toastStore';
  import type { IUserProfile } from '../../interfaces/IUser';
  
  export let profile: IUserProfile | null = null;
  export let isOpen = false;
  
  const dispatch = createEventDispatcher();
  
  let profilePictureFile: File | null = null;
  let bannerFile: File | null = null;
  let profilePicturePreview: string | null = null;
  let bannerPreview: string | null = null;
  let isUploading = false;
  let errorMessage = '';
  
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
</script>

<div class={`modal ${isOpen ? 'active' : ''}`}>
  <div class="modal-backdrop" on:click={handleClose}></div>
  <div class="modal-content">
    <div class="modal-header">
      <h2>Edit Profile Media</h2>
      <button class="close-btn" on:click={handleClose}>&times;</button>
    </div>
    
    <div class="modal-body">
      {#if errorMessage}
        <div class="error-message">{errorMessage}</div>
      {/if}
      
      <div class="media-section">
        <h3>Profile Picture</h3>
        <div class="profile-picture-preview">
          {#if profilePicturePreview}
            <img src={profilePicturePreview} alt="Profile Preview" />
            <button class="remove-btn" on:click={removeProfilePicture}>Remove</button>
          {:else}
            <div class="placeholder">No profile picture selected</div>
          {/if}
        </div>
        <input 
          type="file" 
          id="profilePicture" 
          accept="image/*" 
          on:change={handleProfilePictureChange}
        />
      </div>
      
      <div class="media-section">
        <h3>Banner</h3>
        <div class="banner-preview">
          {#if bannerPreview}
            <img src={bannerPreview} alt="Banner Preview" />
            <button class="remove-btn" on:click={removeBanner}>Remove</button>
          {:else}
            <div class="placeholder">No banner selected</div>
          {/if}
        </div>
        <input 
          type="file" 
          id="banner" 
          accept="image/*" 
          on:change={handleBannerChange}
        />
      </div>
    </div>
    
    <div class="modal-footer">
      <button class="cancel-btn" on:click={handleClose} disabled={isUploading}>Cancel</button>
      <button class="save-btn" on:click={handleSave} disabled={isUploading}>
        {isUploading ? 'Saving...' : 'Save Changes'}
      </button>
    </div>
  </div>
</div>

<style>
  .modal {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    display: none;
    z-index: 1000;
  }
  
  .modal.active {
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .modal-backdrop {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
  }
  
  .modal-content {
    position: relative;
    background-color: #fff;
    border-radius: 8px;
    width: 90%;
    max-width: 500px;
    z-index: 1001;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    max-height: 90vh;
    overflow-y: auto;
  }
  
  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px;
    border-bottom: 1px solid #e1e1e1;
  }
  
  .modal-header h2 {
    margin: 0;
    font-size: 1.2rem;
  }
  
  .close-btn {
    background: none;
    border: none;
    font-size: 1.5rem;
    cursor: pointer;
  }
  
  .modal-body {
    padding: 16px;
  }
  
  .media-section {
    margin-bottom: 24px;
  }
  
  .media-section h3 {
    margin-top: 0;
    margin-bottom: 8px;
    font-size: 1rem;
  }
  
  .profile-picture-preview, .banner-preview {
    margin-bottom: 12px;
    position: relative;
  }
  
  .profile-picture-preview img {
    width: 100px;
    height: 100px;
    border-radius: 50%;
    object-fit: cover;
  }
  
  .banner-preview img {
    width: 100%;
    height: 120px;
    object-fit: cover;
    border-radius: 4px;
  }
  
  .placeholder {
    padding: 16px;
    background-color: #f5f5f5;
    border-radius: 4px;
    text-align: center;
    color: #666;
  }
  
  .remove-btn {
    position: absolute;
    top: 4px;
    right: 4px;
    background-color: rgba(255, 255, 255, 0.8);
    border: none;
    border-radius: 4px;
    padding: 4px 8px;
    cursor: pointer;
    font-size: 0.8rem;
  }
  
  .error-message {
    background-color: #fee;
    color: #d33;
    padding: 8px;
    border-radius: 4px;
    margin-bottom: 16px;
  }
  
  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    padding: 16px;
    border-top: 1px solid #e1e1e1;
  }
  
  button {
    padding: 8px 16px;
    border-radius: 4px;
    cursor: pointer;
  }
  
  button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
  
  .cancel-btn {
    background-color: #f5f5f5;
    border: 1px solid #ddd;
  }
  
  .save-btn {
    background-color: #1da1f2;
    color: white;
    border: none;
  }
</style> 