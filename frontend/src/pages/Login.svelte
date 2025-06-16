<script lang="ts">
  import { useAuth } from "../hooks/useAuth";
  import { useTheme } from "../hooks/useTheme";
  import AuthLayout from "../components/layout/AuthLayout.svelte";
  import GoogleSignInButton from "../components/auth/GoogleSignInButton.svelte";
  import { toastStore } from "../stores/toastStore";
  import appConfig from "../config/appConfig";
  import ReCaptchaWrapper from "../components/auth/ReCaptchaWrapper.svelte";
  import { getAuthToken, clearAuthData } from "../utils/auth";
  import { createLoggerWithPrefix } from "../utils/logger";
  import DebugPanel from "../components/common/DebugPanel.svelte";
  import { onMount } from "svelte";
  import Toast from "../components/common/Toast.svelte";
  import ProfileCompletion from "../components/auth/ProfileCompletion.svelte";

  const logger = createLoggerWithPrefix("Login");
  const { login } = useAuth();

  const { theme } = useTheme();

  $: isDarkMode = $theme === "dark";

  let email = "";
  let password = "";
  let rememberMe = false;
  let error = "";
  let isLoading = false;
  let recaptchaToken: string | null = null;
  let recaptchaWrapper: ReCaptchaWrapper;

  let showProfileCompletion = false;
  let missingProfileFields: string[] = [];

  function handleRecaptchaSuccess(event: CustomEvent<{ token: string }>) {
    recaptchaToken = event.detail.token;
    logger.info(`reCAPTCHA token received: ${recaptchaToken?.substring(0, 10)}...`);
  }

  function handleRecaptchaError(event: CustomEvent<{ message: string }>) {
    recaptchaToken = null;
    const message = event.detail?.message || "Unknown error";
    logger.error(`reCAPTCHA error: ${message}`);

    if (!error) {
      error = `reCAPTCHA verification failed: ${message}`;
      if (appConfig.ui.showErrorToasts) toastStore.showToast(error, "error");
    }
  }

  function handleRecaptchaExpired() {
    recaptchaToken = null;
    logger.warn("reCAPTCHA token expired, please verify again");
    if (appConfig.ui.showErrorToasts) toastStore.showToast("reCAPTCHA verification expired, please try again", "warning");
  }

  onMount(() => {
    const token = getAuthToken();
    logger.info(`Login page - Auth token exists: ${!!token}, Current path: ${window.location.pathname}`);

    clearAuthData();
  });

  async function handleSubmit() {
    if (isLoading) return;

    error = "";
    isLoading = true;

    try {

      if (!email || !password) {
        isLoading = false;
        error = "Please enter both email and password";
        return;
      }

      if (!recaptchaToken && !import.meta.env.DEV) {
        try {
          if (recaptchaWrapper) {
            recaptchaToken = await recaptchaWrapper.execute();
          }
        } catch (recaptchaError) {
          logger.error("reCAPTCHA error:", recaptchaError);

        }
      }

      const result = await login(email, password, recaptchaToken);
      isLoading = false;

      if (result.success) {
        logger.info("Login successful");
        toastStore.showToast("Login successful", "success");

        setTimeout(() => {
          const currentPath = window.location.pathname;
          if (currentPath !== "/feed") {
            logger.info("Login successful, redirecting to feed");
            window.location.href = "/feed";
          }
        }, 100);
      } else {
        errorMessage = result.message || "Login failed. Please check your credentials.";
        logger.error("Login failed with message:", errorMessage);
        error = errorMessage;
        toastStore.showToast(errorMessage, "error");

        try {
          if (recaptchaWrapper && typeof recaptchaWrapper.reset === "function") {
            recaptchaWrapper.reset();
            recaptchaToken = null;
          }
        } catch (resetError) {
          logger.error("Error resetting reCAPTCHA:", resetError);
        }
      }
    } catch (err) {
      isLoading = false;
      const message = err instanceof Error ? err.message : "Login failed. Please try again.";
      logger.error("Login Exception:", err);
      error = message;
      toastStore.showToast(message, "error");

      try {
        if (recaptchaWrapper && typeof recaptchaWrapper.reset === "function") {
          recaptchaWrapper.reset();
          recaptchaToken = null;
        }
      } catch (resetError) {
        logger.error("Error resetting reCAPTCHA:", resetError);
      }
    }
  }

  interface AuthResult {
    success: boolean;
    message?: string;
    [key: string]: any;
  }

  function handleGoogleAuthSuccess(result: AuthResult) {
    const logger = createLoggerWithPrefix("GoogleLogin");
    logger.info("Google auth success in Login page with result:", result);

    const token = getAuthToken();
    if (!token) {
      logger.warn("No auth token found after successful Google login, forcing refresh");
      window.location.reload();
      return;
    }

    if (result.missing_fields && result.missing_fields.length > 0) {
      logger.info(`User needs to complete profile information: ${result.missing_fields.join(", ")}`);
      missingProfileFields = result.missing_fields;
      showProfileCompletion = true;
      toastStore.showToast("Please complete your profile information", "info");
    } else if (result.is_new_user) {

      logger.info("New user detected, showing profile completion form");
      missingProfileFields = ["gender", "date_of_birth", "security_question", "security_answer"];
      showProfileCompletion = true;
      toastStore.showToast("Welcome! Please complete your profile information", "info");
    } else {
      toastStore.showToast("Google login successful", "success");
      logger.info("Redirecting to feed after successful Google login");
      window.location.href = "/feed";
    }
  }

  function handleGoogleAuthError(message: string) {
    console.error("Google auth error in Login page:", message);
    error = message;
    if (appConfig.ui.showErrorToasts) {
      toastStore.showToast(`Google Auth Error: ${message}`, "error");
    }
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
</script>

<AuthLayout
  title={showProfileCompletion ? "Complete Your Profile" : "Sign in to AYCOM"}
  showBackButton={showProfileCompletion}
  onBack={() => showProfileCompletion = false}
>
  {#if !showProfileCompletion}
    <div class="auth-social-btn-container aycom-login-form" data-cy="google-login-button">
      <GoogleSignInButton
        onAuthSuccess={handleGoogleAuthSuccess}
        onAuthError={handleGoogleAuthError}
        class="auth-social-btn {isDarkMode ? "auth-social-btn-dark" : ""}"
      />
    </div>

    <div class="auth-divider {isDarkMode ? "auth-divider-dark" : ""}">
      <span class="auth-divider-text">or</span>
    </div>

    {#if error}
      <div class="bg-red-500 bg-opacity-10 border border-red-500 text-red-500 px-4 py-3 rounded mb-4" data-cy="error-message">
        {error}
      </div>
    {/if}

    <form on:submit|preventDefault={handleSubmit} class="mb-4">
      <div class="auth-input-group">
        <label for="email" class="auth-label">Email</label>
        <input
          type="email"
          id="email"
          bind:value={email}
          class="auth-input {isDarkMode ? "auth-input-dark" : ""} {error && !email ? "auth-input-error" : ""}"
          placeholder="Email"
          required
          data-cy="email-input"
        />
        {#if error && !email}
          <p class="auth-error-message" data-cy="email-error">Email is required</p>
        {/if}
      </div>

      <div class="auth-input-group">
        <div class="flex justify-between items-center mb-1">
          <label for="password" class="auth-label">Password</label>
          <a href="/forgot-password" class="auth-forgot-password" data-cy="forgot-password">Forgot password?</a>
        </div>
        <input
          type="password"
          id="password"
          bind:value={password}
          class="auth-input {isDarkMode ? "auth-input-dark" : ""} {error && !password ? "auth-input-error" : ""}"
          placeholder="Password"
          required
          data-cy="password-input"
        />
        {#if error && !password}
          <p class="auth-error-message" data-cy="password-error">Password is required</p>
        {/if}
      </div>

      <div class="auth-checkbox-group">
        <input
          type="checkbox"
          id="remember-me"
          bind:checked={rememberMe}
          class="auth-checkbox"
          data-cy="remember-me"
        />
        <label for="remember-me" class="auth-checkbox-label">Remember me</label>
      </div>

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
        type="submit"
        class="auth-btn"
        disabled={isLoading}
        data-cy="login-button"
      >
        {#if isLoading}
          <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white inline-block" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          Signing in...
        {:else}
          Sign in
        {/if}
      </button>
    </form>

    <div class="auth-footer">
      Don't have an account? <a href="/register" class="auth-link" data-cy="register-link">Sign up</a>
    </div>
  {:else}
    <ProfileCompletion
      missingFields={missingProfileFields}
      onComplete={handleProfileCompleted}
      onSkip={handleProfileSkipped}
    />
  {/if}
</AuthLayout>

<DebugPanel />

<Toast />

<style>
  .recaptcha-wrapper {
    margin: 1.5rem 0;
    width: 100%;
  }

  :global(.aycom-login-form) {
    width: 100%;
  }
</style>