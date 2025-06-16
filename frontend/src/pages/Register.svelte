<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { useTheme } from "../hooks/useTheme";
  import AuthLayout from "../components/layout/AuthLayout.svelte";
  import RegistrationForm from "../components/auth/RegistrationForm.svelte";
  import VerificationForm from "../components/auth/VerificationForm.svelte";
  import ProfileCompletion from "../components/auth/ProfileCompletion.svelte";
  import { useRegistrationForm } from "../hooks/useRegistrationForm";
  import { useAuth } from "../hooks/useAuth";
  import type { IUserRegistration } from "../interfaces/IAuth";
  import { toastStore } from "../stores/toastStore";
  import { createLoggerWithPrefix } from "../utils/logger";
  import appConfig from "../config/appConfig";
  import DebugPanel from "../components/common/DebugPanel.svelte";

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

  $: isDarkMode = $theme === "dark";

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
      console.log("Registration failed with response:", result);

      if (result.success) {
        startTimer();
        formState.update(state => ({ ...state, step: 2, error: "" }));
        const successEl = document.createElement("div");
        successEl.textContent = "Registration successful";
        successEl.setAttribute("data-cy", "success-message");
        successEl.style.position = "absolute";
        successEl.style.left = "-9999px";
        document.body.appendChild(successEl);
        toastStore.showToast(
          "Registration successful. Please check your email to verify your account.",
          "success"
        );
      } else {
        // Handle validation errors from the API response
        if (result.validation_errors) {
          // Clear previous validation errors
          clearErrors();

          // Set specific field errors directly from API response
          Object.entries(result.validation_errors).forEach(([field, message]) => {
            // Convert field names from snake_case to camelCase
            const camelField = field.replace(/_([a-z])/g, (match, p1) => p1.toUpperCase());

            if (field === "password") {
              setFieldError(camelField, Array.isArray(message) ? message : [String(message)]);
            } else {
              setFieldError(camelField, String(message));
            }
          });

          // Update the form state with general error message
          formState.update(state => ({
            ...state,
            error: "Please fix the validation errors highlighted below.",
            loading: false
          }));
        }
        // Check if we have an error object with validation errors in the message
        else if (result.error?.fields) {
          // Clear previous validation errors
          clearErrors();

          // Set specific field errors from the error fields object
          Object.entries(result.error.fields).forEach(([field, message]) => {
            const camelField = field.replace(/_([a-z])/g, (match, p1) => p1.toUpperCase());

            if (field === "password") {
              setFieldError(camelField, Array.isArray(message) ? message : [String(message)]);
            } else {
              setFieldError(camelField, String(message));
            }
          });

          formState.update(state => ({
            ...state,
            error: "Please fix the validation errors highlighted below.",
            loading: false
          }));
        }
        // Handle cases where we just have a general error message
        else if (result.error && result.error.message) {
          const errorMessage = result.error.message;
          formState.update(state => ({
            ...state,
            error: typeof errorMessage === "string" ? errorMessage : "Registration failed due to validation errors.",
            loading: false
          }));

          // Try to extract field errors from error message string if possible
          if (typeof errorMessage === "string" &&
             (errorMessage.includes("Key:") || errorMessage.includes("validation"))) {

            parseErrorMessageForFieldErrors(errorMessage);
          }
        }

        // Always ensure errorMessage is a string
        const errorMessage = typeof result.message === "string" ? result.message :
          (result.error && typeof result.error.message === "string" ? result.error.message :
            "Registration failed. Please check the form for errors.");

        formState.update(state => ({ ...state, error: errorMessage }));
        if (appConfig.ui.showErrorToasts) {
          // Ensure the errorMessage is a string before checking if it includes something
          const toastErrorMsg = typeof errorMessage === "string" && errorMessage.includes("Registration successful") ?
            "Registration failed. Please check the form for errors." : errorMessage;
          toastStore.showToast(`Registration Error: ${toastErrorMsg}`, "error");
        }
      }
    } catch (err) {
      formState.update(state => ({ ...state, loading: false }));
      console.error("Registration Exception:", err);
      toastStore.showToast("Registration failed. Please try again.", "error");
    }
  }

  function handleGoogleAuthSuccess(result: any) {
    const logger = createLoggerWithPrefix("GoogleRegister");
    logger.info("Google auth success in Register page with result:", result);

    // Check if the user needs to complete their profile
    if (result.missing_fields && result.missing_fields.length > 0) {
      logger.info(`User needs to complete profile information: ${result.missing_fields.join(", ")}`);
      missingProfileFields = result.missing_fields;
      showProfileCompletion = true;
      toastStore.showToast("Please complete your profile information", "info");
    } else if (result.is_new_user) {
      // Even if no missing fields were detected but it's a new user, show profile completion
      logger.info("New user detected, showing profile completion form");
      missingProfileFields = ["gender", "date_of_birth", "security_question", "security_answer"];
      showProfileCompletion = true;
      toastStore.showToast("Welcome! Please complete your profile information", "info");
    } else {
      toastStore.showToast("Google registration successful", "success");
      logger.info("Redirecting to feed after successful Google registration");
      window.location.href = "/feed";
    }
  }

  function handleGoogleAuthError(errorMsg: string) {
    console.error("Google auth error in Register page:", errorMsg);
    formState.update(state => ({ ...state, error: errorMsg }));
    if (appConfig.ui.showErrorToasts) toastStore.showToast(`Google Auth Error: ${errorMsg}`, "error");
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
        window.location.href = "/login";
      } else {
        errorMessage = result.message || "Verification failed. Please check your code and try again.";
        formState.update(state => ({ ...state, error: errorMessage }));
        if (appConfig.ui.showErrorToasts) toastStore.showToast(`Verification Error: ${errorMessage}`);
      }
    } catch (err) {
      formState.update(state => ({ ...state, loading: false }));
      console.error("Verification Exception:", err);
      toastStore.showToast("Verification failed. Please try again.", "error");
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
      toastStore.showToast("Failed to resend verification code. Please try again.", "error");
    }
  }

  function goBack() {
    formState.update(state => ({ ...state, step: 1, error: "" }));
  }

  function handleProfileCompleted() {
    const logger = createLoggerWithPrefix("ProfileCompletion");
    logger.info("Profile completion successful");
    toastStore.showToast("Profile updated successfully", "success");
    logger.info("Redirecting to feed after profile completion");
    window.location.href = "/feed";
  }

  function handleProfileSkipped() {
    const logger = createLoggerWithPrefix("ProfileCompletion");
    logger.info("Profile completion skipped");
    toastStore.showToast("You can complete your profile later in account settings", "info");
    logger.info("Redirecting to feed after skipping profile completion");
    window.location.href = "/feed";
  }

  onDestroy(() => {
    cleanupTimers();
  });

  // Add this function to extract field errors from error message strings
  function parseErrorMessageForFieldErrors(errorMessage) {
    // Common field mappings
    const fieldMappings = {
      "Name": "name",
      "Username": "username",
      "Email": "email",
      "Password": "password",
      "ConfirmPassword": "confirmPassword",
      "Gender": "gender",
      "DateOfBirth": "dateOfBirth",
      "SecurityQuestion": "securityQuestion",
      "SecurityAnswer": "securityAnswer"
    };

    try {
      // Extract field errors from message format: "Key: 'Field' Error:message"
      const errorRegex = /Key:\s*'([^']+)'\s*Error:([^,;]+)/g;
      let match;

      while ((match = errorRegex.exec(errorMessage)) !== null) {
        const fieldName = match[1];
        const errorDesc = match[2].trim();

        // Map the field name to our camelCase version
        const formField = fieldMappings[fieldName] || fieldName.toLowerCase();

        // Set the field error
        if (formField === "password") {
          setFieldError(formField, [errorDesc]);
        } else {
          setFieldError(formField, errorDesc);
        }
      }
    } catch (e) {
      console.error("Error parsing validation message:", e);
    }
  }
</script>

<AuthLayout
  title={$formState.step === 1
    ? (showProfileCompletion ? "Complete Your Profile" : "Create your account")
    : "We sent you a code"}
  showBackButton={$formState.step === 2 || showProfileCompletion}
  onBack={() => showProfileCompletion ? showProfileCompletion = false : goBack()}
>
  {#if $formState.error}
    <div class="bg-red-100 border-l-4 border-red-500 text-red-700 p-4 mb-5 rounded shadow-sm" data-cy="error-message">
      <div class="font-medium text-lg mb-2">Form Validation Errors</div>

      <!-- Display field-specific errors as a list -->
      {#if Object.keys($errors).some(key => !!$errors[key] && (typeof $errors[key] === "string" ? $errors[key].length > 0 : $errors[key].length > 0))}
        <ul class="mt-2 list-disc list-inside text-sm space-y-1">
          {#if $errors.name && $errors.name.length > 0}
            <li><strong>Name:</strong> {$errors.name}</li>
          {/if}
          {#if $errors.username && $errors.username.length > 0}
            <li><strong>Username:</strong> {$errors.username}</li>
          {/if}
          {#if $errors.email && $errors.email.length > 0}
            <li><strong>Email:</strong> {$errors.email}</li>
          {/if}
          {#if $errors.password && $errors.password.length > 0}
            <li>
              <strong>Password:</strong>
              {#if Array.isArray($errors.password) && $errors.password.length > 0}
                <ul class="pl-5 list-disc">
                  {#each $errors.password as error}
                    <li>{error}</li>
                  {/each}
                </ul>
              {:else}
                {$errors.password}
              {/if}
            </li>
          {/if}
          {#if $errors.confirmPassword && $errors.confirmPassword.length > 0}
            <li><strong>Confirm Password:</strong> {$errors.confirmPassword}</li>
          {/if}
          {#if $errors.gender && $errors.gender.length > 0}
            <li><strong>Gender:</strong> {$errors.gender}</li>
          {/if}
          {#if $errors.dateOfBirth && $errors.dateOfBirth.length > 0}
            <li><strong>Date of Birth:</strong> {$errors.dateOfBirth}</li>
          {/if}
          {#if $errors.securityQuestion && $errors.securityQuestion.length > 0}
            <li><strong>Security Question:</strong> {$errors.securityQuestion}</li>
          {/if}
          {#if $errors.securityAnswer && $errors.securityAnswer.length > 0}
            <li><strong>Security Answer:</strong> {$errors.securityAnswer}</li>
          {/if}
          {#if $errors.profilePicture && $errors.profilePicture.length > 0}
            <li><strong>Profile Picture:</strong> {$errors.profilePicture}</li>
          {/if}
          {#if $errors.banner && $errors.banner.length > 0}
            <li><strong>Banner:</strong> {$errors.banner}</li>
          {/if}
        </ul>
        <div class="mt-3 text-sm">Please fix the errors above and try again.</div>
      {:else}
        <div>{$formState.error}</div>
      {/if}
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
        securityAnswerError={$errors.securityAnswer}
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
