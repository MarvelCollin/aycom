<script lang="ts">
  import { useTheme } from "../../hooks/useTheme";
  import GoogleSignInButton from "./GoogleSignInButton.svelte";
  import ReCaptchaWrapper from "./ReCaptchaWrapper.svelte";
  import type { IDateOfBirth, ICustomWindow } from "../../interfaces/IAuth";

  const { theme } = useTheme();

  $: isDarkMode = $theme === "dark";

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

  export let months: string[] = [];
  export let days: string[] = [];
  export let years: string[] = [];
  export let securityQuestions: string[] = [];

  export let nameError = "";
  export let usernameError = "";
  export let emailError = "";
  export let passwordErrors: string[] = [];
  export let confirmPasswordError = "";
  export let genderError = "";
  export let dateOfBirthError = "";
  export let securityQuestionError = "";
  export let securityAnswerError = "";
  export let profilePictureError = "";
  export let bannerError = "";

  export let onSubmit: (token: string | null) => void;
  export let onGoogleAuthSuccess: (result: any) => void;
  export let onGoogleAuthError: (error: string) => void;

  let recaptchaToken: string | null = null;
  let recaptchaWrapper: ReCaptchaWrapper;

  let profilePicturePreview: string | null = null;
  let bannerPreview: string | null = null;

  function handleRecaptchaSuccess(event: CustomEvent<{ token: string }>) {
    recaptchaToken = event.detail.token;
  }

  function handleRecaptchaError() {
    recaptchaToken = null;
  }

  function handleRecaptchaExpired() {
    recaptchaToken = null;
  }

  const isDevelopment = import.meta.env.DEV;

  async function triggerSubmit() {

    if (isDevelopment) {
      onSubmit("dev-mode-token");
      return;
    }

    onSubmit(recaptchaToken);
  }

  function handleProfilePictureChange(e: Event) {
    const target = e.target as HTMLInputElement;
    const file = target.files?.[0] || null;

    if (file) {
      profilePicture = file;

      profilePicturePreview = URL.createObjectURL(file);
    } else {
      profilePicture = null;
      profilePicturePreview = null;
    }
  }

  function handleBannerChange(e: Event) {
    const target = e.target as HTMLInputElement;
    const file = target.files?.[0] || null;

    if (file) {
      banner = file;

      bannerPreview = URL.createObjectURL(file);
    } else {
      banner = null;
      bannerPreview = null;
    }
  }

  import { onDestroy } from "svelte";

  onDestroy(() => {
    if (profilePicturePreview) URL.revokeObjectURL(profilePicturePreview);
    if (bannerPreview) URL.revokeObjectURL(bannerPreview);
  });
</script>

<div class="auth-social-btn-container">
  <GoogleSignInButton
    onAuthSuccess={onGoogleAuthSuccess}
    onAuthError={onGoogleAuthError}
    containerId="google-signin-button"
    class="auth-social-btn {isDarkMode ? "auth-social-btn-dark" : ""}"
  />
</div>

<div class="auth-divider {isDarkMode ? "auth-divider-dark" : ""}">
  <span class="auth-divider-text">or</span>
</div>

<!-- Name Input -->
<div class="auth-input-group">
  <div class="flex justify-between">
    <label for="name" class="auth-label">Name <span class="text-red-500">*</span></label>
    <span class="text-xs text-gray-500 dark:text-gray-400" data-cy="name-char-count">{name.length} / 50</span>
  </div>
  <input
    type="text"
    id="name"
    bind:value={name}
    maxlength="50"
    class="auth-input {isDarkMode ? "auth-input-dark" : ""} {nameError ? "auth-input-error" : ""}"
    placeholder="Name"
    data-cy="name-input"
  />
  {#if nameError}
    <p id="name-error" class="auth-error-message" data-cy="name-error" role="alert">{nameError}</p>
    <div class="text-xs text-gray-500 mt-1">Must be at least 4 characters, no symbols or numbers.</div>
  {/if}
</div>

<!-- Username input -->
<div class="auth-input-group">
  <label for="username" class="auth-label">Username <span class="text-red-500">*</span></label>
  <input
    type="text"
    id="username"
    bind:value={username}
    class="auth-input {isDarkMode ? "auth-input-dark" : ""} {usernameError ? "auth-input-error" : ""}"
    placeholder="Username"
    data-cy="username-input"
  />
  {#if usernameError}
    <p id="username-error" class="auth-error-message" data-cy="username-error" role="alert">{usernameError}</p>
    <div class="text-xs text-gray-500 mt-1">Must be unique, 3-15 characters, letters, numbers, and underscores only.</div>
  {/if}
</div>

<!-- Email input -->
<div class="auth-input-group">
  <label for="email" class="auth-label">Email <span class="text-red-500">*</span></label>
  <input
    type="email"
    id="email"
    bind:value={email}
    class="auth-input {isDarkMode ? "auth-input-dark" : ""} {emailError ? "auth-input-error" : ""}"
    placeholder="Email"
    data-cy="email-input"
  />
  {#if emailError}
    <p id="email-error" class="auth-error-message" data-cy="email-error" role="alert">{emailError}</p>
    <div class="text-xs text-gray-500 mt-1">Must be a valid email format (e.g., name@domain.com) and unique.</div>
  {/if}
</div>

<!-- Password input -->
<div class="auth-input-group">
  <label for="password" class="auth-label">Password <span class="text-red-500">*</span></label>
  <input
    type="password"
    id="password"
    bind:value={password}
    class="auth-input {isDarkMode ? "auth-input-dark" : ""} {passwordErrors.length > 0 ? "auth-input-error" : ""}"
    placeholder="Password"
    data-cy="password-input"
  />
  {#if passwordErrors.length > 0}
    <div id="password-error" class="auth-error-message" data-cy="password-error" role="alert">
      {#each passwordErrors as passError}
        <p>{passError}</p>
      {/each}
    </div>
    <div class="text-xs text-gray-500 mt-1">Must have at least 8 characters, uppercase letter, lowercase letter, number, and special character.</div>
  {/if}
</div>

<!-- Confirm Password input -->
<div class="auth-input-group">
  <label for="confirmPassword" class="auth-label">Confirm Password <span class="text-red-500">*</span></label>
  <input
    type="password"
    id="confirmPassword"
    bind:value={confirmPassword}
    class="auth-input {isDarkMode ? "auth-input-dark" : ""} {confirmPasswordError ? "auth-input-error" : ""}"
    placeholder="Confirm Password"
    data-cy="confirm-password-input"
  />
  {#if confirmPasswordError}
    <p id="password-match-error" class="auth-error-message" data-cy="password-match-error" role="alert">{confirmPasswordError}</p>
    <div class="text-xs text-gray-500 mt-1">Must exactly match the password field.</div>
  {/if}
</div>

<!-- Gender selection -->
<div class="auth-input-group">
  <fieldset>
    <legend class="auth-label">Gender <span class="text-red-500">*</span></legend>
    <div class="auth-radio-group">
      <label class="auth-checkbox-group">
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
    {#if genderError}
      <p id="gender-error" class="auth-error-message" data-cy="gender-error" role="alert">{genderError}</p>
      <div class="text-xs text-gray-500 mt-1">Must select either male or female.</div>
    {/if}
  </fieldset>
</div>

<!-- Date of birth -->
<div class="auth-input-group">
  <fieldset>
    <legend class="auth-label">Date of birth <span class="text-red-500">*</span></legend>
    <p class="auth-helper-text mb-2">This will not be shown publicly. Confirm your own age, even if this account is for a business, a pet, or something else.</p>

    <div class="auth-dob-container">
      <div class="auth-dob-select-group" role="group" aria-labelledby="dob-label">
        <div>
          <label for="dob-month" class="sr-only">Month</label>
          <select
            id="dob-month"
            bind:value={dateOfBirth.month}
            class="auth-input {isDarkMode ? "auth-input-dark" : ""} {dateOfBirthError ? "auth-input-error" : ""}"
            data-cy="dob-month"
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
            class="auth-input {isDarkMode ? "auth-input-dark" : ""} {dateOfBirthError ? "auth-input-error" : ""}"
            data-cy="dob-day"
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
            class="auth-input {isDarkMode ? "auth-input-dark" : ""} {dateOfBirthError ? "auth-input-error" : ""}"
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
        <p id="dob-error" class="auth-error-message" data-cy="dob-error" role="alert">{dateOfBirthError}</p>
        <div class="text-xs text-gray-500 mt-1">You must be at least 13 years old to register.</div>
      {/if}
    </div>
  </fieldset>
</div>

<!-- Security Question -->
<div class="auth-input-group">
  <label for="securityQuestion" class="auth-label">Security Question <span class="text-red-500">*</span></label>
  <select
    id="securityQuestion"
    bind:value={securityQuestion}
    class="auth-input {isDarkMode ? "auth-input-dark" : ""} {securityQuestionError ? "auth-input-error" : ""}"
    data-cy="security-question"
  >
    <option value="">Select a security question</option>
    {#each securityQuestions as question}
      <option value={question}>{question}</option>
    {/each}
  </select>
  {#if securityQuestionError}
    <p id="security-question-error" class="auth-error-message" data-cy="security-question-error" role="alert">{securityQuestionError}</p>
    <div class="text-xs text-gray-500 mt-1">You must select a security question.</div>
  {/if}
</div>

<!-- Security Answer -->
<div class="auth-input-group">
  <label for="securityAnswer" class="auth-label">Security Answer <span class="text-red-500">*</span></label>
  <input
    type="text"
    id="securityAnswer"
    bind:value={securityAnswer}
    class="auth-input {isDarkMode ? "auth-input-dark" : ""} {securityAnswerError ? "auth-input-error" : ""}"
    placeholder="Your answer"
    data-cy="security-answer"
  />
  {#if securityAnswerError}
    <p id="security-answer-error" class="auth-error-message" data-cy="security-answer-error" role="alert">{securityAnswerError}</p>
    <div class="text-xs text-gray-500 mt-1">You must provide an answer to your security question.</div>
  {/if}
</div>

<!-- Profile Picture -->
<div class="auth-input-group">
  <label for="profilePicture" class="auth-label">Profile Picture (optional)</label>
  <div class="aycom-auth-file-input">
    {#if profilePicturePreview}
      <div class="aycom-auth-image-preview">
        <img src={profilePicturePreview} alt="Profile preview" class="aycom-auth-preview-img" />
        <button type="button" class="aycom-auth-remove-img"
          on:click={() => {
            profilePicture = null;
            profilePicturePreview = null;
          }} aria-label="Remove image">
          ×
        </button>
      </div>
    {:else}
      <label class="aycom-auth-file-label" for="profilePicture">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" clip-rule="evenodd" />
        </svg>
        <span class="aycom-auth-file-name">
          Choose profile picture
        </span>
      </label>
    {/if}
    <input
      type="file"
      id="profilePicture"
      accept="image/*"
      on:change={handleProfilePictureChange}
      data-cy="profile-picture"
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
    {#if bannerPreview}
      <div class="aycom-auth-image-preview">
        <img src={bannerPreview} alt="Banner preview" class="aycom-auth-preview-img aycom-auth-banner-preview" />
        <button type="button" class="aycom-auth-remove-img"
          on:click={() => {
            banner = null;
            bannerPreview = null;
          }} aria-label="Remove image">
          ×
        </button>
      </div>
    {:else}
      <label class="aycom-auth-file-label" for="banner">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" clip-rule="evenodd" />
        </svg>
        <span class="aycom-auth-file-name">
          Choose banner image
        </span>
      </label>
    {/if}
    <input
      type="file"
      id="banner"
      accept="image/*"
      on:change={handleBannerChange}
      data-cy="banner"
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
    theme={isDarkMode ? "dark" : "light"}
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

  .aycom-auth-image-preview {
    position: relative;
    margin-bottom: var(--space-2);
    border-radius: var(--radius-md);
    overflow: hidden;
    max-width: 100%;
  }

  .aycom-auth-preview-img {
    width: 100%;
    max-height: 200px;
    object-fit: cover;
    display: block;
    border-radius: var(--radius-md);
  }

  .aycom-auth-banner-preview {
    height: 100px;
  }

  .aycom-auth-remove-img {
    position: absolute;
    top: 5px;
    right: 5px;
    background-color: rgba(0, 0, 0, 0.5);
    color: white;
    border: none;
    border-radius: 50%;
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 18px;
    cursor: pointer;
    padding: 0;
    line-height: 1;
  }

  .aycom-auth-remove-img:hover {
    background-color: rgba(0, 0, 0, 0.7);
  }

  :global(.auth-error-message) {
    color: #dc2626;
    font-size: 0.85rem;
    margin-top: 0.25rem;
    font-weight: 500;
    display: flex;
    align-items: center;
  }

  :global(.auth-error-message::before) {
    content: "⚠️";
    margin-right: 0.25rem;
  }

  :global(.auth-input-error) {
    border-color: #dc2626 !important;
    background-color: rgba(255, 0, 0, 0.03);
  }
</style>