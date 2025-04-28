<script lang="ts">
  import { onMount } from 'svelte';
  import { useProfile } from '../hooks/useProfile';
  import type { IUserUpdateRequest, IUser } from '../interfaces/IUser';
  
  // Get profile management functions from the hook
  const { 
    profile, 
    loading, 
    error, 
    fetchProfile, 
    updateProfile,
    followUser,
    unfollowUser,
    getFollowers,
    getFollowing
  } = useProfile();
  
  let editMode = false;
  let followers: IUser[] = [];
  let following: IUser[] = [];
  let showFollowers = false;
  let showFollowing = false;
  
  // Form data for profile updates
  let formData: IUserUpdateRequest = {
    name: '',
    username: '',
    bio: '',
    location: '',
    website: ''
  };
  
  let userId = 'user-123'; // This would typically come from auth context or route params
  
  // Fetch profile data on component mount
  onMount(async () => {
    await fetchProfile(userId);
    
    // Initialize form data with current profile
    if ($profile) {
      formData = {
        name: $profile.name || '',
        username: $profile.username || '',
        bio: $profile.bio || '',
        location: $profile.location || '',
        website: $profile.website || ''
      };
    }
  });
  
  // Handle profile update
  async function handleUpdateProfile() {
    const result = await updateProfile(formData);
    if (result.success) {
      editMode = false;
    }
  }
  
  // Toggle edit mode
  function toggleEditMode() {
    editMode = !editMode;
    
    // Reset form data when entering edit mode
    if (editMode && $profile) {
      formData = {
        name: $profile.name || '',
        username: $profile.username || '',
        bio: $profile.bio || '',
        location: $profile.location || '',
        website: $profile.website || ''
      };
    }
  }
  
  // Load followers
  async function loadFollowers() {
    showFollowers = true;
    showFollowing = false;
    followers = await getFollowers(userId);
  }
  
  // Load following
  async function loadFollowing() {
    showFollowing = true;
    showFollowers = false;
    following = await getFollowing(userId);
  }
  
  // Handle follow/unfollow
  async function handleFollowToggle(id: string, isFollowing: boolean) {
    if (isFollowing) {
      await unfollowUser(id);
    } else {
      await followUser(id);
    }
  }
</script>

<svelte:head>
  <title>{$profile ? $profile.name : 'Profile'} | AYCOM</title>
</svelte:head>

<div class="profile-container">
  {#if $loading}
    <div class="loading-spinner">Loading profile...</div>
  {:else if $error}
    <div class="error-message">{$error}</div>
  {:else if $profile}
    <!-- Profile Banner -->
    <div class="profile-banner" style="background-image: url({$profile.banner || 'https://via.placeholder.com/1500x300'})">
      {#if editMode}
        <button class="edit-banner-btn">Update Banner</button>
      {/if}
    </div>
    
    <!-- Profile Header -->
    <div class="profile-header">
      <div class="profile-picture-container">
        <img src={$profile.profile_picture || 'https://via.placeholder.com/150'} alt="{$profile.name}'s profile" class="profile-picture" />
        {#if editMode}
          <button class="edit-picture-btn">Update Picture</button>
        {/if}
      </div>
      
      <div class="profile-actions">
        {#if editMode}
          <button class="primary-btn" on:click={handleUpdateProfile}>Save</button>
          <button class="secondary-btn" on:click={toggleEditMode}>Cancel</button>
        {:else}
          <button class="primary-btn" on:click={toggleEditMode}>Edit Profile</button>
        {/if}
      </div>
    </div>
    
    <!-- Profile Info -->
    <div class="profile-info">
      {#if editMode}
        <div class="edit-form">
          <div class="form-group">
            <label for="name">Name</label>
            <input id="name" type="text" bind:value={formData.name} maxlength="50" />
          </div>
          
          <div class="form-group">
            <label for="username">Username</label>
            <input id="username" type="text" bind:value={formData.username} maxlength="15" />
          </div>
          
          <div class="form-group">
            <label for="bio">Bio</label>
            <textarea id="bio" bind:value={formData.bio} maxlength="160"></textarea>
          </div>
          
          <div class="form-group">
            <label for="location">Location</label>
            <input id="location" type="text" bind:value={formData.location} maxlength="30" />
          </div>
          
          <div class="form-group">
            <label for="website">Website</label>
            <input id="website" type="url" bind:value={formData.website} placeholder="https://" />
          </div>
        </div>
      {:else}
        <div class="profile-details">
          <h1 class="profile-name">{$profile.name} {#if $profile.verified}<span class="verified-badge">✓</span>{/if}</h1>
          <p class="profile-username">@{$profile.username}</p>
          
          {#if $profile.bio}
            <p class="profile-bio">{$profile.bio}</p>
          {/if}
          
          <div class="profile-metadata">
            {#if $profile.location}
              <span class="location"><i class="icon-location"></i> {$profile.location}</span>
            {/if}
            
            {#if $profile.website}
              <span class="website"><i class="icon-link"></i> <a href={$profile.website} target="_blank" rel="noopener">{$profile.website.replace(/^https?:\/\//, '')}</a></span>
            {/if}
            
            <span class="join-date"><i class="icon-calendar"></i> Joined {new Date($profile.joined_date).toLocaleDateString('en-US', { month: 'long', year: 'numeric' })}</span>
          </div>
          
          <div class="profile-stats">
            <button class="stat-item" on:click={loadFollowing}>
              <span class="stat-value">{$profile.following_count}</span>
              <span class="stat-label">Following</span>
            </button>
            <button class="stat-item" on:click={loadFollowers}>
              <span class="stat-value">{$profile.followers_count}</span>
              <span class="stat-label">Followers</span>
            </button>
            <div class="stat-item">
              <span class="stat-value">{$profile.tweets_count}</span>
              <span class="stat-label">Posts</span>
            </div>
          </div>
        </div>
      {/if}
    </div>
    
    <!-- Followers/Following Modal -->
    {#if showFollowers && followers.length > 0}
      <div class="modal">
        <div class="modal-content">
          <div class="modal-header">
            <h2>Followers</h2>
            <button class="close-btn" on:click={() => showFollowers = false}>×</button>
          </div>
          <div class="modal-body">
            <ul class="user-list">
              {#each followers as user}
                <li class="user-item">
                  <img src={user.profile_picture} alt="{user.name}'s profile" class="user-avatar" />
                  <div class="user-info">
                    <div class="user-name">{user.name} {#if user.verified}<span class="verified-badge">✓</span>{/if}</div>
                    <div class="user-username">@{user.username}</div>
                  </div>
                  <button class="follow-btn" on:click={() => handleFollowToggle(user.id, user.isFollowing)}>
                    {user.isFollowing ? 'Unfollow' : 'Follow'}
                  </button>
                </li>
              {/each}
            </ul>
          </div>
        </div>
      </div>
    {/if}
    
    {#if showFollowing && following.length > 0}
      <div class="modal">
        <div class="modal-content">
          <div class="modal-header">
            <h2>Following</h2>
            <button class="close-btn" on:click={() => showFollowing = false}>×</button>
          </div>
          <div class="modal-body">
            <ul class="user-list">
              {#each following as user}
                <li class="user-item">
                  <img src={user.profile_picture} alt="{user.name}'s profile" class="user-avatar" />
                  <div class="user-info">
                    <div class="user-name">{user.name} {#if user.verified}<span class="verified-badge">✓</span>{/if}</div>
                    <div class="user-username">@{user.username}</div>
                  </div>
                  <button class="follow-btn" on:click={() => handleFollowToggle(user.id, user.isFollowing)}>
                    {user.isFollowing ? 'Unfollow' : 'Follow'}
                  </button>
                </li>
              {/each}
            </ul>
          </div>
        </div>
      </div>
    {/if}
  {:else}
    <div class="not-found">Profile not found</div>
  {/if}
</div>

<style>
  .profile-container {
    max-width: 800px;
    margin: 0 auto;
    padding-bottom: 2rem;
  }
  
  .loading-spinner, .error-message, .not-found {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 300px;
    font-size: 1.2rem;
  }
  
  .error-message {
    color: #e53935;
  }
  
  .profile-banner {
    height: 200px;
    background-size: cover;
    background-position: center;
    position: relative;
    border-radius: 15px 15px 0 0;
  }
  
  .edit-banner-btn {
    position: absolute;
    bottom: 10px;
    right: 10px;
    background: rgba(0, 0, 0, 0.6);
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 20px;
    cursor: pointer;
  }
  
  .profile-header {
    display: flex;
    justify-content: space-between;
    padding: 0 1rem;
    margin-top: -40px;
  }
  
  .profile-picture-container {
    position: relative;
  }
  
  .profile-picture {
    width: 150px;
    height: 150px;
    border-radius: 50%;
    border: 4px solid white;
    background-color: white;
    object-fit: cover;
  }
  
  .edit-picture-btn {
    position: absolute;
    bottom: 10px;
    right: 10px;
    background: rgba(0, 0, 0, 0.6);
    color: white;
    border: none;
    padding: 0.3rem 0.5rem;
    border-radius: 20px;
    font-size: 0.8rem;
    cursor: pointer;
  }
  
  .profile-actions {
    margin-top: 1rem;
  }
  
  .primary-btn, .secondary-btn {
    padding: 0.5rem 1.5rem;
    border-radius: 20px;
    font-weight: bold;
    cursor: pointer;
  }
  
  .primary-btn {
    background-color: #1da1f2;
    color: white;
    border: none;
  }
  
  .secondary-btn {
    background-color: transparent;
    color: #1da1f2;
    border: 1px solid #1da1f2;
    margin-left: 0.5rem;
  }
  
  .profile-info {
    padding: 1rem;
    margin-top: 1rem;
  }
  
  .profile-details {
    margin-top: 1rem;
  }
  
  .profile-name {
    font-size: 1.5rem;
    font-weight: bold;
    margin: 0;
    display: flex;
    align-items: center;
  }
  
  .verified-badge {
    display: inline-block;
    margin-left: 5px;
    color: white;
    background-color: #1da1f2;
    border-radius: 50%;
    font-size: 0.8rem;
    width: 18px;
    height: 18px;
    text-align: center;
    line-height: 18px;
  }
  
  .profile-username {
    color: #657786;
    margin: 0;
  }
  
  .profile-bio {
    margin: 1rem 0;
  }
  
  .profile-metadata {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    color: #657786;
    margin-bottom: 1rem;
  }
  
  .profile-metadata a {
    color: #1da1f2;
    text-decoration: none;
  }
  
  .profile-stats {
    display: flex;
    gap: 1.5rem;
    margin-top: 1rem;
  }
  
  .stat-item {
    display: flex;
    gap: 0.3rem;
    align-items: center;
    background: none;
    border: none;
    padding: 0;
    font-size: inherit;
    cursor: pointer;
  }
  
  .stat-value {
    font-weight: bold;
  }
  
  .stat-label {
    color: #657786;
  }
  
  .edit-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  
  .form-group label {
    font-weight: bold;
    font-size: 0.9rem;
  }
  
  .form-group input, .form-group textarea {
    padding: 0.8rem;
    border: 1px solid #ccd6dd;
    border-radius: 4px;
    font-size: 1rem;
  }
  
  .form-group textarea {
    resize: vertical;
    min-height: 100px;
  }
  
  .modal {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 100;
  }
  
  .modal-content {
    background-color: white;
    border-radius: 15px;
    width: 90%;
    max-width: 600px;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
  }
  
  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    border-bottom: 1px solid #eee;
  }
  
  .modal-header h2 {
    margin: 0;
  }
  
  .close-btn {
    background: none;
    border: none;
    font-size: 1.5rem;
    cursor: pointer;
  }
  
  .modal-body {
    padding: 1rem;
    overflow-y: auto;
  }
  
  .user-list {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  
  .user-item {
    display: flex;
    align-items: center;
    padding: 1rem 0;
    border-bottom: 1px solid #eee;
  }
  
  .user-avatar {
    width: 50px;
    height: 50px;
    border-radius: 50%;
    margin-right: 1rem;
  }
  
  .user-info {
    flex: 1;
  }
  
  .user-name {
    font-weight: bold;
    display: flex;
    align-items: center;
  }
  
  .user-username {
    color: #657786;
    font-size: 0.9rem;
  }
  
  .follow-btn {
    background-color: #1da1f2;
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 20px;
    font-weight: bold;
    cursor: pointer;
  }
  
  .follow-btn:hover {
    background-color: #1a91da;
  }
</style> 