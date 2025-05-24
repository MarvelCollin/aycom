<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { useTheme } from '../hooks/useTheme';
  import AuthLayout from '../components/layout/AuthLayout.svelte';
  import RegistrationForm from '../components/auth/RegistrationForm.svelte';
  import VerificationForm from '../components/auth/VerificationForm.svelte';
  import { useRegistrationForm } from '../hooks/useRegistrationForm';
  import { useAuth } from '../hooks/useAuth';
  import type { IUserRegistration } from '../interfaces/IAuth';
  import { toastStore } from '../stores/toastStore';
  import appConfig from '../config/appConfig';
  import DebugPanel from '../components/common/DebugPanel.svelte';
  
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
  
  const { register, verifyEmail, resendVerificationCode, registerWithMedia } = useAuth();
  
  const { theme } = useTheme();
  
  $: isDarkMode = $theme === 'dark';
  
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
  
  async function submitRegistration(recaptchaToken: string | null) {
    if (!recaptchaToken && !import.meta.env.DEV) {
      const errorMessage = "reCAPTCHA verification failed. Please try again.";
      formState.update(state => ({ ...state, error: errorMessage }));
      if (appConfig.ui.showErrorToasts) toastStore.showToast(errorMessage);
      return;
    }
    
    let errorMessage = "";
    if (validateStep1()) {
      formState.update(state => ({ ...state, loading: true }));
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
        recaptcha_token: recaptchaToken || (import.meta.env.DEV ? "dev-mode-token" : "")
      };
      try {
        let result;
        if ($formData.profilePicture instanceof File || $formData.banner instanceof File) {
          result = await registerWithMedia(
            userData, 
            $formData.profilePicture instanceof File ? $formData.profilePicture : null,
            $formData.banner instanceof File ? $formData.banner : null
          );
        } else {
          result = await register(userData);
        }
        
        formState.update(state => ({ ...state, loading: false }));
        if (result.success) {
          startTimer();
          formState.update(state => ({ ...state, step: 2, error: "" }));
          const successEl = document.createElement('div');
          successEl.textContent = "Registration successful";
          successEl.setAttribute('data-cy', 'success-message');
          successEl.style.position = 'absolute';
          successEl.style.left = '-9999px';
          document.body.appendChild(successEl);
          toastStore.showToast(
            "Registration successful. Please check your email to verify your account.", 
            "success"
          );
        } else {
          errorMessage = result.message || "Registration failed. Please try again.";
          formState.update(state => ({ ...state, error: errorMessage }));
          if (appConfig.ui.showErrorToasts) toastStore.showToast(`Registration Error: ${errorMessage}`);
        }
      } catch (err) {
        formState.update(state => ({ ...state, loading: false }));
        console.error("Registration Exception:", err);
        toastStore.showToast('Registration failed. Please try again.', 'error');
      }
    } else {
      errorMessage = "Please correct the errors in the form.";
      if (appConfig.ui.showErrorToasts) toastStore.showToast(errorMessage);
    }
  }
  
  function handleGoogleAuthSuccess(result: any) {
    console.log('Google auth success in Register page');
    toastStore.showToast('Google registration successful', 'success'); 
    window.location.href = '/feed';
  }
  
  function handleGoogleAuthError(errorMsg: string) {
    console.error('Google auth error in Register page:', errorMsg);
    formState.update(state => ({ ...state, error: errorMsg }));
    if (appConfig.ui.showErrorToasts) toastStore.showToast(`Google Auth Error: ${errorMsg}`, 'error');
  }
  
  async function submitVerification() {
    let errorMessage = "";
    if (!$formData.verificationCode) {
      errorMessage = "Please enter the verification code sent to your email";
      formState.update(state => ({ ...state, error: errorMessage }));
      if (appConfig.ui.showErrorToasts) toastStore.showToast(errorMessage);
      return;
    }
    
    formState.update(state => ({ ...state, loading: true }));
    
    try {
      const result = await verifyEmail($formData.email, $formData.verificationCode);
      formState.update(state => ({ ...state, loading: false }));
      
      if (result.success) {
        window.location.href = '/login';
      } else {
        errorMessage = result.message || "Verification failed. Please check your code and try again.";
        formState.update(state => ({ ...state, error: errorMessage }));
        if (appConfig.ui.showErrorToasts) toastStore.showToast(`Verification Error: ${errorMessage}`);
      }
    } catch (err) {
      formState.update(state => ({ ...state, loading: false }));
      console.error("Verification Exception:", err);
      toastStore.showToast('Verification failed. Please try again.', 'error');
    }
  }
  
  async function resendCode() {
    let errorMessage = "";
    formState.update(state => ({ ...state, loading: true }));
    
    try {
      const result = await resendVerificationCode($formData.email);
      formState.update(state => ({ ...state, loading: false }));
      
      if (result.success) {
        formState.update(state => ({ ...state, showResendOption: false, error: "" }));
        startTimer();
        toastStore.showToast("Verification code has been resent.", "success"); 
      } else {
        errorMessage = result.message || "Failed to resend verification code.";
        formState.update(state => ({ ...state, error: errorMessage }));
        if (appConfig.ui.showErrorToasts) toastStore.showToast(`Resend Code Error: ${errorMessage}`);
      }
    } catch (err) {
      formState.update(state => ({ ...state, loading: false }));
      console.error("Resend Code Exception:", err);
      toastStore.showToast('Failed to resend verification code. Please try again.', 'error');
    }
  }
  
  function goBack() {
    formState.update(state => ({ ...state, step: 1, error: "" }));
  }
  
  onDestroy(() => {
    cleanupTimers();
  });
</script>

<AuthLayout 
  title={$formState.step === 1 ? "Create your account" : "We sent you a code"}
  showBackButton={$formState.step === 2} 
  onBack={goBack}
>
  {#if $formState.error}
    <div class="bg-red-500 bg-opacity-10 border border-red-500 text-red-500 px-4 py-3 rounded mb-4" data-cy="error-message">
      {$formState.error}
    </div>
  {/if}
  
  {#if $formState.step === 1}
    <div data-cy="google-login-button">
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
    </div>
      
    <div class="auth-footer">
      Already have an account? <a href="/login" class="auth-link" data-cy="login-link">Sign in</a>
    </div>
  {:else}
    <p class="auth-subtitle">Enter it below to verify {$formData.email}</p>
    
    <VerificationForm
      bind:verificationCode={$formData.verificationCode}
      showResendOption={$formState.showResendOption}
      timeLeft={formatTimeLeft()}
      onVerify={submitVerification}
      onResend={resendCode}
    />
  {/if}
</AuthLayout>

<DebugPanel />
