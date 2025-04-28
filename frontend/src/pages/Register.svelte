<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import AuthLayout from '../components/layout/AuthLayout.svelte';
  import RegistrationForm from '../components/auth/RegistrationForm.svelte';
  import VerificationForm from '../components/auth/VerificationForm.svelte';
  import { useRegistrationForm } from '../hooks/useRegistrationForm';
  import { useAuth } from '../hooks/useAuth';
  import { useExternalServices } from '../hooks/useExternalServices';
  import type { IUserRegistration } from '../interfaces/IAuth';
  
  // Get registration form functionality
  const {
    formData,
    errors,
    formState,
    months,
    days,
    years,
    securityQuestions,
    validateFormField,
    validateStep1,
    startTimer,
    formatTimeLeft,
    cleanupTimers
  } = useRegistrationForm();
  
  // Get auth functions
  const { register, verifyEmail, resendVerificationCode } = useAuth();
  
  // Get external services
  const { loadRecaptcha } = useExternalServices();
  
  // Validation wrapper functions
  function validateNameAndUpdate() {
    $formData.name && validateFormField('name', $formData.name);
  }
  
  function validateUsernameAndUpdate() {
    $formData.username && validateFormField('username', $formData.username);
  }
  
  function validateEmailAndUpdate() {
    $formData.email && validateFormField('email', $formData.email);
  }
  
  function validatePasswordAndUpdate() {
    $formData.password && validateFormField('password', $formData.password);
  }
  
  function validateConfirmPasswordAndUpdate() {
    $formData.confirmPassword && validateFormField('confirmPassword', $formData.confirmPassword);
  }
  
  function validateGenderAndUpdate() {
    $formData.gender && validateFormField('gender', $formData.gender);
  }
  
  function validateDateOfBirthAndUpdate() {
    validateFormField('dateOfBirth', $formData.dateOfBirth);
  }
  
  function validateSecurityQuestionAndUpdate() {
    validateFormField('securityQuestion', $formData.securityQuestion);
  }
  
  function validateSecurityAnswerAndUpdate() {
    validateFormField('securityAnswer', $formData.securityAnswer);
  }
  
  // Make the function async to use await
  async function submitRegistration() {
    if (validateStep1()) {
      // Update loading state
      formState.update(state => ({ ...state, loading: true }));
      
      // Prepare registration data
      const userData: IUserRegistration = {
        name: $formData.name,
        username: $formData.username,
        email: $formData.email,
        password: $formData.password,
        confirm_password: $formData.confirmPassword,
        gender: $formData.gender,
        date_of_birth: months.indexOf($formData.dateOfBirth.month) + '-' + $formData.dateOfBirth.day + '-' + $formData.dateOfBirth.year,
        security_question: $formData.securityQuestion,
        security_answer: $formData.securityAnswer,
        subscribe_to_newsletter: $formData.subscribeToNewsletter,
        recaptcha_token: ''
      };
      
      // Call the register function from auth hook
      const result = await register(userData);
      
      // Update loading state
      formState.update(state => ({ ...state, loading: false }));
      
      if (result.success) {
        // Start timer for verification code
        startTimer();
        
        // Move to step 2
        formState.update(state => ({ ...state, step: 2 }));
      } else {
        // Display error message
        formState.update(state => ({ ...state, error: result.message || "Registration failed. Please try again." }));
      }
    }
  }
  
  // Handle Google authentication success
  function handleGoogleAuthSuccess(result: any) {
    window.location.href = '/feed';
  }
  
  // Handle Google authentication error
  function handleGoogleAuthError(error: string) {
    formState.update(state => ({ ...state, error }));
  }
  
  // Make the function async to use await
  async function submitVerification() {
    if (!$formData.verificationCode) {
      formState.update(state => ({ ...state, error: "Please enter the verification code sent to your email" }));
      return;
    }
    
    // Update loading state
    formState.update(state => ({ ...state, loading: true }));
    
    // Call the verifyEmail function
    const result = await verifyEmail($formData.email, $formData.verificationCode);
    
    // Update loading state
    formState.update(state => ({ ...state, loading: false }));
    
    if (result.success) {
      // Redirect to login page
      window.location.href = '/login';
    } else {
      formState.update(state => ({ ...state, error: result.message || "Verification failed. Please check your code and try again." }));
    }
  }
  
  // Make the function async to use await
  async function resendCode() {
    // Update loading state
    formState.update(state => ({ ...state, loading: true }));
    
    // Call the resendVerificationCode function
    const result = await resendVerificationCode($formData.email);
    
    // Update loading state
    formState.update(state => ({ ...state, loading: false }));
    
    if (result.success) {
      formState.update(state => ({ ...state, showResendOption: false }));
      startTimer();
      alert("Verification code has been sent to your email.");
    } else {
      formState.update(state => ({ ...state, error: result.message || "Failed to resend verification code." }));
    }
  }
  
  // Handle back button
  function goBack() {
    formState.update(state => ({ ...state, step: 1, error: "" }));
  }

  onMount(() => {
    // Load reCAPTCHA
    const recaptchaCleanup = loadRecaptcha((token) => {
      // This callback will be called when the token is updated
      console.log("reCAPTCHA token updated");
    });
    
    return () => {
      recaptchaCleanup();
    };
  });
  
  onDestroy(() => {
    // Clean up timers
    cleanupTimers();
  });
</script>

<AuthLayout 
  title={$formState.step === 1 ? "Create your account" : "We sent you a code"}
  showBackButton={$formState.step === 2} 
  onBack={goBack}
>
  {#if $formState.error}
    <div class="bg-red-500 bg-opacity-10 border border-red-500 text-red-500 px-4 py-3 rounded mb-4">
      {$formState.error}
    </div>
  {/if}
  
  {#if $formState.step === 1}
    <RegistrationForm
      bind:name={$formData.name}
      bind:username={$formData.username}
      bind:email={$formData.email}
      bind:password={$formData.password}
      bind:confirmPassword={$formData.confirmPassword}
      bind:gender={$formData.gender}
      bind:dateOfBirth={$formData.dateOfBirth}
      bind:profilePicture={$formData.profilePicture}
      bind:banner={$formData.banner}
      bind:securityQuestion={$formData.securityQuestion}
      bind:securityAnswer={$formData.securityAnswer}
      bind:subscribeToNewsletter={$formData.subscribeToNewsletter}
      {months}
      {days}
      {years}
      {securityQuestions}
      nameError={$errors.name}
      usernameError={$errors.username}
      emailError={$errors.email}
      passwordErrors={$errors.password}
      confirmPasswordError={$errors.confirmPassword}
      genderError={$errors.gender}
      dateOfBirthError={$errors.dateOfBirth}
      securityQuestionError={$errors.securityQuestion}
      profilePictureError={$errors.profilePicture}
      bannerError={$errors.banner}
      onNameBlur={validateNameAndUpdate}
      onUsernameBlur={validateUsernameAndUpdate}
      onEmailBlur={validateEmailAndUpdate}
      onPasswordBlur={validatePasswordAndUpdate}
      onConfirmPasswordBlur={validateConfirmPasswordAndUpdate}
      onGenderChange={validateGenderAndUpdate}
      onDateOfBirthChange={validateDateOfBirthAndUpdate}
      onSecurityQuestionChange={validateSecurityQuestionAndUpdate}
      onSecurityAnswerBlur={validateSecurityAnswerAndUpdate}
      onSubmit={submitRegistration}
      onGoogleAuthSuccess={handleGoogleAuthSuccess}
      onGoogleAuthError={handleGoogleAuthError}
    />
    
    <!-- Login link -->
    <p class="text-sm mt-6 text-center">
      Already have an account? <a href="/login" class="text-blue-500 hover:underline">Sign in</a>
    </p>
  {:else}
    <!-- Step 2: Verification Code Input -->
    <p class="text-center mb-6">Enter it below to verify {$formData.email}</p>
    
    <VerificationForm
      bind:verificationCode={$formData.verificationCode}
      showResendOption={$formState.showResendOption}
      timeLeft={formatTimeLeft()}
      onVerify={submitVerification}
      onResend={resendCode}
    />
  {/if}
</AuthLayout>
