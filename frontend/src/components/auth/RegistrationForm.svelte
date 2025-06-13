<script lang="ts">
  import { useTheme } from '../../hooks/useTheme';
  import GoogleSignInButton from './GoogleSignInButton.svelte';
  import type { IDateOfBirth, ICustomWindow } from '../../interfaces/IAuth';

  const { theme } = useTheme();
  
  $: isDarkMode = $theme === 'dark';

  export let name = "";
  export let username = "";
  export let email = "";
  export let password = "";
  export let confirmPassword = "";
  export let gender = "";
  export let dateOfBirth: IDateOfBirth = { month: "", day: "", year: "" };
  export let profilePicture: File | string | null = null;
  export let banner: File | string | null = null;
  export let securityQuestion = "";
  export let securityAnswer = "";
  export let subscribeToNewsletter = false;
  // Form options
  export let months: string[] = [];
  export let days: string[] = [];
  export let years: string[] = [];
  export let securityQuestions: string[] = [];

  export let onSubmit: (token: string | null) => void; 
  export let onGoogleAuthSuccess: (result: any) => void;
  export let onGoogleAuthError: (error: string) => void;

  // Check if we're in development mode
  const isDevelopment = import.meta.env.DEV;

  async function triggerSubmit() {
    // In development mode, just submit with a development placeholder token
    if (isDevelopment) {
      onSubmit('dev-mode-token');
      return;
    }
    
    // In production, use the reCAPTCHA API if available
    if (typeof window !== 'undefined') {
      const customWindow = window as ICustomWindow;
      if (customWindow.grecaptcha) {
        try {
          // Using empty string as site key since it's likely already configured when the reCAPTCHA script loaded
          const token = await customWindow.grecaptcha.execute('', { action: 'register' });
          onSubmit(token);
        } catch (error) {
          console.error('reCAPTCHA error:', error);
          onSubmit(null);
        }
      } else {
        console.warn('reCAPTCHA not loaded');
        onSubmit(null);
      }
    }
  }

  // Type-safe event handlers for file inputs
  function handleProfilePictureChange(e: Event) {
    const target = e.target as HTMLInputElement;
    profilePicture = target.files ? target.files[0] : null;
  }

  function handleBannerChange(e: Event) {
    const target = e.target as HTMLInputElement;
    banner = target.files ? target.files[0] : null;
  }
</script>

<div class="auth-social-btn-container">
  <GoogleSignInButton
    onAuthSuccess={onGoogleAuthSuccess}
    onAuthError={onGoogleAuthError}
    containerId="google-signin-button"
    class="auth-social-btn {isDarkMode ? 'auth-social-btn-dark' : ''}"
  />
</div>

<div class="auth-divider {isDarkMode ? 'auth-divider-dark' : ''}">
  <span class="auth-divider-text">or create account with email</span>
</div>

<!-- Personal Information Section -->
<div class="auth-section">
  <h3 class="auth-section-title">Personal Information</h3>
    <!-- Name Input -->
  <div class="auth-input-group">
    <div class="flex justify-between">
      <label for="name" class="auth-label">Full Name</label>
      <span class="text-xs text-gray-500 dark:text-gray-400" data-cy="name-char-count">{name.length} / 50</span>
    </div>
    <input 
      type="text" 
      id="name" 
      bind:value={name} 
      maxlength="50"
      class="auth-input {isDarkMode ? 'auth-input-dark' : ''}"
      placeholder="Enter your full name"
      data-cy="name-input"
    />
  </div>

  <!-- Username input -->
  <div class="auth-input-group">
    <label for="username" class="auth-label">Username</label>
    <input 
      type="text" 
      id="username" 
      bind:value={username} 
      class="auth-input {isDarkMode ? 'auth-input-dark' : ''}"
      placeholder="Choose a unique username"
      data-cy="username-input"
    />
  </div>

  <!-- Gender selection -->  <div class="auth-input-group">
    <fieldset>
      <legend class="auth-label">Gender</legend>
      <div class="flex space-x-6">        <label class="auth-checkbox-group">
          <input 
            type="radio" 
            name="gender" 
            value="male" 
            bind:group={gender} 
            class="auth-checkbox"
            data-cy="gender-male"
          />
          <span class="auth-checkbox-label">Male</span>
        </label>
        <label class="auth-checkbox-group">
          <input 
            type="radio" 
            name="gender" 
            value="female" 
            bind:group={gender} 
            class="auth-checkbox"
            data-cy="gender-female"
          />
          <span class="auth-checkbox-label">Female</span>
        </label>
      </div>
    </fieldset>
  </div>

  <!-- Date of birth -->
  <div class="auth-input-group">
    <fieldset>
      <legend class="auth-label">Date of Birth</legend>
      <p class="auth-helper-text mb-3">This will not be shown publicly. Confirm your own age, even if this account is for a business, a pet, or something else.</p>
      
      <div class="flex space-x-3" role="group" aria-labelledby="dob-label">
        <div class="flex-1">
          <label for="dob-month" class="sr-only">Month</label>          <select 
            id="dob-month"
            bind:value={dateOfBirth.month} 
            class="auth-input {isDarkMode ? 'auth-input-dark' : ''}"
            data-cy="dob-month"
          >
            <option value="">Month</option>
            {#each months as month}
              <option value={month}>{month}</option>
            {/each}
          </select>
        </div>
        <div class="flex-1">
          <label for="dob-day" class="sr-only">Day</label>
          <select            id="dob-day"
            bind:value={dateOfBirth.day}
            class="auth-input {isDarkMode ? 'auth-input-dark' : ''}"
            data-cy="dob-day"
          >
            <option value="">Day</option>
            {#each days as day}
              <option value={day}>{day}</option>
            {/each}
          </select>
        </div>
        <div class="flex-1">
          <label for="dob-year" class="sr-only">Year</label>
          <select 
            id="dob-year"            bind:value={dateOfBirth.year}
            class="auth-input {isDarkMode ? 'auth-input-dark' : ''}"
            data-cy="dob-year"
          >
            <option value="">Year</option>
            {#each years as year}
              <option value={year}>{year}</option>
            {/each}
          </select>        </div>
      </div>
    </fieldset>
  </div>
</div>

<!-- Account Security Section -->
<div class="auth-section">
  <h3 class="auth-section-title">Account Security</h3>
  
  <!-- Email input -->
  <div class="auth-input-group">
    <label for="email" class="auth-label">Email Address</label>
    <input 
      type="email" 
      id="email"      bind:value={email} 
      class="auth-input {isDarkMode ? 'auth-input-dark' : ''}"
      placeholder="Enter your email address"
      data-cy="email-input"
    />
  </div>
  <!-- Password input -->
  <div class="auth-input-group">
    <label for="password" class="auth-label">Password</label>
    <input 
      type="password" 
      id="password" 
      bind:value={password} 
      class="auth-input {isDarkMode ? 'auth-input-dark' : ''}"
      placeholder="Create a secure password"
      data-cy="password-input"
    />
  </div>
  <!-- Confirm Password input -->
  <div class="auth-input-group">
    <label for="confirmPassword" class="auth-label">Confirm Password</label>
    <input 
      type="password"
      id="confirmPassword" 
      bind:value={confirmPassword} 
      class="auth-input {isDarkMode ? 'auth-input-dark' : ''}"
      placeholder="Confirm your password"
      data-cy="confirm-password-input"
    />
  </div>
  <!-- Security Question -->
  <div class="auth-input-group">
    <label for="securityQuestion" class="auth-label">Security Question</label>
    <select
      id="securityQuestion"
      bind:value={securityQuestion}
      class="auth-input {isDarkMode ? 'auth-input-dark' : ''}"
      data-cy="security-question"
    >
      <option value="">Select a security question</option>
      {#each securityQuestions as question}
        <option value={question}>{question}</option>
      {/each}
    </select>
  </div>
  <!-- Security Answer -->
  <div class="auth-input-group">
    <label for="securityAnswer" class="auth-label">Security Answer</label>
    <input 
      type="text"
      id="securityAnswer"
      bind:value={securityAnswer}
      class="auth-input {isDarkMode ? 'auth-input-dark' : ''}"
      placeholder="Enter your answer"
      data-cy="security-answer"
    />
  </div>
</div>

<!-- Optional Information Section -->
<div class="auth-section">
  <h3 class="auth-section-title">Profile Customization (Optional)</h3>
    <!-- Profile Picture -->
  <div class="auth-input-group">
    <label for="profilePicture" class="auth-label">Profile Picture</label>
    <input 
      type="file"
      id="profilePicture"
      accept="image/*"
      on:change={handleProfilePictureChange}
      class="auth-input {isDarkMode ? 'auth-input-dark' : ''}"
      data-cy="profile-picture"
    />
  </div>

  <!-- Banner Picture -->
  <div class="auth-input-group">
    <label for="banner" class="auth-label">Banner Image</label>
    <input 
      type="file"
      id="banner"
      accept="image/*"
      on:change={handleBannerChange}      class="auth-input {isDarkMode ? 'auth-input-dark' : ''}"
      data-cy="banner"
    />
  </div>

  <!-- Subscribe to newsletter -->
  <div class="auth-checkbox-group">
    <input 
      type="checkbox"
      id="subscribeToNewsletter"
      bind:checked={subscribeToNewsletter}
      class="auth-checkbox"
      data-cy="newsletter"
      aria-describedby="newsletter-description"
    />
    <label for="subscribeToNewsletter" class="auth-checkbox-label">
      Subscribe to our newsletter to receive news and updates about our platform
    </label>
    <span id="newsletter-description" class="sr-only">Receive news and updates about our platform via email</span>
  </div>
</div>

<button 
  type="button"
  on:click={triggerSubmit}
  class="auth-btn"
  data-cy="register-button"
  aria-label="Complete registration"
>
  Create My Account
</button>

<p class="text-xs mt-6 text-gray-400 text-center leading-relaxed">
  By signing up, you agree to our <a href="/terms" class="text-blue-500 hover:underline font-medium">Terms of Service</a> and 
  <a href="/privacy" class="text-blue-500 hover:underline font-medium">Privacy Policy</a>, including <a href="/cookies" class="text-blue-500 hover:underline font-medium">Cookie Use</a>.
</p>

