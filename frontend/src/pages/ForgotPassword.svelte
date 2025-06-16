<script lang="ts">
import { onMount } from "svelte";
import { useTheme } from "../hooks/useTheme";
import AuthLayout from "../components/layout/AuthLayout.svelte";
import ReCaptchaWrapper from "../components/auth/ReCaptchaWrapper.svelte";
import { toastStore } from "../stores/toastStore";
import { getAuthToken } from "../utils/auth";
import { getSecurityQuestion, verifySecurityAnswer, resetPassword } from "../api/passwordReset";
import { handleApiError } from "../utils/common";
import { createLoggerWithPrefix } from "../utils/logger";

const logger = createLoggerWithPrefix("ForgotPassword");

let step = 1;
let email = "";
let emailError = "";
let securityQuestion = "";
let securityAnswer = "";
let answerError = "";
let newPassword = "";
let newPasswordError = "";
let oldPasswordHash = "";
let resetToken = "";
let recaptchaToken: string | null = null;
let recaptchaWrapper: ReCaptchaWrapper;
let isLoading = false;
const { theme } = useTheme();
$: isDarkMode = $theme === "dark";

onMount(() => {
  if (getAuthToken()) {
    window.location.href = "/feed";
  }
  logger.info("ForgotPassword component mounted, initial step:", step);
});

async function handleEmailSubmit() {
  emailError = "";
  if (!email) {
    emailError = "Email is required";
    return;
  }
  isLoading = true;
  try {
    logger.info("Submitting email for password reset:", email);
    const result = await getSecurityQuestion(email, recaptchaToken);
    logger.info("Received security question result:", result);
    securityQuestion = result.securityQuestion;
    oldPasswordHash = result.oldPasswordHash;
    email = result.email; // Update email from response in case it was normalized
    logger.debug("Security question set to:", securityQuestion);
    step = 2;
  } catch (error) {
    const errorResponse = handleApiError(error);
    logger.error("Error getting security question:", errorResponse);
    emailError = errorResponse.message;
  } finally {
    isLoading = false;
  }
}

async function handleAnswerSubmit() {
  answerError = "";
  if (!securityAnswer) {
    answerError = "Answer is required";
    return;
  }
  isLoading = true;
  try {
    logger.info("Submitting security answer");
    const result = await verifySecurityAnswer(email, securityAnswer);
    logger.info("Security answer verification result:", result);
    resetToken = result.token;
    step = 3;
  } catch (error) {
    const errorResponse = handleApiError(error);
    logger.error("Error verifying security answer:", errorResponse);
    answerError = errorResponse.message;
  } finally {
    isLoading = false;
  }
}

async function handlePasswordSubmit() {
  newPasswordError = "";
  if (!newPassword) {
    newPasswordError = "Password is required";
    return;
  }
  if (newPassword === oldPasswordHash) {
    newPasswordError = "New password cannot be the same as the old password.";
    return;
  }
  isLoading = true;
  try {
    logger.info("Submitting new password");
    const result = await resetPassword(email, newPassword, resetToken);
    logger.debug("Reset password result:", result);
    toastStore.showToast(result.message || "Password reset successful. Please login.", "success");
    window.location.href = "/login";
  } catch (error) {
    const errorResponse = handleApiError(error);
    logger.error("Error resetting password:", errorResponse);
    newPasswordError = errorResponse.message;
  } finally {
    isLoading = false;
  }
}

function handleRecaptchaSuccess(event: CustomEvent<{ token: string }>) {
  recaptchaToken = event.detail.token;
  logger.debug("Recaptcha token received");
}
function handleRecaptchaError() {
  recaptchaToken = null;
  logger.warn("Recaptcha error");
}
function handleRecaptchaExpired() {
  recaptchaToken = null;
  logger.warn("Recaptcha expired");
}
</script>

<AuthLayout title="Recover your password">
  {#if step === 1}
    <form on:submit|preventDefault={handleEmailSubmit} class="mb-4">
      <div class="auth-input-group">
        <label for="email" class="auth-label">Email</label>
        <input
          type="email"
          id="email"
          bind:value={email}
          class="auth-input {isDarkMode ? "auth-input-dark" : ""} {emailError ? "auth-input-error" : ""}"
          placeholder="Enter your account email"
          required
        />
        {#if emailError}
          <p class="auth-error-message">{emailError}</p>
        {/if}
      </div>

      <div class="auth-input-group">
        <ReCaptchaWrapper
          bind:this={recaptchaWrapper}
          siteKey="6Ld6UysrAAAAAPW3XRLe-M9bGDgOPJ2kml1yCozA"
          theme={isDarkMode ? "dark" : "light"}
          on:success={handleRecaptchaSuccess}
          on:error={handleRecaptchaError}
          on:expired={handleRecaptchaExpired}
        />
      </div>

      <button
        type="submit"
        class="auth-btn"
        disabled={isLoading}
      >
        {isLoading ? "Processing..." : "Continue"}
      </button>
    </form>

    <div class="auth-footer">
      <a href="/login" class="auth-link">Back to login</a>
    </div>  {:else if step === 2}
    <p class="auth-subtitle mb-6">Please answer your security question to reset your password</p>

    <form on:submit|preventDefault={handleAnswerSubmit} class="mb-4">
      <div class="auth-input-group">
        <div class="security-question-container">
          <h3 class="security-question-title">Security Question:</h3>
          <div class="security-question-display {isDarkMode ? "security-question-dark" : ""}">
            {#if securityQuestion}
              {securityQuestion}
            {:else}
              <span class="text-red-500">No security question found. Please contact support.</span>
            {/if}
          </div>
        </div>
      </div>

      <div class="auth-input-group">
        <label for="securityAnswer" class="auth-label">Your Answer</label>
        <input
          type="text"
          id="securityAnswer"
          bind:value={securityAnswer}
          class="auth-input {isDarkMode ? "auth-input-dark" : ""} {answerError ? "auth-input-error" : ""}"
          placeholder="Enter your answer to the security question"
          required
        />
        {#if answerError}
          <p class="auth-error-message">{answerError}</p>
        {/if}
      </div>

      <button
        type="submit"
        class="auth-btn"
        disabled={isLoading}
      >
        {isLoading ? "Verifying..." : "Verify Answer"}
      </button>
    </form>

    <div class="auth-footer">
      <a href="/login" class="auth-link">Back to login</a>
    </div>
  {:else if step === 3}
    <form on:submit|preventDefault={handlePasswordSubmit} class="mb-4">
      <div class="auth-input-group">
        <label for="newPassword" class="auth-label">New Password</label>
        <input
          type="password"
          id="newPassword"
          bind:value={newPassword}
          class="auth-input {isDarkMode ? "auth-input-dark" : ""} {newPasswordError ? "auth-input-error" : ""}"
          placeholder="Enter new password"
          required
        />
        {#if newPasswordError}
          <p class="auth-error-message">{newPasswordError}</p>
        {/if}
      </div>

      <button
        type="submit"
        class="auth-btn"
        disabled={isLoading}
      >
        {isLoading ? "Resetting..." : "Reset Password"}
      </button>
    </form>

    <div class="auth-footer">
      <a href="/login" class="auth-link">Back to login</a>
    </div>
  {/if}
</AuthLayout>

<style>
  .security-question-container {
    margin-bottom: 1.5rem;
  }

  .security-question-title {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--text-primary);
    margin-bottom: 0.5rem;
  }
    .security-question-display {
    background-color: var(--bg-secondary, #f8f9fa);
    border: 2px solid var(--border-color, #e1e5e9);
    border-radius: var(--radius-md, 8px);
    padding: 1rem;
    font-size: 1rem;
    font-weight: 500;
    color: var(--text-primary, #1a1a1a);
    line-height: 1.5;
    min-height: 3rem;
    display: flex;
    align-items: center;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .security-question-dark {
    background-color: var(--dark-bg-secondary, #2d3748);
    border-color: var(--dark-border-color, #4a5568);
    color: var(--dark-text-primary, #f7fafc);
  }

  .auth-subtitle {
    color: var(--text-secondary);
    text-align: center;
    font-size: 0.875rem;
    line-height: 1.5;
  }
</style>