<script lang="ts">
  import MainLayout from '../components/layout/MainLayout.svelte';
  import { useTheme } from '../hooks/useTheme';
  import { getProfile, updateProfile } from '../api/user';
  import { onMount } from 'svelte';
  import { toastStore } from '../stores/toastStore';
  import ThemeToggle from '../components/common/ThemeToggle.svelte';
  
  const { theme } = useTheme();
  $: isDarkMode = $theme === 'dark';
  
  let isLoading = true;
  let isSaving = false;
  
  // User settings form data
  let userData = {
    name: '',
    username: '',
    email: '',
    bio: '',
    location: '',
    website: '',
    allow_mentions: true,
    notification_emails: true,
    marketing_emails: false,
    display_sensitive_content: false
  };
  
  // Load user data on mount
  onMount(async () => {
    isLoading = true;
    try {
      const response = await getProfile();
      const profile = response.user || (response.data && response.data.user);
      
      if (profile) {
        userData = {
          name: profile.name || profile.display_name || '',
          username: profile.username || '',
          email: profile.email || '',
          bio: profile.bio || '',
          location: profile.location || '',
          website: profile.website || '',
          allow_mentions: profile.allow_mentions !== false,
          notification_emails: profile.notification_emails !== false,
          marketing_emails: profile.marketing_emails === true,
          display_sensitive_content: profile.display_sensitive_content === true
        };
      }
    } catch (error) {
      console.error('Error loading profile:', error);
      toastStore.showToast('Failed to load profile settings', 'error');
    } finally {
      isLoading = false;
    }
  });
  
  async function handleSaveSettings() {
    isSaving = true;
    try {
      await updateProfile(userData);
      toastStore.showToast('Settings saved successfully', 'success');
    } catch (error) {
      console.error('Error saving settings:', error);
      toastStore.showToast('Failed to save settings', 'error');
    } finally {
      isSaving = false;
    }
  }
</script>

<MainLayout>
  <div class="page-header {isDarkMode ? 'page-header-dark' : ''}">
    <h1 class="page-title">Settings</h1>
  </div>
  
  <div class="settings-container {isDarkMode ? 'settings-container-dark' : ''}">
    {#if isLoading}
      <div class="settings-loading">
        <div class="settings-loading-spinner"></div>
        <p>Loading settings...</p>
      </div>
    {:else}
      <div class="settings-sections">
        <!-- Profile Settings Section -->
        <div class="settings-section {isDarkMode ? 'settings-section-dark' : ''}">
          <h2 class="settings-section-title">Profile Settings</h2>
          
          <div class="settings-form-group">
            <label for="name">Display Name</label>
            <input 
              type="text" 
              id="name" 
              class="settings-input {isDarkMode ? 'settings-input-dark' : ''}"
              bind:value={userData.name}
              placeholder="Your display name"
            />
          </div>
          
          <div class="settings-form-group">
            <label for="username">Username</label>
            <input 
              type="text" 
              id="username" 
              class="settings-input {isDarkMode ? 'settings-input-dark' : ''}"
              bind:value={userData.username}
              placeholder="username"
              disabled
            />
            <small>Username cannot be changed</small>
          </div>
          
          <div class="settings-form-group">
            <label for="bio">Bio</label>
            <textarea 
              id="bio" 
              class="settings-textarea {isDarkMode ? 'settings-textarea-dark' : ''}"
              bind:value={userData.bio}
              placeholder="Tell us about yourself"
              rows="3"
            ></textarea>
          </div>
          
          <div class="settings-form-group">
            <label for="location">Location</label>
            <input 
              type="text" 
              id="location" 
              class="settings-input {isDarkMode ? 'settings-input-dark' : ''}"
              bind:value={userData.location}
              placeholder="Your location"
            />
          </div>
          
          <div class="settings-form-group">
            <label for="website">Website</label>
            <input 
              type="url" 
              id="website" 
              class="settings-input {isDarkMode ? 'settings-input-dark' : ''}"
              bind:value={userData.website}
              placeholder="https://yourdomain.com"
            />
          </div>
        </div>
        
        <!-- Privacy Settings Section -->
        <div class="settings-section {isDarkMode ? 'settings-section-dark' : ''}">
          <h2 class="settings-section-title">Privacy Settings</h2>
          
          <div class="settings-toggle-group">
            <div class="settings-toggle-label">
              <span>Allow mentions</span>
              <small>Let people mention you in their posts</small>
            </div>
            <label class="toggle">
              <input type="checkbox" bind:checked={userData.allow_mentions}>
              <span class="toggle-slider"></span>
            </label>
          </div>
          
          <div class="settings-toggle-group">
            <div class="settings-toggle-label">
              <span>Display sensitive content</span>
              <small>Show content that may be sensitive or inappropriate</small>
            </div>
            <label class="toggle">
              <input type="checkbox" bind:checked={userData.display_sensitive_content}>
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
        
        <!-- Notification Settings Section -->
        <div class="settings-section {isDarkMode ? 'settings-section-dark' : ''}">
          <h2 class="settings-section-title">Notification Settings</h2>
          
          <div class="settings-toggle-group">
            <div class="settings-toggle-label">
              <span>Email notifications</span>
              <small>Receive notifications via email</small>
            </div>
            <label class="toggle">
              <input type="checkbox" bind:checked={userData.notification_emails}>
              <span class="toggle-slider"></span>
            </label>
          </div>
          
          <div class="settings-toggle-group">
            <div class="settings-toggle-label">
              <span>Marketing emails</span>
              <small>Receive promotional content and updates</small>
            </div>
            <label class="toggle">
              <input type="checkbox" bind:checked={userData.marketing_emails}>
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
        
        <!-- Theme Settings Section -->
        <div class="settings-section {isDarkMode ? 'settings-section-dark' : ''}">
          <h2 class="settings-section-title">Display Settings</h2>
          
          <div class="settings-toggle-group theme-toggle-container">
            <div class="settings-toggle-label">
              <span>Dark mode</span>
              <small>Switch between light and dark theme</small>
            </div>
            <ThemeToggle size="md" />
          </div>
        </div>
        
        <div class="settings-actions">
          <button 
            class="settings-save-btn"
            on:click={handleSaveSettings}
            disabled={isSaving}
          >
            {isSaving ? 'Saving...' : 'Save Settings'}
          </button>
        </div>
      </div>
    {/if}
  </div>
</MainLayout>

<style>
  .settings-container {
    padding: var(--space-4);
  }
  
  .settings-loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--space-8);
  }
  
  .settings-loading-spinner {
    width: 40px;
    height: 40px;
    border: 3px solid var(--border-color);
    border-top: 3px solid var(--color-primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: var(--space-4);
  }
  
  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
  
  .settings-sections {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }
  
  .settings-section {
    background-color: var(--bg-secondary);
    border-radius: var(--radius-lg);
    padding: var(--space-4);
  }
  
  .settings-section-dark {
    background-color: var(--dark-bg-secondary);
  }
  
  .settings-section-title {
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-bold);
    margin-bottom: var(--space-4);
    padding-bottom: var(--space-2);
    border-bottom: 1px solid var(--border-color);
  }
  
  .settings-form-group {
    margin-bottom: var(--space-4);
  }
  
  .settings-form-group label {
    display: block;
    margin-bottom: var(--space-2);
    font-weight: var(--font-weight-medium);
  }
  
  .settings-form-group small {
    display: block;
    font-size: var(--font-size-xs);
    color: var(--text-tertiary);
    margin-top: var(--space-1);
  }
  
  .settings-input,
  .settings-textarea {
    width: 100%;
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    background-color: var(--bg-primary);
    color: var(--text-primary);
    font-size: var(--font-size-base);
  }
  
  .settings-input-dark,
  .settings-textarea-dark {
    border-color: var(--border-color-dark);
    background-color: var(--bg-primary-dark);
    color: var(--text-primary-dark);
  }
  
  .settings-input:focus,
  .settings-textarea:focus {
    border-color: var(--color-primary);
    outline: none;
  }
  
  .settings-textarea {
    resize: vertical;
    min-height: 80px;
  }
  
  .settings-toggle-group {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--space-4);
  }
  
  .settings-toggle-label {
    display: flex;
    flex-direction: column;
  }
  
  .settings-toggle-label small {
    font-size: var(--font-size-xs);
    color: var(--text-tertiary);
    margin-top: var(--space-1);
  }
  
  .toggle {
    position: relative;
    display: inline-block;
    width: 52px;
    height: 26px;
  }
  
  .toggle input {
    opacity: 0;
    width: 0;
    height: 0;
  }
  
  .toggle-slider {
    position: absolute;
    cursor: pointer;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: var(--border-color);
    transition: .4s;
    border-radius: 34px;
  }
  
  .toggle-slider:before {
    position: absolute;
    content: "";
    height: 20px;
    width: 20px;
    left: 3px;
    bottom: 3px;
    background-color: white;
    transition: .4s;
    border-radius: 50%;
  }
  
  input:checked + .toggle-slider {
    background-color: var(--color-primary);
  }
  
  input:checked + .toggle-slider:before {
    transform: translateX(26px);
  }
  
  .theme-toggle-container {
    margin-bottom: 0;
  }
  
  .settings-actions {
    margin-top: var(--space-4);
    display: flex;
    justify-content: flex-end;
  }
  
  .settings-save-btn {
    padding: var(--space-2) var(--space-5);
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: var(--radius-full);
    font-weight: var(--font-weight-bold);
    cursor: pointer;
    transition: background-color 0.2s;
  }
  
  .settings-save-btn:hover {
    background-color: var(--color-primary-hover);
  }
  
  .settings-save-btn:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }
  
  @media (max-width: 768px) {
    .settings-toggle-group {
      flex-direction: column;
      align-items: flex-start;
      gap: var(--space-2);
    }
  }
</style>
