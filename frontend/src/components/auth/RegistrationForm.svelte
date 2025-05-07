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

<GoogleSignInButton
  onAuthSuccess={onGoogleAuthSuccess}
  onAuthError={onGoogleAuthError}
  containerId="google-signin-button"
/>

<div class="flex items-center mb-4">
  <div class="flex-grow h-px bg-gray-300 dark:bg-gray-700"></div>
  <span class="px-2 text-sm text-gray-500 dark:text-gray-400">or</span>
  <div class="flex-grow h-px bg-gray-300 dark:bg-gray-700"></div>
</div>

<!-- Name Input -->
<div class="mb-4">
  <div class="flex justify-between">
    <label for="name" class="block text-sm font-medium mb-1">Name</label>
    <span class="text-xs text-gray-500 dark:text-gray-400" data-cy="name-char-count">{name.length} / 50</span>
  </div>
  <input 
    type="text" 
    id="name" 
    bind:value={name} 
    on:blur={onNameBlur}
    maxlength="50"
    class="w-full p-2 border {isDarkMode ? 'border-gray-700 bg-gray-800' : 'border-gray-300 bg-white'} rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
    placeholder="Name"
    data-cy="name-input"
  />
  {#if nameError}
    <p class="text-red-500 text-xs mt-1" data-cy="name-error">{nameError}</p>
  {/if}
</div>

<!-- Username input -->
<div class="mb-4">
  <label for="username" class="block text-sm font-medium mb-1">Username</label>
  <input 
    type="text" 
    id="username" 
    bind:value={username} 
    on:blur={onUsernameBlur}
    class="w-full p-2 border {isDarkMode ? 'border-gray-700 bg-gray-800' : 'border-gray-300 bg-white'} rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
    placeholder="Username"
    data-cy="username-input"
  />
  {#if usernameError}
    <p class="text-red-500 text-xs mt-1" data-cy="username-error">{usernameError}</p>
  {/if}
</div>

<!-- Email input -->
<div class="mb-4">
  <label for="email" class="block text-sm font-medium mb-1">Email</label>
  <input 
    type="email" 
    id="email" 
    bind:value={email} 
    on:blur={onEmailBlur}
    class="w-full p-2 border {isDarkMode ? 'border-gray-700 bg-gray-800' : 'border-gray-300 bg-white'} rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
    placeholder="Email"
    data-cy="email-input"
  />
  {#if emailError}
    <p class="text-red-500 text-xs mt-1" data-cy="email-error">{emailError}</p>
  {/if}
</div>

<!-- Password input -->
<div class="mb-4">
  <label for="password" class="block text-sm font-medium mb-1">Password</label>
  <input 
    type="password" 
    id="password" 
    bind:value={password} 
    on:blur={onPasswordBlur}
    class="w-full p-2 border {isDarkMode ? 'border-gray-700 bg-gray-800' : 'border-gray-300 bg-white'} rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
    placeholder="Password"
    data-cy="password-input"
  />
  {#if passwordErrors.length > 0}
    <div class="text-red-500 text-xs mt-1" data-cy="password-error">
      {#each passwordErrors as error}
        <p>{error}</p>
      {/each}
    </div>
  {/if}
</div>

<!-- Confirm Password input -->
<div class="mb-4">
  <label for="confirmPassword" class="block text-sm font-medium mb-1">Confirm Password</label>
  <input 
    type="password" 
    id="confirmPassword" 
    bind:value={confirmPassword} 
    on:blur={onConfirmPasswordBlur}
    class="w-full p-2 border {isDarkMode ? 'border-gray-700 bg-gray-800' : 'border-gray-300 bg-white'} rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
    placeholder="Confirm Password"
    data-cy="confirm-password-input"
  />
  {#if confirmPasswordError}
    <p class="text-red-500 text-xs mt-1" data-cy="password-match-error">{confirmPasswordError}</p>
  {/if}
</div>

<!-- Gender selection -->
<div class="mb-4">
  <label class="block text-sm font-medium mb-1">Gender</label>
  <div class="flex space-x-4">
    <label class="flex items-center">
      <input 
        type="radio" 
        name="gender" 
        value="male" 
        bind:group={gender} 
        on:change={onGenderChange}
        class="mr-2"
        data-cy="gender-male"
      />
      <span>Male</span>
    </label>
    <label class="flex items-center">
      <input 
        type="radio" 
        name="gender" 
        value="female" 
        bind:group={gender} 
        on:change={onGenderChange}
        class="mr-2"
        data-cy="gender-female"
      />
      <span>Female</span>
    </label>
  </div>
  {#if genderError}
    <p class="text-red-500 text-xs mt-1" data-cy="gender-error">{genderError}</p>
  {/if}
</div>

<!-- Date of birth -->
<div class="mb-4">
  <label class="block text-sm font-medium mb-1">Date of birth</label>
  <p class="text-xs text-gray-500 dark:text-gray-400 mb-2">This will not be shown publicly. Confirm your own age, even if this account is for a business, a pet, or something else.</p>
  
  <div class="flex space-x-2">
    <div class="w-1/3">
      <select 
        bind:value={dateOfBirth.month} 
        on:change={onDateOfBirthChange}
        class="w-full p-2 border {isDarkMode ? 'border-gray-700 bg-gray-800' : 'border-gray-300 bg-white'} rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
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
        class="w-full p-2 border {isDarkMode ? 'border-gray-700 bg-gray-800' : 'border-gray-300 bg-white'} rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
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
        class="w-full p-2 border {isDarkMode ? 'border-gray-700 bg-gray-800' : 'border-gray-300 bg-white'} rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
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
    <p class="text-red-500 text-xs mt-1" data-cy="dob-error">{dateOfBirthError}</p>
  {/if}
</div>

<!-- Profile picture upload -->
<div class="mb-4">
  <label for="profilePicture" class="block text-sm font-medium mb-1">Profile Picture</label>
  <input 
    type="file" 
    id="profilePicture" 
    accept="image/*" 
    class="w-full p-2 border {isDarkMode ? 'border-gray-700 bg-gray-800' : 'border-gray-300 bg-white'} rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
    data-cy="profile-picture-input"
    on:change={(e) => { 
      const input = e.target as HTMLInputElement;
      if (input.files && input.files.length > 0) {
        profilePicture = input.files[0];
      }
    }}
  />
  {#if profilePictureError}
    <p class="text-red-500 text-xs mt-1" data-cy="profile-picture-error">{profilePictureError}</p>
  {/if}
</div>

<div class="mb-4">
  <label for="banner" class="block text-sm font-medium mb-1">Banner</label>
  <input 
    type="file" 
    id="banner" 
    accept="image/*" 
    class="w-full p-2 border {isDarkMode ? 'border-gray-700 bg-gray-800' : 'border-gray-300 bg-white'} rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
    data-cy="banner-input"
    on:change={(e) => {
      const input = e.target as HTMLInputElement;
      if (input.files && input.files.length > 0) {
        banner = input.files[0];
      }
    }}
  />
  {#if bannerError}
    <p class="text-red-500 text-xs mt-1" data-cy="banner-error">{bannerError}</p>
  {/if}
</div>

<div class="mb-4">
  <label for="securityQuestion" class="block text-sm font-medium mb-1">Security question</label>
  <select 
    id="securityQuestion"
    bind:value={securityQuestion}
    on:change={onSecurityQuestionChange}
    class="w-full p-2 border {isDarkMode ? 'border-gray-700 bg-gray-800' : 'border-gray-300 bg-white'} rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
    data-cy="security-question"
  >
    <option value="">Select a security question</option>
    {#each securityQuestions as question}
      <option value={question}>{question}</option>
    {/each}
  </select>
  {#if securityQuestionError}
    <p class="text-red-500 text-xs mt-1" data-cy="security-question-error">{securityQuestionError}</p>
  {/if}
</div>

<div class="mb-6">
  <label for="securityAnswer" class="block text-sm font-medium mb-1">Security answer</label>
  <input 
    type="text" 
    id="securityAnswer" 
    bind:value={securityAnswer} 
    on:blur={onSecurityAnswerBlur}
    class="w-full p-2 border {isDarkMode ? 'border-gray-700 bg-gray-800' : 'border-gray-300 bg-white'} rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
    placeholder="Your answer"
    data-cy="security-answer"
  />
</div>

<!-- Newsletter subscription -->
<div class="mb-6">
  <label class="flex items-start">
    <input 
      type="checkbox" 
      bind:checked={subscribeToNewsletter} 
      class="mt-1 mr-2"
      data-cy="newsletter-checkbox"
    />
    <span class="text-sm">Subscribe to newsletter and other promotional emails</span>
  </label>
</div>

<button 
  on:click={triggerSubmit} 
  type="button" 
  class="w-full py-3 bg-blue-500 text-white text-center rounded-full font-semibold hover:bg-blue-600 transition-colors"
  data-cy="register-button"
>
  Create account
</button>

<p class="text-xs mt-4 text-gray-400 text-center">
  By signing up, you agree to the <a href="/terms" class="text-blue-500 hover:underline">Terms of Service</a> and 
  <a href="/privacy" class="text-blue-500 hover:underline">Privacy Policy</a>, including <a href="/cookies" class="text-blue-500 hover:underline">Cookie Use</a>.
</p>