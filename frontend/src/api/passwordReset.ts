import appConfig from '../config/appConfig';
import { createLoggerWithPrefix } from '../utils/logger';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('PasswordResetAPI');

export async function getSecurityQuestion(email: string, recaptchaToken: string | null) {
  logger.debug('Fetching security question for email:', email);
  
  const response = await fetch(`${API_BASE_URL}/auth/forgot-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email })
  });

  const responseData = await response.json();
  logger.debug('Received response:', responseData);

  if (!response.ok) {
    throw new Error(responseData.message || 'Invalid email or account is banned.');
  }
  
  // Extract data from the nested structure
  const data = responseData.data || responseData;
  
  return {
    securityQuestion: data.security_question,
    oldPasswordHash: data.old_password_hash || '',
    email: data.email
  };
}

export async function verifySecurityAnswer(email: string, answer: string) {
  const response = await fetch(`${API_BASE_URL}/auth/verify-security-answer`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, security_answer: answer })
  });

  const responseData = await response.json();
  logger.debug('Verify security answer response:', responseData);

  if (!response.ok) {
    throw new Error(responseData.message || 'Incorrect answer.');
  }

  // Extract data from the nested structure
  const data = responseData.data || responseData;

  return { 
    success: data.success || responseData.success, 
    token: data.token || data.reset_token,
    email: data.email,
    expires_at: data.expires_at || data.token_expiration_time
  };
}

export async function resetPassword(email: string, newPassword: string, token: string) {
  logger.debug('Resetting password for email:', email);
  
  try {
    const response = await fetch(`${API_BASE_URL}/auth/reset-password`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ 
        email, 
        new_password: newPassword, 
        token 
      })
    });

    const responseData = await response.json();
    logger.debug('Reset password response:', responseData);

    if (!response.ok) {
      logger.error('Error in resetPassword:', responseData);
      throw new Error(responseData.message || 'Failed to reset password.');
    }

    return {
      success: true,
      message: responseData.message || 'Password has been reset successfully.'
    };
  } catch (error) {
    logger.error('Error in resetPassword:', error);
    throw error;
  }
}