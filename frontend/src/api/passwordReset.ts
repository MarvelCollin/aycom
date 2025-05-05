import appConfig from '../config/appConfig';

const API_BASE_URL = appConfig.api.baseUrl;

/**
 * Request security question for password reset
 * @param email User's email
 * @param recaptchaToken reCAPTCHA token for verification
 * @returns Security question and password hash
 */
export async function getSecurityQuestion(email: string, recaptchaToken: string | null) {
  const response = await fetch(`${API_BASE_URL}/auth/forgot-password-question`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, recaptcha_token: recaptchaToken })
  });
  
  const data = await response.json();
  
  if (!response.ok) {
    throw new Error(data.message || 'Invalid email or account is banned.');
  }
  
  return {
    securityQuestion: data.security_question,
    oldPasswordHash: data.old_password_hash
  };
}

/**
 * Verify security question answer
 * @param email User's email
 * @param answer Security question answer
 * @returns Success status
 */
export async function verifySecurityAnswer(email: string, answer: string) {
  const response = await fetch(`${API_BASE_URL}/auth/forgot-password-verify`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, answer })
  });
  
  const data = await response.json();
  
  if (!response.ok) {
    throw new Error(data.message || 'Incorrect answer.');
  }
  
  return { success: true };
}

/**
 * Reset password
 * @param email User's email
 * @param newPassword New password
 * @returns Success status
 */
export async function resetPassword(email: string, newPassword: string) {
  const response = await fetch(`${API_BASE_URL}/auth/forgot-password-reset`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, new_password: newPassword })
  });
  
  const data = await response.json();
  
  if (!response.ok) {
    throw new Error(data.message || 'Failed to reset password.');
  }
  
  return { success: true, message: 'Password reset successful.' };
} 