<!-- ProfileEditModal.svelte -->
<script lang="ts">
  import { createEventDispatcher, onMount } from "svelte";
  import { uploadFile } from "../../utils/supabase";
  import { updateProfile } from "../../api/user";
  import { toastStore } from "../../stores/toastStore";
  import { formatStorageUrl } from "../../utils/common";
  import XIcon from "svelte-feather-icons/src/icons/XIcon.svelte";
  import CameraIcon from "svelte-feather-icons/src/icons/CameraIcon.svelte";
  import type { IUserProfile } from "../../interfaces/IUser";
  import { useTheme } from "../../hooks/useTheme";

  const dispatch = createEventDispatcher();
  const { theme } = useTheme();

  $: isDarkMode = $theme === "dark";

  export let profile: IUserProfile | null = null;
  export let isOpen: boolean = false;

  let formData = {
    name: "",
    bio: "",
    email: "",
    date_of_birth: "",
    gender: ""
  };

  let profilePictureFile: File | null = null;
  let bannerFile: File | null = null;
  let profilePicturePreview: string | null = null;
  let bannerPreview: string | null = null;
  let isUploading = false;
  let errorMessage = "";

  $: if (profile) {
    formData = {
      name: profile.name || "",
      bio: profile.bio || "",
      email: profile.email || "",
      date_of_birth: profile.date_of_birth || "",
      gender: profile.gender || ""
    };

    profilePicturePreview = profile.profile_picture_url || "";
    bannerPreview = profile.banner_url || "";

    console.log("[ProfileEditModal] Initializing form data from profile:", formData);
    console.log("[ProfileEditModal] Profile picture URL:", profilePicturePreview);
    console.log("[ProfileEditModal] Banner URL:", bannerPreview);
  }

  onMount(() => {
    if (profile) {
      formData = {
        name: profile.name || "",
        bio: profile.bio || "",
        email: profile.email || "",
        date_of_birth: profile.date_of_birth || "",
        gender: profile.gender || ""
      };

      profilePicturePreview = profile.profile_picture_url || "";
      bannerPreview = profile.banner_url || "";

      console.log("[ProfileEditModal] Profile data on mount:", {
        name: profile.name,
        bio: profile.bio,
        email: profile.email,
        date_of_birth: profile.date_of_birth,
        gender: profile.gender,
        profile_picture_url: profilePicturePreview,
        banner_url: bannerPreview,
      });
    }
  });

  function handleProfilePictureChange(e: Event) {
    const input = e.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      profilePictureFile = input.files[0];

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
    if (profile) {
      if (formData.name !== profile.name ||
          formData.bio !== profile.bio ||
          formData.email !== profile.email ||
          formData.date_of_birth !== profile.date_of_birth ||
          formData.gender !== profile.gender) {
        dispatch("updateProfile", formData);
      }
    }

    if (!profilePictureFile && !bannerFile) {
      dispatch("close");
      return;
    }

    isUploading = true;
    errorMessage = "";

    try {
      if (profilePictureFile) {
        console.log("Uploading profile picture:", profilePictureFile.name);
        const profilePictureUrl = await uploadFile(profilePictureFile);
        if (profilePictureUrl) {
          toastStore.showToast("Profile picture updated successfully", "success");
          console.log("Profile picture updated successfully:", profilePictureUrl);
          dispatch("profilePictureUpdated", { url: profilePictureUrl });
        } else {
          throw new Error("Failed to upload profile picture");
        }
      }

      if (bannerFile) {
        console.log("Uploading banner:", bannerFile.name);
        const bannerUrl = await uploadFile(bannerFile);
        if (bannerUrl) {
          toastStore.showToast("Banner updated successfully", "success");
          console.log("Banner updated successfully:", bannerUrl);
          dispatch("bannerUpdated", { url: bannerUrl });
        } else {
          throw new Error("Failed to upload banner");
        }
      }

      dispatch("close");
    } catch (err) {
      console.error("Error updating profile media:", err);
      errorMessage = err instanceof Error ? err.message : "Failed to update profile media";
      toastStore.showToast(errorMessage, "error");
    } finally {
      isUploading = false;
    }
  }

  function handleClose() {
    dispatch("close");
  }

  function handleModalClick(e: MouseEvent) {

    if ((e.currentTarget as HTMLElement) === e.target) {
      handleClose();
    }
  }

  function handleModalKeydown(e: KeyboardEvent) {
    if (e.key === "Escape") {
      handleClose();
    }
  }

  function handleInput(event: Event) {
    const target = event.target as HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement;
    const field = target.id;
    const value = target.value;

    formData = {
      ...formData,
      [field]: value
    };
  }
</script>

{#if isOpen}
<div
  class="modal-overlay"
  on:click={handleModalClick}
  on:keydown={handleModalKeydown}
  role="dialog"
  aria-modal="true"
  aria-labelledby="modal-title"
  tabindex="0"
>
  <div
    class="modal-container"
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
              {formData.name.charAt(0).toUpperCase()}
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
        <label for="name" class="form-label">
          Display name
        </label>
        <input
          type="text"
          id="name"
          value={formData.name}
          on:input={handleInput}
          maxlength="50"
          class="form-input"
          placeholder="Your display name"
        />
        <p class="form-help-text">
          {formData.name.length}/50
        </p>
      </div>

      <!-- Bio -->
      <div class="form-field">
        <label for="bio" class="form-label">
          Bio
        </label>
        <textarea
          id="bio"
          value={formData.bio}
          on:input={handleInput}
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
          value={formData.email}
          on:input={handleInput}
          class="form-input"
          placeholder="Your email address"
        />
      </div>

      <!-- Date of birth -->
      <div class="form-field">
        <label for="date_of_birth" class="form-label">
          Date of birth
        </label>
        <input
          type="date"
          id="date_of_birth"
          value={formData.date_of_birth}
          on:input={handleInput}
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
          value={formData.gender}
          on:input={handleInput}
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

  .modal-overlay {
    position: fixed;
    inset: 0;
    background-color: var(--modal-overlay-bg);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 50;
  }

  .modal-container {
    background-color: var(--bg-color);
    border-radius: 16px;
    width: 100%;
    max-width: 600px;
    margin: 0 16px;
    max-height: 90vh;
    overflow-y: auto;
  }

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