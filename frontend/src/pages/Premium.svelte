<script lang="ts">
  import { onMount } from 'svelte';
  import { submitPremiumRequest } from '../api/user';
  import { toastStore } from '../stores/toastStore';
  import { useTheme } from '../hooks/useTheme';
  import MainLayout from '../components/layout/MainLayout.svelte';
  import Button from '../components/common/Button.svelte';
  import { 
    AlertCircleIcon 
  } from 'svelte-feather-icons';
  import CheckIcon from 'svelte-feather-icons/src/icons/CheckIcon.svelte';
  import UserIcon from 'svelte-feather-icons/src/icons/UserIcon.svelte';
  import ShieldIcon from 'svelte-feather-icons/src/icons/ShieldIcon.svelte';
  import UploadIcon from 'svelte-feather-icons/src/icons/UploadIcon.svelte';
  
  const { theme } = useTheme();
  
  $: isDarkMode = $theme === 'dark';
  
  let showVerificationForm = false;
  let identityCardNumber = '';
  let reason = '';
  let facePhoto: File | null = null;
  let facePhotoURL = '';
  let formError = '';
  let isSubmitting = false;

  function handleFileChange(event: Event) {
    const target = event.target as HTMLInputElement;
    const files = target.files;
    
    if (files && files.length > 0) {
      facePhoto = files[0];
      facePhotoURL = URL.createObjectURL(facePhoto);
    }
  }

  // Function to resize image
  async function resizeImage(file: File, maxWidth = 800, maxHeight = 600, quality = 0.8): Promise<string> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.readAsDataURL(file);
      
      reader.onload = (e) => {
        const img = new Image();
        img.src = e.target?.result as string;
        
        img.onload = () => {
          // Calculate new dimensions while maintaining aspect ratio
          let width = img.width;
          let height = img.height;
          
          if (width > height) {
            if (width > maxWidth) {
              height = Math.round(height * maxWidth / width);
              width = maxWidth;
            }
          } else {
            if (height > maxHeight) {
              width = Math.round(width * maxHeight / height);
              height = maxHeight;
            }
          }
          
          // Create canvas and resize
          const canvas = document.createElement('canvas');
          canvas.width = width;
          canvas.height = height;
          
          const ctx = canvas.getContext('2d');
          if (!ctx) {
            reject(new Error('Failed to get canvas context'));
            return;
          }
          
          ctx.drawImage(img, 0, 0, width, height);
          
          // Convert to data URL
          const resizedDataURL = canvas.toDataURL('image/jpeg', quality);
          resolve(resizedDataURL);
        };
        
        img.onerror = () => {
          reject(new Error('Failed to load image'));
        };
      };
      
      reader.onerror = () => {
        reject(new Error('Failed to read file'));
      };
    });
  }
  
  async function handleSubmit() {
    formError = '';
    
    if (!reason) {
      formError = 'Please provide a reason for verification';
      return;
    }
    
    if (!identityCardNumber) {
      formError = 'Please enter your identity card number';
      return;
    }
    
    if (!facePhoto) {
      formError = 'Please upload a photo of your face for verification';
      return;
    }
    
    isSubmitting = true;
    
    try {
      // Log the file details
      console.log(`Processing face photo: ${facePhoto.name}, size: ${facePhoto.size}B, type: ${facePhoto.type}`);
      
      // Always resize the image to reduce its size (max 800x600px, quality 0.7)
      console.log('Resizing image to ensure it fits within size limits...');
      try {
        // Resize the image to reduce its size
        const resizedPhotoDataURL = await resizeImage(facePhoto, 600, 400, 0.7);
        console.log(`Resized photo to data URL of length: ${resizedPhotoDataURL.length}`);
        
        // Submit with the resized image
        const success = await submitPremiumRequest(
          reason,
          identityCardNumber,
          resizedPhotoDataURL
        );
        
        if (success) {
          toastStore.showToast('Your verification request has been submitted', 'success');
          // Reset form and hide it
          showVerificationForm = false;
          identityCardNumber = '';
          reason = '';
          facePhoto = null;
          if (facePhotoURL) {
            URL.revokeObjectURL(facePhotoURL);
            facePhotoURL = '';
          }
        } else {
          formError = 'Failed to submit verification request. Please try again.';
        }
      } catch (resizeError) {
        console.error('Error resizing image:', resizeError);
        formError = 'Error processing image. Please try with a smaller photo.';
      }
      
      isSubmitting = false;
      
    } catch (error) {
      console.error('Error uploading face photo:', error);
      formError = 'Error processing image. Please try again.';
      isSubmitting = false;
    }
  }
</script>

<MainLayout>
  <div class="page-header {isDarkMode ? 'page-header-dark' : ''}">
    <h1 class="page-title">Get Verified</h1>
  </div>
  
  <div class="premium-container {isDarkMode ? 'premium-container-dark' : ''}">
    <div class="premium-header">
      <div class="main-benefit">
        <div class="checkmark-badge">
          <CheckIcon size="32" />
        </div>
        <h2>Get the blue checkmark</h2>
        <p>Stand out in the community with a verified profile</p>
      </div>
      
      <div class="premium-info">
        <p class="highlight-text">The blue checkmark badge increases your credibility and visibility across AYCOM</p>
      </div>
    </div>
    
    {#if showVerificationForm}
      <div class="verification-form {isDarkMode ? 'verification-form-dark' : ''}">
        <h3>Verification Request</h3>
        <p>Please complete this form to request account verification</p>
        
        <div class="verification-steps">
          <div class="step">
            <div class="step-icon user-check-icon">
              <UserIcon size="20" />
            </div>
            <div class="step-text">1. Submit your information</div>
          </div>
          <div class="step-connector"></div>
          <div class="step">
            <div class="step-icon"><ShieldIcon size="20" /></div>
            <div class="step-text">2. Admin review</div>
          </div>
          <div class="step-connector"></div>
          <div class="step">
            <div class="step-icon"><CheckIcon size="20" /></div>
            <div class="step-text">3. Get verified</div>
          </div>
        </div>
        
        {#if formError}
          <div class="form-error">
            <AlertCircleIcon size="18" />
            <span>{formError}</span>
          </div>
        {/if}
        
        <form on:submit|preventDefault={handleSubmit}>
          <div class="form-group">
            <label for="identity-card">National Identity Card Number</label>
            <input 
              type="text" 
              id="identity-card" 
              bind:value={identityCardNumber} 
              placeholder="Enter your ID number"
              required
              disabled={isSubmitting}
            />
            <small>Your ID number is securely encrypted using SHA-256 before storage</small>
          </div>
          
          <div class="form-group">
            <label for="reason">Reason for Verification</label>
            <textarea 
              id="reason" 
              bind:value={reason} 
              placeholder="Why do you want to be verified?"
              required
              disabled={isSubmitting}
            ></textarea>
          </div>
          
          <div class="form-group">
            <label for="face-photo">Photo for Verification</label>
            <div class="photo-upload">
              <input 
                type="file" 
                id="face-photo" 
                accept="image/*"
                on:change={handleFileChange}
                disabled={isSubmitting}
              />
              <label for="face-photo" class="upload-button">
                <UploadIcon size="16" />
                {facePhoto ? 'Change Photo' : 'Upload Photo'}
              </label>
            </div>
            
            {#if facePhotoURL}
              <div class="photo-preview">
                <img src={facePhotoURL} alt="Face verification preview" />
              </div>
            {/if}
          </div>
          
          <div class="security-notice">
            <ShieldIcon size="16" />
            <p>Your personal information is protected with enterprise-grade security and only used for verification purposes</p>
          </div>
          
          <div class="form-actions">
            <button 
              type="button" 
              class="cancel-btn"
              on:click={() => { showVerificationForm = false; }}
              disabled={isSubmitting}
            >
              Cancel
            </button>
            <button 
              type="submit" 
              class="submit-btn"
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Submitting...' : 'Submit for Verification'}
            </button>
          </div>
        </form>
      </div>
    {:else}
      <div class="benefits-container">
        <div class="verification-card">
          <div class="verification-card-content">
            <h3>Get Verified Today</h3>
            <div class="verify-badge-large">
              <CheckIcon size="32" />
            </div>
            <ul class="benefits-list">
              <li>
                <CheckIcon size="16" />
                <span>Blue checkmark verification badge</span>
              </li>
              <li>
                <CheckIcon size="16" />
                <span>Increased visibility and credibility</span>
              </li>
              <li>
                <CheckIcon size="16" />
                <span>Stand out in comments and posts</span>
              </li>
            </ul>
            
            <div class="verification-process">
              <h4>How it works:</h4>
              <ol>
                <li>Submit your national identity card number</li>
                <li>Provide a reason for verification</li>
                <li>Upload a photo of your face</li>
                <li>Admin team reviews your application</li>
              </ol>
            </div>
            
            <button 
              class="verify-btn"
              on:click={() => showVerificationForm = true}
            >
              Apply for Verification
            </button>
          </div>
        </div>
      </div>
      
      <div class="security-container">
        <div class="security-header">
          <ShieldIcon size="20" />
          <h3>Your Security is Our Priority</h3>
        </div>
        <ul class="security-features">
          <li>Your identity card number is encrypted using SHA-256</li>
          <li>All verification information is stored securely</li>
          <li>Your data is only used for verification purposes</li>
          <li>Our admin team follows strict privacy protocols</li>
        </ul>
      </div>
    {/if}
  </div>
</MainLayout>

<style>
  .premium-container {
    padding: var(--space-4);
    max-width: 900px;
    margin: 0 auto;
  }
  
  .premium-header {
    text-align: center;
    margin-bottom: var(--space-6);
  }
  
  .main-benefit {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-6);
  }
  
  .checkmark-badge {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 64px;
    height: 64px;
    background-color: #1DA1F2;
    color: white;
    border-radius: var(--radius-full);
    margin-bottom: var(--space-2);
  }
  
  .premium-header h2 {
    font-size: var(--font-size-2xl);
    margin-bottom: var(--space-2);
    font-weight: var(--font-weight-bold);
  }
  
  .premium-info {
    margin-top: var(--space-4);
  }
  
  .highlight-text {
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-medium);
    color: var(--text-primary);
    background-color: var(--bg-secondary);
    padding: var(--space-3);
    border-radius: var(--radius-md);
    max-width: 600px;
    margin: 0 auto;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }
  
  .benefits-container {
    display: flex;
    justify-content: center;
    margin-bottom: var(--space-6);
  }
  
  .verification-card {
    background-color: var(--bg-secondary);
    border-radius: var(--radius-lg);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    overflow: hidden;
    width: 100%;
    max-width: 500px;
  }
  
  .verification-card-content {
    padding: var(--space-5);
    text-align: center;
  }
  
  .verification-card h3 {
    font-size: var(--font-size-xl);
    margin-bottom: var(--space-4);
  }
  
  .verify-badge-large {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 80px;
    height: 80px;
    background-color: #1DA1F2;
    color: white;
    border-radius: var(--radius-full);
    margin: 0 auto var(--space-5);
  }
  
  .benefits-list {
    list-style: none;
    padding: 0;
    margin: 0 0 var(--space-5) 0;
    text-align: left;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  
  .benefits-list li {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--font-size-md);
  }
  
  .benefits-list li :global(svg) {
    color: var(--color-success);
    flex-shrink: 0;
  }
  
  .verification-process {
    text-align: left;
    margin-bottom: var(--space-5);
    padding: var(--space-3);
    background-color: var(--bg-primary);
    border-radius: var(--radius-md);
  }
  
  .verification-process h4 {
    margin-bottom: var(--space-2);
  }
  
  .verification-process ol {
    padding-left: var(--space-4);
    margin: 0;
  }
  
  .verification-process li {
    margin-bottom: var(--space-2);
  }
  
  .verify-btn {
    width: 100%;
    padding: var(--space-3);
    background-color: #1DA1F2;
    color: white;
    border: none;
    border-radius: var(--radius-full);
    font-weight: var(--font-weight-bold);
    font-size: var(--font-size-lg);
    cursor: pointer;
    transition: background-color 0.2s;
  }
  
  .verify-btn:hover {
    background-color: #1A91DA;
  }
  
  .verification-form {
    max-width: 600px;
    margin: 0 auto var(--space-6);
    background-color: var(--bg-secondary);
    border-radius: var(--radius-lg);
    padding: var(--space-5);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }
  
  .verification-form-dark {
    background-color: var(--dark-bg-secondary);
  }
  
  .verification-form h3 {
    font-size: var(--font-size-xl);
    margin-bottom: var(--space-2);
    text-align: center;
  }
  
  .verification-form p {
    text-align: center;
    margin-bottom: var(--space-4);
    color: var(--text-secondary);
  }
  
  .verification-steps {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-5);
    padding: var(--space-3);
    background-color: var(--bg-primary);
    border-radius: var(--radius-md);
  }
  
  .step {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-1);
    text-align: center;
  }
  
  .step-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 36px;
    border-radius: var(--radius-full);
    background-color: var(--color-primary);
    color: white;
  }
  
  .step-text {
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
  }
  
  .step-connector {
    flex-grow: 1;
    height: 2px;
    background-color: var(--border-color);
    margin: 0 var(--space-2);
  }
  
  .form-error {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    background-color: var(--color-error-bg);
    color: var(--color-error);
    padding: var(--space-2);
    border-radius: var(--radius-md);
    margin-bottom: var(--space-4);
  }
  
  .form-error :global(svg) {
    flex-shrink: 0;
  }
  
  .form-group {
    margin-bottom: var(--space-4);
  }
  
  .form-group label {
    display: block;
    margin-bottom: var(--space-1);
    font-weight: var(--font-weight-medium);
  }
  
  .form-group small {
    display: block;
    color: var(--text-tertiary);
    margin-top: var(--space-1);
    font-size: var(--font-size-xs);
  }
  
  .form-group input[type="text"],
  .form-group textarea {
    width: 100%;
    padding: var(--space-2);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    background-color: var(--bg-primary);
    color: var(--text-primary);
  }
  
  .form-group textarea {
    min-height: 100px;
    resize: vertical;
  }
  
  .photo-upload {
    display: flex;
    align-items: center;
  }
  
  .photo-upload input[type="file"] {
    display: none;
  }
  
  .upload-button {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    padding: var(--space-2) var(--space-3);
    background-color: var(--color-primary);
    color: white;
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: background-color 0.2s;
  }
  
  .upload-button:hover {
    background-color: var(--color-primary-hover);
  }
  
  .photo-preview {
    margin-top: var(--space-3);
    max-width: 200px;
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    overflow: hidden;
  }
  
  .photo-preview img {
    width: 100%;
    height: auto;
    display: block;
  }
  
  .security-notice {
    display: flex;
    align-items: flex-start;
    gap: var(--space-2);
    padding: var(--space-3);
    background-color: var(--bg-primary);
    border-radius: var(--radius-md);
    margin-bottom: var(--space-4);
  }
  
  .security-notice :global(svg) {
    color: var(--color-primary);
    margin-top: 3px;
    flex-shrink: 0;
  }
  
  .security-notice p {
    margin: 0;
    font-size: var(--font-size-sm);
    text-align: left;
    color: var(--text-primary);
  }
  
  .form-actions {
    display: flex;
    justify-content: space-between;
    gap: var(--space-3);
    margin-top: var(--space-5);
  }
  
  .cancel-btn {
    flex: 1;
    padding: var(--space-3);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    background-color: transparent;
    color: var(--text-primary);
    cursor: pointer;
    transition: background-color 0.2s;
  }
  
  .cancel-btn:hover {
    background-color: var(--bg-hover);
  }
  
  .submit-btn {
    flex: 2;
    padding: var(--space-3);
    background-color: var(--color-primary);
    color: white;
    border: none;
    border-radius: var(--radius-md);
    font-weight: var(--font-weight-medium);
    cursor: pointer;
    transition: background-color 0.2s;
  }
  
  .submit-btn:hover:not(:disabled) {
    background-color: var(--color-primary-hover);
  }
  
  .submit-btn:disabled,
  .cancel-btn:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }
  
  .security-container {
    margin-top: var(--space-6);
    background-color: var(--bg-secondary);
    border-radius: var(--radius-lg);
    padding: var(--space-4);
  }
  
  .security-header {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-3);
  }
  
  .security-header :global(svg) {
    color: var(--color-success);
  }
  
  .security-header h3 {
    margin: 0;
    font-size: var(--font-size-lg);
  }
  
  .security-features {
    margin: 0;
    padding-left: var(--space-4);
  }
  
  .security-features li {
    margin-bottom: var(--space-2);
  }
  
  @media (max-width: 768px) {
    .verification-steps {
      flex-direction: column;
      gap: var(--space-3);
    }
    
    .step-connector {
      width: 2px;
      height: 20px;
      margin: 0;
    }
    
    .security-container {
      padding: var(--space-3);
    }
    
    .form-actions {
      flex-direction: column;
    }
    
    .submit-btn, .cancel-btn {
      width: 100%;
    }
  }
  
  .premium-container-dark .security-container {
    background-color: var(--bg-dark-secondary);
    border: 1px solid var(--border-dark);
  }
  
  .premium-container-dark .security-header,
  .premium-container-dark .security-features li {
    color: var(--text-dark-primary);
  }
  
  /* Custom user-check-icon style to replace UserCheckIcon */
  .user-check-icon {
    position: relative;
  }
  
  .user-check-icon::after {
    content: '✓';
    position: absolute;
    bottom: -2px;
    right: -2px;
    background-color: #1DA1F2;
    color: white;
    border-radius: 50%;
    width: 12px;
    height: 12px;
    font-size: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
  }
</style> 