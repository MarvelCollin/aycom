<script lang="ts">
  import { useTheme } from '../../hooks/useTheme';
  import GoogleSignInButton from './GoogleSignInButton.svelte';
  import ReCaptchaWrapper from './ReCaptchaWrapper.svelte';
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
  
  // reCAPTCHA variables
  let recaptchaToken: string | null = null;
  let recaptchaWrapper: ReCaptchaWrapper;
  
  function handleRecaptchaSuccess(event: CustomEvent<{ token: string }>) {
    recaptchaToken = event.detail.token;
  }

  function handleRecaptchaError() {
    recaptchaToken = null;
  }

  function handleRecaptchaExpired() {
    recaptchaToken = null;
  }

  // Check if we're in development mode
  const isDevelopment = import.meta.env.DEV;

  async function triggerSubmit() {
    // In development mode, just submit with a development placeholder token
    if (isDevelopment) {
      onSubmit('dev-mode-token');
      return;
    }
    
    // Use reCAPTCHA token if available
    onSubmit(recaptchaToken);
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
    aria-invalid={nameError ? "true" : "false"}
    aria-describedby={nameError ? "name-error" : undefined}
  />
  {#if nameError}
    <p id="name-error" class="auth-error-message" data-cy="name-error" role="alert">{nameError}</p>
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
    aria-invalid={usernameError ? "true" : "false"}
    aria-describedby={usernameError ? "username-error" : undefined}
  />
  {#if usernameError}
    <p id="username-error" class="auth-error-message" data-cy="username-error" role="alert">{usernameError}</p>
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
    aria-invalid={emailError ? "true" : "false"}
    aria-describedby={emailError ? "email-error" : undefined}
  />
  {#if emailError}
    <p id="email-error" class="auth-error-message" data-cy="email-error" role="alert">{emailError}</p>
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
    aria-invalid={passwordErrors.length > 0 ? "true" : "false"}
    aria-describedby={passwordErrors.length > 0 ? "password-error" : undefined}
  />
  {#if passwordErrors.length > 0}
    <div id="password-error" class="auth-error-message" data-cy="password-error" role="alert">
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
    aria-invalid={confirmPasswordError ? "true" : "false"}
    aria-describedby={confirmPasswordError ? "password-match-error" : undefined}
  />
  {#if confirmPasswordError}
    <p id="password-match-error" class="auth-error-message" data-cy="password-match-error" role="alert">{confirmPasswordError}</p>
  {/if}
</div>

<!-- Gender selection -->
<div class="auth-input-group">
  <fieldset>
    <legend class="auth-label">Gender</legend>
    <div class="auth-radio-group">
      <label class="auth-checkbox-group">
        <input 
          type="radio" 
          name="gender" 
          value="male" 
          bind:group={gender} 
          on:change={onGenderChange}
          class="auth-checkbox"
          data-cy="gender-male"
          aria-invalid={genderError ? "true" : "false"}
          aria-describedby={genderError ? "gender-error" : undefined}
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
          aria-invalid={genderError ? "true" : "false"}
          aria-describedby={genderError ? "gender-error" : undefined}
        />
        <span class="auth-checkbox-label">Female</span>
      </label>
    </div>
    {#if genderError}
      <p id="gender-error" class="auth-error-message" data-cy="gender-error" role="alert">{genderError}</p>
    {/if}
  </fieldset>
</div>

<!-- Date of birth -->
<div class="auth-input-group">
  <fieldset>
    <legend class="auth-label">Date of birth</legend>
    <p class="auth-helper-text mb-2">This will not be shown publicly. Confirm your own age, even if this account is for a business, a pet, or something else.</p>
    
    <div class="auth-dob-container">
      <div class="auth-dob-select-group" role="group" aria-labelledby="dob-label">
        <div>
          <label for="dob-month" class="sr-only">Month</label>
          <select 
            id="dob-month"
            bind:value={dateOfBirth.month} 
            on:change={onDateOfBirthChange}
            class="auth-input {isDarkMode ? 'auth-input-dark' : ''} {dateOfBirthError ? 'auth-input-error' : ''}"
            data-cy="dob-month"
            aria-invalid={dateOfBirthError ? "true" : "false"}
            aria-describedby={dateOfBirthError ? "dob-error" : undefined}
          >
            <option value="">Month</option>
            {#each months as month}
              <option value={month}>{month}</option>
            {/each}
          </select>
        </div>
        <div>
          <label for="dob-day" class="sr-only">Day</label>
          <select 
            id="dob-day"
            bind:value={dateOfBirth.day}
            on:change={onDateOfBirthChange}
            class="auth-input {isDarkMode ? 'auth-input-dark' : ''} {dateOfBirthError ? 'auth-input-error' : ''}"
            data-cy="dob-day"
            aria-invalid={dateOfBirthError ? "true" : "false"}
            aria-describedby={dateOfBirthError ? "dob-error" : undefined}
          >
            <option value="">Day</option>
            {#each days as day}
              <option value={day}>{day}</option>
            {/each}
          </select>
        </div>
        <div>
          <label for="dob-year" class="sr-only">Year</label>
          <select 
            id="dob-year"
            bind:value={dateOfBirth.year}
            on:change={onDateOfBirthChange}
            class="auth-input {isDarkMode ? 'auth-input-dark' : ''} {dateOfBirthError ? 'auth-input-error' : ''}"
            data-cy="dob-year"
            aria-invalid={dateOfBirthError ? "true" : "false"}
            aria-describedby={dateOfBirthError ? "dob-error" : undefined}
          >
            <option value="">Year</option>
            {#each years as year}
              <option value={year}>{year}</option>
            {/each}
          </select>
        </div>
      </div>
      {#if dateOfBirthError}
        <p id="dob-error" class="auth-error-message" data-cy="dob-error" role="alert">{dateOfBirthError}</p>
      {/if}
    </div>
  </fieldset>
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
    aria-invalid={securityQuestionError ? "true" : "false"}
    aria-describedby={securityQuestionError ? "security-question-error" : undefined}
  >
    <option value="">Select a security question</option>
    {#each securityQuestions as question}
      <option value={question}>{question}</option>
    {/each}
  </select>
  {#if securityQuestionError}
    <p id="security-question-error" class="auth-error-message" data-cy="security-question-error" role="alert">{securityQuestionError}</p>
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
    aria-required="true"
  />
</div>

<!-- Profile Picture -->
<div class="auth-input-group">
  <label for="profilePicture" class="auth-label">Profile Picture (optional)</label>
  <div class="aycom-auth-file-input">
    <label class="aycom-auth-file-label" for="profilePicture">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" clip-rule="evenodd" />
      </svg>
      <span class="aycom-auth-file-name">
        {profilePicture instanceof File ? profilePicture.name : 'Choose profile picture'}
      </span>
    </label>
    <input 
      type="file"
      id="profilePicture"
      accept="image/*"
      on:change={handleProfilePictureChange}
      data-cy="profile-picture"
      aria-invalid={profilePictureError ? "true" : "false"}
      aria-describedby={profilePictureError ? "profile-picture-error" : undefined}
    />
  </div>
  {#if profilePictureError}
    <p id="profile-picture-error" class="auth-error-message" data-cy="profile-picture-error" role="alert">{profilePictureError}</p>
  {/if}
</div>

<!-- Banner Picture -->
<div class="auth-input-group">
  <label for="banner" class="auth-label">Banner Image (optional)</label>
  <div class="aycom-auth-file-input">
    <label class="aycom-auth-file-label" for="banner">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" clip-rule="evenodd" />
      </svg>
      <span class="aycom-auth-file-name">
        {banner instanceof File ? banner.name : 'Choose banner image'}
      </span>
    </label>
    <input 
      type="file"
      id="banner"
      accept="image/*"
      on:change={handleBannerChange}
      data-cy="banner"
      aria-invalid={bannerError ? "true" : "false"}
      aria-describedby={bannerError ? "banner-error" : undefined}
    />
  </div>
  {#if bannerError}
    <p id="banner-error" class="auth-error-message" data-cy="banner-error" role="alert">{bannerError}</p>
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
    aria-describedby="newsletter-description"
  />
  <label for="subscribeToNewsletter" class="auth-checkbox-label">
    Subscribe to our newsletter
  </label>
  <span id="newsletter-description" class="sr-only">Receive news and updates about our platform via email</span>
</div>

<!-- reCAPTCHA -->
<div class="recaptcha-wrapper">
  <ReCaptchaWrapper
    bind:this={recaptchaWrapper}
    theme={isDarkMode ? 'dark' : 'light'}
    size="normal"
    position="inline"
    on:success={handleRecaptchaSuccess}
    on:error={handleRecaptchaError}
    on:expired={handleRecaptchaExpired}
  />
</div>

<button 
  type="button"
  on:click={triggerSubmit}
  class="auth-btn"
  data-cy="register-button"
  aria-label="Complete registration"
>
  Sign up
</button>

<p class="text-xs mt-4 text-gray-400 text-center">
  By signing up, you agree to the <a href="/terms" class="text-blue-500 hover:underline">Terms of Service</a> and 
  <a href="/privacy" class="text-blue-500 hover:underline">Privacy Policy</a>, including <a href="/cookies" class="text-blue-500 hover:underline">Cookie Use</a>.
</p>

<style>
  .recaptcha-wrapper {
    margin: 1.5rem 0;
    width: 100%;
  }
</style>