<script lang="ts">
  import { onMount } from 'svelte';
  import { useTheme } from '../hooks/useTheme';
  import { useAuth } from '../hooks/useAuth';
  import { useValidation } from '../hooks/useValidation';
  import { useExternalServices } from '../hooks/useExternalServices';
  import type { DateOfBirth, GoogleCredentialResponse } from '../interfaces/auth';
  
  interface GoogleAccountsId {
    initialize: (config: any) => void;
    renderButton: (element: HTMLElement, options: any) => void;
  }
  
  interface GoogleAccounts {
    id: GoogleAccountsId;
  }
  
  interface Google {
    accounts: GoogleAccounts;
  }
  
  interface RecaptchaInstance {
    ready: (callback: () => void) => void;
    render: (container: string, options: any) => number;
  }
  
  interface CustomWindow extends Window {
    google?: Google;
    grecaptcha?: RecaptchaInstance;
  }
  
  const { theme } = useTheme();
  
  const { register, verifyEmail, resendVerificationCode, handleGoogleAuth } = useAuth();
  
  const { 
    validateName, 
    validateUsername, 
    validateEmail, 
    validatePassword, 
    validateConfirmPassword, 
    validateGender, 
    validateDateOfBirth, 
    validateSecurityQuestion, 
    formatDateOfBirth 
  } = useValidation();
  
  // Get external services functions
  const { loadRecaptcha, loadGoogleAuth, getRecaptchaToken } = useExternalServices();
  
  // Reactive declaration to update isDarkMode when theme changes
  $: isDarkMode = $theme === 'dark';
  
  // Registration state
  let step = 1;
  let name = "";
  let email = "";
  let username = "";
  let password = "";
  let confirmPassword = "";
  let gender = "";
  let dateOfBirth: DateOfBirth = {
    month: "",
    day: "",
    year: ""
  };
  let profilePicture: File | null = null;
  let banner: File | null = null;
  let securityQuestion = "";
  let securityAnswer = "";
  let subscribeToNewsletter = false;
  let verificationCode = "";
  let showResendOption = false;
  
  let recaptchaToken = "";
  let recaptchaLoaded = false;
  let googleAuthLoaded = false;
  
  // Validation states
  let nameError = "";
  let usernameError = "";
  let emailError = "";
  let passwordErrors: string[] = [];
  let confirmPasswordError = "";
  let genderError = "";
  let dateOfBirthError = "";
  let securityQuestionError = "";
  
  // List of security questions
  const securityQuestions = [
    "What was the name of your first pet?",
    "What city were you born in?",
    "What is your favorite video game?",
    "What was the name of your first school?",
    "What was your childhood nickname?"
  ];
  
  // List of months for date of birth dropdown
  const months = [
    "January", "February", "March", "April", "May", "June", 
    "July", "August", "September", "October", "November", "December"
  ];
  
  // Generate days 1-31
  const days = Array.from({ length: 31 }, (_, i) => (i + 1).toString());
  
  // Generate years from current year - 100 to current year
  const currentYear = new Date().getFullYear();
  const years = Array.from({ length: 100 }, (_, i) => (currentYear - i).toString());
  
  // Timer for verification code
  let timeLeft = 300; // 5 minutes in seconds
  let timerId: number | undefined;
  
  // Validation wrapper functions that update error states
  function validateNameAndUpdate() {
    nameError = validateName(name);
    return !nameError;
  }
  
  function validateUsernameAndUpdate() {
    usernameError = validateUsername(username);
    return !usernameError;
  }
  
  function validateEmailAndUpdate() {
    emailError = validateEmail(email);
    return !emailError;
  }
  
  function validatePasswordAndUpdate() {
    passwordErrors = validatePassword(password);
    return passwordErrors.length === 0;
  }
  
  function validateConfirmPasswordAndUpdate() {
    confirmPasswordError = validateConfirmPassword(password, confirmPassword);
    return !confirmPasswordError;
  }
  
  function validateGenderAndUpdate() {
    genderError = validateGender(gender);
    return !genderError;
  }
  
  function validateDateOfBirthAndUpdate() {
    dateOfBirthError = validateDateOfBirth(dateOfBirth, months);
    return !dateOfBirthError;
  }
  
  function validateSecurityQuestionAndUpdate() {
    securityQuestionError = validateSecurityQuestion(securityQuestion, securityAnswer);
    return !securityQuestionError;
  }
  
  function validateStep1() {
    const isNameValid = validateNameAndUpdate();
    const isUsernameValid = validateUsernameAndUpdate();
    const isEmailValid = validateEmailAndUpdate();
    const isPasswordValid = validatePasswordAndUpdate();
    const isConfirmPasswordValid = validateConfirmPasswordAndUpdate();
    const isGenderValid = validateGenderAndUpdate();
    const isDateOfBirthValid = validateDateOfBirthAndUpdate();
    const isSecurityQuestionValid = validateSecurityQuestionAndUpdate();
    
    return isNameValid && isUsernameValid && isEmailValid && isPasswordValid && 
           isConfirmPasswordValid && isGenderValid && isDateOfBirthValid && isSecurityQuestionValid;
  }
  
  // Make the function async to use await
  async function submitStep1() {
    if (validateStep1()) {
      // Check if reCAPTCHA is completed
      const recaptchaToken = getRecaptchaToken();
      if (!recaptchaToken) {
        alert("Please complete the reCAPTCHA verification");
        return;
      }
      
      // Prepare registration data
      const userData = {
        name,
        username,
        email,
        password,
        confirm_password: confirmPassword,
        gender,
        date_of_birth: formatDateOfBirth(dateOfBirth, months),
        security_question: securityQuestion,
        security_answer: securityAnswer,
        subscribe_to_newsletter: subscribeToNewsletter,
        recaptcha_token: recaptchaToken
      };
      
      // Call the register function from auth hook
      const result = await register(userData);
      
      if (result.success) {
        // Start timer for verification code
        startTimer();
        
        // Move to step 2
        step = 2;
      } else {
        // Display error message
        alert(result.message || "Registration failed. Please try again.");
      }
    }
  }
  
  function startTimer() {
    timeLeft = 300; // Reset to 5 minutes
    clearInterval(timerId);
    
    timerId = window.setInterval(() => {
      timeLeft -= 1;
      
      if (timeLeft <= 0) {
        clearInterval(timerId);
        showResendOption = true;
      }
    }, 1000);
  }
  
  function formatTimeLeft() {
    const minutes = Math.floor(timeLeft / 60);
    const seconds = timeLeft % 60;
    return `${minutes}:${seconds < 10 ? '0' + seconds : seconds}`;
  }
  
  // Make the function async to use await
  async function resendCode() {
    // Call the resendVerificationCode function from auth hook
    const result = await resendVerificationCode(email);
    
    if (result.success) {
      showResendOption = false;
      startTimer();
      alert("Verification code has been sent to your email.");
    } else {
      alert(result.message || "Failed to resend verification code.");
    }
  }
  
  // Make the function async to use await
  async function verifyCode() {
    // Call the verifyEmail function from auth hook
    const result = await verifyEmail(email, verificationCode);
    
    if (result.success) {
      // Redirect to login page
      window.location.href = '/login';
    } else {
      alert("Verification failed. Please check your code and try again.");
    }
  }
  
  function goBack() {
    step = 1;
    clearInterval(timerId);
  }
  
  // Google authentication handler
  async function handleGoogleCredentialResponse(response: GoogleCredentialResponse) {
    // Call the handleGoogleAuth function from auth hook
    const result = await handleGoogleAuth(response);
    
    if (result.success) {
      // Redirect to dashboard
      window.location.href = '/dashboard';
    } else {
      console.error('Google authentication failed:', result.message);
      alert(`Google authentication failed: ${result.message || 'Unknown error'}`);
    }
  }
  
  onMount(() => {
    // Clean up timer on component unmount
    const cleanupFn = () => {
      clearInterval(timerId);
    };
    
    // Load reCAPTCHA
    const recaptchaCleanup = loadRecaptcha((token) => {
      // This callback will be called when the token is updated
      console.log("reCAPTCHA token updated");
    });
    
    // Load Google Sign-In
    const googleCleanup = loadGoogleAuth(
      'google-signin-button', 
      isDarkMode, 
      handleGoogleCredentialResponse
    );
    
    // Return cleanup function
    return () => {
      cleanupFn();
      recaptchaCleanup();
      googleCleanup();
    };
  });
</script>

<div class="{isDarkMode ? 'bg-black text-white' : 'bg-white text-black'} min-h-screen w-full flex justify-center items-center p-4">
  <div class="w-full max-w-md bg-dark-900 rounded-lg shadow-lg p-6">
    <div class="flex items-center justify-between mb-6">
      {#if step === 2}
        <button 
          class="text-blue-500 hover:text-blue-600 transition-colors"
          on:click={goBack}
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
        </button>
      {:else}
        <a href="/" class="text-blue-500 hover:text-blue-600 transition-colors">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </a>
      {/if}
      
      <div class="mx-auto">
        <img 
          src={isDarkMode ? "/src/assets/logo/light-logo.jpeg" : "/src/assets/logo/dark-logo.jpeg"} 
          alt="AYCOM Logo" 
          class="h-8 w-auto"
        />
      </div>
    </div>
    
    {#if step === 1}
      <h1 class="text-2xl font-bold mb-6 text-center">Create your account</h1>
      
      <div id="google-signin-button" class="w-full mb-4"></div>
      
      <div class="flex items-center mb-4">
        <div class="flex-grow h-px bg-gray-600"></div>
        <span class="px-2 text-sm text-gray-400">or</span>
        <div class="flex-grow h-px bg-gray-600"></div>
      </div>
      
      <div class="mb-4">
        <div class="flex justify-between">
          <label for="name" class="block text-sm font-medium mb-1">Name</label>
          <span class="text-xs text-gray-400">{name.length} / 50</span>
        </div>
        <input 
          type="text" 
          id="name" 
          bind:value={name} 
          on:blur={validateNameAndUpdate}
          maxlength="50"
          class="w-full p-2 border border-gray-600 rounded bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Name"
        />
        {#if nameError}
          <p class="text-red-500 text-xs mt-1">{nameError}</p>
        {/if}
      </div>
      
      <!-- Username input -->
      <div class="mb-4">
        <label for="username" class="block text-sm font-medium mb-1">Username</label>
        <input 
          type="text" 
          id="username" 
          bind:value={username} 
          on:blur={validateUsernameAndUpdate}
          class="w-full p-2 border border-gray-600 rounded bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Username"
        />
        {#if usernameError}
          <p class="text-red-500 text-xs mt-1">{usernameError}</p>
        {/if}
      </div>
      
      <!-- Email input -->
      <div class="mb-4">
        <label for="email" class="block text-sm font-medium mb-1">Email</label>
        <input 
          type="email" 
          id="email" 
          bind:value={email} 
          on:blur={validateEmailAndUpdate}
          class="w-full p-2 border border-gray-600 rounded bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Email"
        />
        {#if emailError}
          <p class="text-red-500 text-xs mt-1">{emailError}</p>
        {/if}
      </div>
      
      <!-- Password input -->
      <div class="mb-4">
        <label for="password" class="block text-sm font-medium mb-1">Password</label>
        <input 
          type="password" 
          id="password" 
          bind:value={password} 
          on:blur={validatePasswordAndUpdate}
          class="w-full p-2 border border-gray-600 rounded bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Password"
        />
        {#if passwordErrors.length > 0}
          <div class="text-red-500 text-xs mt-1">
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
          on:blur={validateConfirmPasswordAndUpdate}
          class="w-full p-2 border border-gray-600 rounded bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Confirm Password"
        />
        {#if confirmPasswordError}
          <p class="text-red-500 text-xs mt-1">{confirmPasswordError}</p>
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
              on:change={validateGenderAndUpdate}
              class="mr-2"
            />
            <span>Male</span>
          </label>
          <label class="flex items-center">
            <input 
              type="radio" 
              name="gender" 
              value="female" 
              bind:group={gender} 
              on:change={validateGenderAndUpdate}
              class="mr-2"
            />
            <span>Female</span>
          </label>
        </div>
        {#if genderError}
          <p class="text-red-500 text-xs mt-1">{genderError}</p>
        {/if}
      </div>
      
      <!-- Date of birth -->
      <div class="mb-4">
        <label class="block text-sm font-medium mb-1">Date of birth</label>
        <p class="text-xs text-gray-400 mb-2">This will not be shown publicly. Confirm your own age, even if this account is for a business, a pet, or something else.</p>
        
        <div class="flex space-x-2">
          <div class="w-1/3">
            <select 
              bind:value={dateOfBirth.month} 
              on:change={validateDateOfBirthAndUpdate}
              class="w-full p-2 border border-gray-600 rounded bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="" disabled selected>Month</option>
              {#each months as month}
                <option value={month}>{month}</option>
              {/each}
            </select>
          </div>
          <div class="w-1/3">
            <select 
              bind:value={dateOfBirth.day} 
              on:change={validateDateOfBirthAndUpdate}
              class="w-full p-2 border border-gray-600 rounded bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="" disabled selected>Day</option>
              {#each days as day}
                <option value={day}>{day}</option>
              {/each}
            </select>
          </div>
          <div class="w-1/3">
            <select 
              bind:value={dateOfBirth.year} 
              on:change={validateDateOfBirthAndUpdate}
              class="w-full p-2 border border-gray-600 rounded bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="" disabled selected>Year</option>
              {#each years as year}
                <option value={year}>{year}</option>
              {/each}
            </select>
          </div>
        </div>
        
        {#if dateOfBirthError}
          <p class="text-red-500 text-xs mt-1">{dateOfBirthError}</p>
        {/if}
      </div>
      
      <!-- Profile picture upload -->
      <div class="mb-4">
        <label for="profilePicture" class="block text-sm font-medium mb-1">Profile Picture</label>
        <input 
          type="file" 
          id="profilePicture" 
          accept="image/*" 
          class="w-full p-2 border border-gray-600 rounded bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>
      
      <!-- Banner upload -->
      <div class="mb-4">
        <label for="banner" class="block text-sm font-medium mb-1">Banner</label>
        <input 
          type="file" 
          id="banner" 
          accept="image/*" 
          class="w-full p-2 border border-gray-600 rounded bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>
      
      <!-- Security question -->
      <div class="mb-4">
        <label for="securityQuestion" class="block text-sm font-medium mb-1">Security Question</label>
        <select 
          id="securityQuestion" 
          bind:value={securityQuestion}
          on:change={validateSecurityQuestionAndUpdate}
          class="w-full p-2 border border-gray-600 rounded bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500 mb-2"
        >
          <option value="" disabled selected>Select a security question</option>
          {#each securityQuestions as question}
            <option value={question}>{question}</option>
          {/each}
        </select>
        
        <input 
          type="text" 
          bind:value={securityAnswer} 
          on:blur={validateSecurityQuestionAndUpdate}
          class="w-full p-2 border border-gray-600 rounded bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Your answer"
        />
        
        {#if securityQuestionError}
          <p class="text-red-500 text-xs mt-1">{securityQuestionError}</p>
        {/if}
      </div>
      
      <!-- Newsletter subscription -->
      <div class="mb-6">
        <label class="flex items-center">
          <input 
            type="checkbox" 
            bind:checked={subscribeToNewsletter} 
            class="mr-2"
          />
          <span class="text-sm">Subscribe to newsletter</span>
        </label>
      </div>
      
      <!-- reCAPTCHA placeholder -->
      <div id="recaptcha-container" class="mb-6"></div>
      
      <!-- Submit button -->
      <button 
        class="w-full py-3 bg-blue-500 text-white text-center rounded-full font-semibold hover:bg-blue-600 transition-colors"
        on:click={submitStep1}
      >
        Next
      </button>
      
      <!-- Terms & services text -->
      <p class="text-xs mt-4 text-gray-400 text-center">
        By signing up, you agree to the <a href="#" class="text-blue-500 hover:underline">Terms of Service</a> and 
        <a href="#" class="text-blue-500 hover:underline">Privacy Policy</a>, including <a href="#" class="text-blue-500 hover:underline">Cookie Use</a>.
      </p>
      
      <!-- Login link -->
      <p class="text-sm mt-6 text-center">
        Already have an account? <a href="/login" class="text-blue-500 hover:underline">Sign in</a>
      </p>
    {:else}
      <!-- Step 2: Verification Code Input -->
      <h1 class="text-2xl font-bold mb-6 text-center">We sent you a code</h1>
      <p class="text-center mb-6">Enter it below to verify {email}</p>
      
      <div class="mb-6">
        <label for="verificationCode" class="block text-sm font-medium mb-1">Verification code</label>
        <input 
          type="text" 
          id="verificationCode" 
          bind:value={verificationCode} 
          class="w-full p-2 border border-gray-600 rounded bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Verification code"
        />
      </div>
      
      {#if !showResendOption}
        <p class="text-sm text-center mb-4">Code expires in {formatTimeLeft()}</p>
      {/if}
      
      {#if showResendOption}
        <button 
          class="w-full text-blue-500 hover:underline mb-4 text-center"
          on:click={resendCode}
        >
          Didn't receive email?
        </button>
      {/if}
      
      <button 
        class="w-full py-3 bg-blue-500 text-white text-center rounded-full font-semibold hover:bg-blue-600 transition-colors"
        on:click={verifyCode}
      >
        Next
      </button>
    {/if}
  </div>
</div>
