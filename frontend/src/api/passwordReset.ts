import appConfig from '../config/appConfig';

const API_BASE_URL = appConfig.api.baseUrl;

export async function getSecurityQuestion(email: string, recaptchaToken: string | null) {
  const response = await fetch(`${API_BASE_URL}/v1/auth/forgot-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email })
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Invalid email or account is banned.');
  }

  return {
    security_question: data.security_question,
    old_password_hash: data.old_password_hash || '',
    email: data.email
  };
}

export async function verifySecurityAnswer(email: string, answer: string) {
  const response = await fetch(`${API_BASE_URL}/v1/auth/verify-security-answer`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, security_answer: answer })
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Incorrect answer.');
  }

  return { 
    success: data.success, 
    token: data.token || data.reset_token,
    email: data.email,
    expires_at: data.expires_at || data.token_expiration_time
  };
}

export async function resetPassword(email: string, newPassword: string, token: string) {
  const response = await fetch(`${API_BASE_URL}/v1/auth/reset-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, new_password: newPassword, token })
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || 'Failed to reset password.');
  }

  return { success: data.success, message: data.message };
}