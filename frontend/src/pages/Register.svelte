<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { useTheme } from '../hooks/useTheme';
  import AuthLayout from '../components/layout/AuthLayout.svelte';
  import RegistrationForm from '../components/auth/RegistrationForm.svelte';
  import VerificationForm from '../components/auth/VerificationForm.svelte';
  import ProfileCompletion from '../components/auth/ProfileCompletion.svelte';
  import { useRegistrationForm } from '../hooks/useRegistrationForm';
  import { useAuth } from '../hooks/useAuth';
  import type { IUserRegistration } from '../interfaces/IAuth';
  import { toastStore } from '../stores/toastStore';
  import { createLoggerWithPrefix } from '../utils/logger';
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
    setFieldError,
    setServerErrors,
    clearErrors,
    startTimer,
    formatTimeLeft,
    cleanupTimers
  } = useRegistrationForm();
  
  const { register, verifyEmail, resendVerificationCode, registerWithMedia } = useAuth();
  
  const { theme } = useTheme();
  
  $: isDarkMode = $theme === 'dark';
  
  // New state for Google auth profile completion
  let showProfileCompletion = false;
  let missingProfileFields: string[] = [];
  
  function validateNameAndUpdate() {
    // Removed client-side validation
  }
  
  function validateUsernameAndUpdate() {
    // Removed client-side validation
  }
  
  function validateEmailAndUpdate() {
    // Removed client-side validation
  }
  
  function validatePasswordAndUpdate() {
    // Removed client-side validation
  }
  
  function validateConfirmPasswordAndUpdate() {
    // Removed client-side validation
  }
  
  function validateGenderAndUpdate() {
    // Removed client-side validation
  }
  
  function validateDateOfBirthAndUpdate() {
    // Removed client-side validation
  }
  
  function validateSecurityQuestionAndUpdate() {
    // Removed client-side validation
  }
  
  function validateSecurityAnswerAndUpdate() {
    // Removed client-side validation
  }
  
  let recaptchaToken: string | null = null;
  
  async function submitRegistration(token: string | null) {
    recaptchaToken = token;
    
    if (!recaptchaToken && !import.meta.env.DEV) {
      const errorMessage = "Please complete the reCAPTCHA verification";
      formState.update(state => ({ ...state, error: errorMessage }));
      if (appConfig.ui.showErrorToasts) toastStore.showToast(errorMessage);
      return;
    }
    
    // Clear any previous errors before submission
    clearErrors();
    
    formState.update(state => ({ ...state, loading: true }));
    
    // Format date of birth in the expected backend format: month_index-day-year
    // Where month_index starts from 0 (January = 0, February = 1, etc.)
    const monthIndex = months.indexOf($formData.dateOfBirth.month);
    const formattedDateOfBirth = `${monthIndex}-${$formData.dateOfBirth.day}-${$formData.dateOfBirth.year}`;
    
    const userData: IUserRegistration = {
      name: $formData.name,
      username: $formData.username,
      email: $formData.email,
      password: $formData.password,
      confirm_password: $formData.confirmPassword,
      gender: $formData.gender,
      date_of_birth: formattedDateOfBirth,
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
        // Handle validation errors from the backend
        console.log("Registration failed with response:", result);
        
        // Check if we have structured validation errors from the backend
        if (result.validation_errors) {
          // Set field-specific errors
          Object.entries(result.validation_errors as Record<string, string | string[]>).forEach(([field, message]) => {
            if (field === 'password') {
              setFieldError(field, Array.isArray(message) ? message : [message.toString()]);
            } else {
              setFieldError(field, message.toString());
            }
          });
        }
        // Check if we have an error object with validation errors
        else if (result.error && result.error.message) {
          // Parse the validation error message
          const errorMessage = result.error.message;
          formState.update(state => ({ ...state, error: errorMessage }));
          
          // Log the full error to help debugging
          console.error("Backend validation error:", errorMessage);
          
          // Try to extract field-specific errors
          if (errorMessage.includes("Key:") || errorMessage.includes("Validation failed:")) {
            let errorPairs: {field: string; message: string}[] = [];
            
            // Check for Key:... Error:... format
            const keyErrorRegex = /Key: '(\w+)' Error:([^,]+)/g;
            let match;
            while ((match = keyErrorRegex.exec(errorMessage)) !== null) {
              errorPairs.push({ field: match[1], message: match[2].trim() });
            }
            
            // Check for field: message format in "Validation failed: ..." messages
            if (errorMessage.includes("Validation failed:")) {
              const validationErrorsText = errorMessage.split("Validation failed:")[1].trim();
              const validationErrors = validationErrorsText.split(";").map(err => err.trim()).filter(Boolean);
              
              for (const validationError of validationErrors) {
                // Try to determine which field this error belongs to
                if (validationError.toLowerCase().includes("name must")) {
                  errorPairs.push({ field: "Name", message: validationError });
                } else if (validationError.toLowerCase().includes("username")) {
                  errorPairs.push({ field: "Username", message: validationError });
                } else if (validationError.toLowerCase().includes("email")) {
                  errorPairs.push({ field: "Email", message: validationError });
                } else if (validationError.toLowerCase().includes("password") && !validationError.toLowerCase().includes("confirm")) {
                  errorPairs.push({ field: "Password", message: validationError });
                } else if (validationError.toLowerCase().includes("match") || 
                           validationError.toLowerCase().includes("confirm")) {
                  errorPairs.push({ field: "ConfirmPassword", message: validationError });
                } else if (validationError.toLowerCase().includes("gender")) {
                  errorPairs.push({ field: "Gender", message: validationError });
                } else if (validationError.toLowerCase().includes("birth") || 
                           validationError.toLowerCase().includes("age") ||
                           validationError.toLowerCase().includes("13 year")) {
                  errorPairs.push({ field: "DateOfBirth", message: validationError });
                } else if (validationError.toLowerCase().includes("security question")) {
                  errorPairs.push({ field: "SecurityQuestion", message: validationError });
                } else if (validationError.toLowerCase().includes("security answer")) {
                  errorPairs.push({ field: "SecurityAnswer", message: validationError });
                }
              }
            }
            
            // Convert backend field names to frontend field names and set errors
            for (const { field, message } of errorPairs) {
              const fieldMapping: Record<string, string> = {
                'Name': 'name',
                'Username': 'username',
                'Email': 'email',
                'Password': 'password',
                'ConfirmPassword': 'confirmPassword',
                'Gender': 'gender',
                'DateOfBirth': 'dateOfBirth',
                'SecurityQuestion': 'securityQuestion',
                'SecurityAnswer': 'securityAnswer'
              };
              
              const frontendField = fieldMapping[field] || field.toLowerCase();
              
              // Set the error for the specific field
              if (frontendField === 'password') {
                setFieldError(frontendField, [message]);
              } else {
                setFieldError(frontendField, message);
              }
              
              // Also update the general error message
              formState.update(state => {
                if (!state.error) {
                  return { ...state, error: "Please fix the validation errors" };
                }
                return state;
              });
            }
          }
        }
        
        const errorMessage = result.message || "Registration failed. Please check the form for errors.";
        formState.update(state => ({ ...state, error: errorMessage }));
        if (appConfig.ui.showErrorToasts) {
          // Fix here: Don't use the same success message text in the error toast
          // Make sure the error message doesn't contain the success message text
          const toastErrorMsg = errorMessage.includes("Registration successful") ? 
            "Registration failed. Please check the form for errors." : errorMessage;
          toastStore.showToast(`Registration Error: ${toastErrorMsg}`, 'error');
        }
      }
    } catch (err) {
      formState.update(state => ({ ...state, loading: false }));
      console.error("Registration Exception:", err);
      toastStore.showToast('Registration failed. Please try again.', 'error');
    }
  }

  function handleGoogleAuthSuccess(result: any) {
    const logger = createLoggerWithPrefix('GoogleRegister');
    logger.info('Google auth success in Register page with result:', result);
    
    // Check if the user needs to complete their profile
    if (result.missing_fields && result.missing_fields.length > 0) {
      logger.info(`User needs to complete profile information: ${result.missing_fields.join(', ')}`);
      missingProfileFields = result.missing_fields;
      showProfileCompletion = true;
      toastStore.showToast('Please complete your profile information', 'info');
    } else if (result.is_new_user) {
      // Even if no missing fields were detected but it's a new user, show profile completion
      logger.info('New user detected, showing profile completion form');
      missingProfileFields = ['gender', 'date_of_birth', 'security_question', 'security_answer'];
      showProfileCompletion = true;
      toastStore.showToast('Welcome! Please complete your profile information', 'info');
    } else {
      toastStore.showToast('Google registration successful', 'success'); 
      logger.info('Redirecting to feed after successful Google registration');
      window.location.href = '/feed';
    }
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
  
  function handleProfileCompleted() {
    const logger = createLoggerWithPrefix('ProfileCompletion');
    logger.info('Profile completion successful');
    toastStore.showToast('Profile updated successfully', 'success');
    logger.info('Redirecting to feed after profile completion');
    window.location.href = '/feed';
  }
  
  function handleProfileSkipped() {
    const logger = createLoggerWithPrefix('ProfileCompletion');
    logger.info('Profile completion skipped');
    toastStore.showToast('You can complete your profile later in account settings', 'info');
    logger.info('Redirecting to feed after skipping profile completion');
    window.location.href = '/feed';
  }
  
  onDestroy(() => {
    cleanupTimers();
  });
</script>

<AuthLayout 
  title={$formState.step === 1 
    ? (showProfileCompletion ? "Complete Your Profile" : "Create your account") 
    : "We sent you a code"}
  showBackButton={$formState.step === 2 || showProfileCompletion} 
  onBack={() => showProfileCompletion ? showProfileCompletion = false : goBack()}
>
  {#if $formState.error}
    <div class="bg-red-500 bg-opacity-10 border border-red-500 text-red-500 px-4 py-3 rounded mb-4" data-cy="error-message">
      {$formState.error}
    </div>
  {/if}
  
  {#if showProfileCompletion}
    <ProfileCompletion 
      missingFields={missingProfileFields} 
      onComplete={handleProfileCompleted}
      onSkip={handleProfileSkipped}
    />
  {:else if $formState.step === 1}
    <div data-cy="google-login-button" class="aycom-register-form">
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

<style>
  :global(.aycom-register-form) {
    width: 100%;
  }
</style>
