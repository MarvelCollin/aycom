import { onDestroy } from 'svelte';
import type { GoogleCredentialResponse, CustomWindow } from '../interfaces/auth';

export function useExternalServices() {
  let recaptchaToken = "";
  let recaptchaLoaded = false;
  let googleAuthLoaded = false;
  
  // Direct implementation with hardcoded values as requested
  const RECAPTCHA_SITE_KEY = '6Lfj6CYrAAAAAHmB9dsCaAzKt06ebbXkBduL-bxe';
  const GOOGLE_CLIENT_ID = '161144128362-3jdhmpm3kfr253crkmv23jfqa9ubs2o8.apps.googleusercontent.com';
  const GOOGLE_REDIRECT_URI = 'http://localhost:3000/google-callback';
  
  const loadRecaptcha = (callback: (token: string) => void) => {
    // For Cypress tests, skip actual loading
    if (typeof window !== 'undefined' && (window as any).Cypress) {
      console.log('Mock reCAPTCHA for Cypress');
      recaptchaLoaded = true;
      recaptchaToken = "test-recaptcha-token";
      callback(recaptchaToken);
      return () => {};
    }
    
    console.log('Loading reCAPTCHA with site key:', RECAPTCHA_SITE_KEY);
    
    // Load reCAPTCHA API with invisible approach
    const recaptchaScript = document.createElement('script');
    recaptchaScript.src = `https://www.google.com/recaptcha/api.js?render=${RECAPTCHA_SITE_KEY}`;
    recaptchaScript.async = true;
    document.head.appendChild(recaptchaScript);
    
    recaptchaScript.onload = () => {
      console.log('reCAPTCHA script loaded');
      const customWindow = window as CustomWindow;
      
      if (customWindow.grecaptcha) {
        customWindow.grecaptcha.ready(() => {
          console.log('reCAPTCHA is ready');
          recaptchaLoaded = true;
          
          // Generate token
          if (customWindow.grecaptcha && typeof customWindow.grecaptcha.execute === 'function') {
            customWindow.grecaptcha
              .execute(RECAPTCHA_SITE_KEY, { action: 'register' })
              .then((token: string) => {
                console.log('reCAPTCHA token generated');
                recaptchaToken = token;
                callback(token);
              })
              .catch((error: any) => {
                console.error('reCAPTCHA error:', error);
                callback("");
              });
          }
        });
      }
    };
    
    return () => {
      if (document.head.contains(recaptchaScript)) {
        document.head.removeChild(recaptchaScript);
      }
    };
  };
  
  const refreshRecaptchaToken = (callback: (token: string) => void) => {
    if (typeof window !== 'undefined' && (window as any).Cypress) {
      callback("test-recaptcha-token");
      return;
    }
    
    const customWindow = window as CustomWindow;
    
    if (customWindow.grecaptcha && recaptchaLoaded) {
      customWindow.grecaptcha
        .execute(RECAPTCHA_SITE_KEY, { action: 'register' })
        .then((token: string) => {
          recaptchaToken = token;
          callback(token);
        })
        .catch((error: any) => {
          console.error('Failed to refresh reCAPTCHA token:', error);
          callback("");
        });
    } else {
      console.error('Cannot refresh reCAPTCHA token: not loaded');
      callback("");
    }
  };
  
  const loadGoogleAuth = (
    buttonId: string, 
    isDarkMode: boolean, 
    handleCredentialResponse: (response: GoogleCredentialResponse) => void
  ) => {
    // For Cypress tests, create a mock button
    if (typeof window !== 'undefined' && (window as any).Cypress) {
      console.log('Mock Google Auth for Cypress');
      googleAuthLoaded = true;
      
      setTimeout(() => {
        const buttonElement = document.getElementById(buttonId);
        if (buttonElement) {
          buttonElement.innerHTML = '<button class="google-button">Sign in with Google</button>';
          buttonElement.onclick = () => {
            handleCredentialResponse({ credential: 'test-google-credential' });
          };
        }
      }, 100);
      
      (window as any).handleGoogleCredentialResponse = handleCredentialResponse;
      return () => {};
    }
    
    // Expose the handler to window
    const customWindow = window as CustomWindow;
    customWindow.handleGoogleCredentialResponse = handleCredentialResponse;
    
    // Load Google Sign-In API
    const googleScript = document.createElement('script');
    googleScript.src = "https://accounts.google.com/gsi/client";
    googleScript.async = true;
    googleScript.defer = true;
    document.head.appendChild(googleScript);
    
    googleScript.onload = () => {
      console.log('Google Sign-In script loaded');
      googleAuthLoaded = true;
      
      if (customWindow.google && customWindow.google.accounts) {
        console.log('Initializing Google Sign-In with:', GOOGLE_CLIENT_ID, GOOGLE_REDIRECT_URI);
        
        try {
          customWindow.google.accounts.id.initialize({
            client_id: GOOGLE_CLIENT_ID,
            callback: handleCredentialResponse,
            ux_mode: 'redirect',
            redirect_uri: GOOGLE_REDIRECT_URI
          });
          
          const buttonElement = document.getElementById(buttonId);
          if (buttonElement) {
            // Try to render the official button
            try {
              customWindow.google.accounts.id.renderButton(
                buttonElement, 
                { 
                  theme: isDarkMode ? 'filled_black' : 'outline',
                  size: 'large',
                  width: 300,
                  text: 'continue_with'
                }
              );
            } catch (btnError) {
              console.warn('Could not render official Google button, using custom fallback:', btnError);
              
              // Create a manual Google button as fallback
              buttonElement.innerHTML = `
                <button class="google-signin-button" style="
                  display: flex;
                  align-items: center;
                  justify-content: center;
                  background-color: ${isDarkMode ? '#000' : '#fff'};
                  color: ${isDarkMode ? '#fff' : '#000'};
                  border: 1px solid #dadce0;
                  border-radius: 4px;
                  font-family: 'Google Sans', Roboto, Arial, sans-serif;
                  font-size: 14px;
                  font-weight: 500;
                  height: 40px;
                  padding: 0 12px;
                  text-align: center;
                  width: 100%;
                  cursor: pointer;">
                  <svg width="18" height="18" xmlns="http://www.w3.org/2000/svg" style="margin-right: 8px">
                    <g fill="#000" fill-rule="evenodd">
                      <path d="M9 3.48c1.69 0 2.83.73 3.48 1.34l2.54-2.48C13.46.89 11.43 0 9 0 5.48 0 2.44 2.02.96 4.96l2.91 2.26C4.6 5.05 6.62 3.48 9 3.48z" fill="#EA4335"></path>
                      <path d="M17.64 9.2c0-.74-.06-1.28-.19-1.84H9v3.34h4.96c-.1.83-.64 2.08-1.84 2.92l2.84 2.2c1.7-1.57 2.68-3.88 2.68-6.62z" fill="#4285F4"></path>
                      <path d="M3.88 10.78A5.54 5.54 0 0 1 3.58 9c0-.62.11-1.22.29-1.78L.96 4.96A9.008 9.008 0 0 0 0 9c0 1.45.35 2.82.96 4.04l2.92-2.26z" fill="#FBBC05"></path>
                      <path d="M9 18c2.43 0 4.47-.8 5.96-2.18l-2.84-2.2c-.76.53-1.78.9-3.12.9-2.38 0-4.4-1.57-5.12-3.74L.97 13.04C2.45 15.98 5.48 18 9 18z" fill="#34A853"></path>
                      <path fill="none" d="M0 0h18v18H0z"></path>
                    </g>
                  </svg>
                  Continue with Google
                </button>
              `;
              
              // Add click handler that redirects to Google's OAuth URL
              buttonElement.querySelector('.google-signin-button')?.addEventListener('click', () => {
                const authUrl = `https://accounts.google.com/o/oauth2/v2/auth?client_id=${GOOGLE_CLIENT_ID}&redirect_uri=${encodeURIComponent(GOOGLE_REDIRECT_URI)}&response_type=token&scope=email%20profile`;
                window.location.href = authUrl;
              });
            }
          }
        } catch (error) {
          console.error('Error initializing Google Sign-In:', error);
          
          // Show fallback button on error
          const buttonElement = document.getElementById(buttonId);
          if (buttonElement) {
            buttonElement.innerHTML = `<div class="google-button-error">Google Sign-In Error</div>`;
            buttonElement.style.border = '1px solid #e0e0e0';
            buttonElement.style.padding = '10px';
            buttonElement.style.textAlign = 'center';
          }
        }
      }
    };
    
    return () => {
      if (document.head.contains(googleScript)) {
        document.head.removeChild(googleScript);
      }
    };
  };
  
  const getRecaptchaToken = () => {
    if (typeof window !== 'undefined' && (window as any).Cypress) {
      return "test-recaptcha-token";
    }
    return recaptchaToken;
  };
  
  return {
    loadRecaptcha,
    loadGoogleAuth,
    getRecaptchaToken,
    refreshRecaptchaToken,
    isRecaptchaLoaded: () => recaptchaLoaded,
    isGoogleAuthLoaded: () => googleAuthLoaded
  };
} 