<script lang="ts">
  import { useTheme } from '../../hooks/useTheme';
  import GoogleSignInButton from './GoogleSignInButton.svelte';
  import type { IDateOfBirth } from '../../interfaces/IAuth';

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

  // Validation errors
  export let nameError = "";
  export let usernameError = "";
  export let emailError = "";
  export let passwordErrors: string[] = [];
  export let confirmPasswordError = "";
  export let genderError = "";
  export let dateOfBirthError = "";
  export let securityQuestionError = "";
  export let profilePictureError = "";
  export let bannerError = "";

  export let onNameBlur: () => void;
  export let onUsernameBlur: () => void;
  export let onEmailBlur: () => void;
  export let onPasswordBlur: () => void;
  export let onConfirmPasswordBlur: () => void;
  export let onGenderChange: () => void;
  export let onDateOfBirthChange: () => void;
  export let onSecurityQuestionChange: () => void;
  export let onSecurityAnswerBlur: () => void;

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
    if (typeof window !== 'undefined' && window.grecaptcha) {
      try {
        const token = await window.grecaptcha.execute();
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
  <span class="auth-divider-text">or</span>
</div>

<!-- Name Input -->
<div class="auth-input-group">
  <div class="flex justify-between">
    <label for="name" class="auth-label">Name</label>
    <span class="text-xs text-gray-500 dark:text-gray-400" data-cy="name-char-count">{name.length} / 50</span>
  </div>
  <input 
    type="text" 
    id="name" 
    bind:value={name} 
    on:blur={onNameBlur}
    maxlength="50"
    class="auth-input {isDarkMode ? 'auth-input-dark' : ''} {nameError ? 'auth-input-error' : ''}"
    placeholder="Name"
    data-cy="name-input"
  />
  {#if nameError}
    <p class="auth-error-message" data-cy="name-error">{nameError}</p>
  {/if}
</div>

<!-- Username input -->
<div class="auth-input-group">
  <label for="username" class="auth-label">Username</label>
  <input 
    type="text" 
    id="username" 
    bind:value={username} 
    on:blur={onUsernameBlur}
    class="auth-input {isDarkMode ? 'auth-input-dark' : ''} {usernameError ? 'auth-input-error' : ''}"
    placeholder="Username"
    data-cy="username-input"
  />
  {#if usernameError}
    <p class="auth-error-message" data-cy="username-error">{usernameError}</p>
  {/if}
</div>

<!-- Email input -->
<div class="auth-input-group">
  <label for="email" class="auth-label">Email</label>
  <input 
    type="email" 
    id="email" 
    bind:value={email} 
    on:blur={onEmailBlur}
    class="auth-input {isDarkMode ? 'auth-input-dark' : ''} {emailError ? 'auth-input-error' : ''}"
    placeholder="Email"
    data-cy="email-input"
  />
  {#if emailError}
    <p class="auth-error-message" data-cy="email-error">{emailError}</p>
  {/if}
</div>

<!-- Password input -->
<div class="auth-input-group">
  <label for="password" class="auth-label">Password</label>
  <input 
    type="password" 
    id="password" 
    bind:value={password} 
    on:blur={onPasswordBlur}
    class="auth-input {isDarkMode ? 'auth-input-dark' : ''} {passwordErrors.length > 0 ? 'auth-input-error' : ''}"
    placeholder="Password"
    data-cy="password-input"
  />
  {#if passwordErrors.length > 0}
    <div class="auth-error-message" data-cy="password-error">
      {#each passwordErrors as error}
        <p>{error}</p>
      {/each}
    </div>
  {/if}
</div>

<!-- Confirm Password input -->
<div class="auth-input-group">
  <label for="confirmPassword" class="auth-label">Confirm Password</label>
  <input 
    type="password" 
    id="confirmPassword" 
    bind:value={confirmPassword} 
    on:blur={onConfirmPasswordBlur}
    class="auth-input {isDarkMode ? 'auth-input-dark' : ''} {confirmPasswordError ? 'auth-input-error' : ''}"
    placeholder="Confirm Password"
    data-cy="confirm-password-input"
  />
  {#if confirmPasswordError}
    <p class="auth-error-message" data-cy="password-match-error">{confirmPasswordError}</p>
  {/if}
</div>

<!-- Gender selection -->
<div class="auth-input-group">
  <label class="auth-label">Gender</label>
  <div class="flex space-x-4">
    <label class="auth-checkbox-group">
      <input 
        type="radio" 
        name="gender" 
        value="male" 
        bind:group={gender} 
        on:change={onGenderChange}
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
        on:change={onGenderChange}
        class="auth-checkbox"
        data-cy="gender-female"
      />
      <span class="auth-checkbox-label">Female</span>
    </label>
  </div>
  {#if genderError}
    <p class="auth-error-message" data-cy="gender-error">{genderError}</p>
  {/if}
</div>

<!-- Date of birth -->
<div class="auth-input-group">
  <label class="auth-label">Date of birth</label>
  <p class="auth-helper-text mb-2">This will not be shown publicly. Confirm your own age, even if this account is for a business, a pet, or something else.</p>
  
  <div class="flex space-x-2">
    <div class="w-1/3">
      <select 
        bind:value={dateOfBirth.month} 
        on:change={onDateOfBirthChange}
        class="auth-input {isDarkMode ? 'auth-input-dark' : ''} {dateOfBirthError ? 'auth-input-error' : ''}"
        data-cy="dob-month"
      >
        <option value="">Month</option>
        {#each months as month}
          <option value={month}>{month}</option>
        {/each}
      </select>
    </div>
    <div class="w-1/3">
      <select 
        bind:value={dateOfBirth.day}
        on:change={onDateOfBirthChange}
        class="auth-input {isDarkMode ? 'auth-input-dark' : ''} {dateOfBirthError ? 'auth-input-error' : ''}"
        data-cy="dob-day"
      >
        <option value="">Day</option>
        {#each days as day}
          <option value={day}>{day}</option>
        {/each}
      </select>
    </div>
    <div class="w-1/3">
      <select 
        bind:value={dateOfBirth.year}
        on:change={onDateOfBirthChange}
        class="auth-input {isDarkMode ? 'auth-input-dark' : ''} {dateOfBirthError ? 'auth-input-error' : ''}"
        data-cy="dob-year"
      >
        <option value="">Year</option>
        {#each years as year}
          <option value={year}>{year}</option>
        {/each}
      </select>
    </div>
  </div>
  {#if dateOfBirthError}
    <p class="auth-error-message" data-cy="dob-error">{dateOfBirthError}</p>
  {/if}
</div>

<!-- Security Question -->
<div class="auth-input-group">
  <label for="securityQuestion" class="auth-label">Security Question</label>
  <select 
    id="securityQuestion"
    bind:value={securityQuestion}
    on:change={onSecurityQuestionChange}
    class="auth-input {isDarkMode ? 'auth-input-dark' : ''} {securityQuestionError ? 'auth-input-error' : ''}"
    data-cy="security-question"
  >
    <option value="">Select a security question</option>
    {#each securityQuestions as question}
      <option value={question}>{question}</option>
    {/each}
  </select>
  {#if securityQuestionError}
    <p class="auth-error-message" data-cy="security-question-error">{securityQuestionError}</p>
  {/if}
</div>

<!-- Security Answer -->
<div class="auth-input-group">
  <label for="securityAnswer" class="auth-label">Security Answer</label>
  <input 
    type="text"
    id="securityAnswer"
    bind:value={securityAnswer}
    on:blur={onSecurityAnswerBlur}
    class="auth-input {isDarkMode ? 'auth-input-dark' : ''}"
    placeholder="Your answer"
    data-cy="security-answer"
  />
</div>

<!-- Profile Picture -->
<div class="auth-input-group">
  <label for="profilePicture" class="auth-label">Profile Picture (optional)</label>
  <input 
    type="file"
    id="profilePicture"
    accept="image/*"
    on:change={(e) => profilePicture = e.target.files ? e.target.files[0] : null}
    class="auth-input {isDarkMode ? 'auth-input-dark' : ''}"
    data-cy="profile-picture"
  />
  {#if profilePictureError}
    <p class="auth-error-message" data-cy="profile-picture-error">{profilePictureError}</p>
  {/if}
</div>

<!-- Banner Picture -->
<div class="auth-input-group">
  <label for="banner" class="auth-label">Banner Image (optional)</label>
  <input 
    type="file"
    id="banner"
    accept="image/*"
    on:change={(e) => banner = e.target.files ? e.target.files[0] : null}
    class="auth-input {isDarkMode ? 'auth-input-dark' : ''}"
    data-cy="banner"
  />
  {#if bannerError}
    <p class="auth-error-message" data-cy="banner-error">{bannerError}</p>
  {/if}
</div>

<!-- Subscribe to newsletter -->
<div class="auth-checkbox-group">
  <input 
    type="checkbox"
    id="subscribeToNewsletter"
    bind:checked={subscribeToNewsletter}
    class="auth-checkbox"
    data-cy="newsletter"
  />
  <label for="subscribeToNewsletter" class="auth-checkbox-label">
    Subscribe to our newsletter
  </label>
</div>

<button 
  type="button"
  on:click={triggerSubmit}
  class="auth-btn"
  data-cy="register-button"
>
  Sign up
</button>

<p class="text-xs mt-4 text-gray-400 text-center">
  By signing up, you agree to the <a href="/terms" class="text-blue-500 hover:underline">Terms of Service</a> and 
  <a href="/privacy" class="text-blue-500 hover:underline">Privacy Policy</a>, including <a href="/cookies" class="text-blue-500 hover:underline">Cookie Use</a>.
</p>