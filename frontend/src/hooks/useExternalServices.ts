import { onDestroy } from 'svelte';
import type { GoogleCredentialResponse, CustomWindow } from '../interfaces/auth';

export function useExternalServices() {
  let recaptchaToken = "";
  let recaptchaLoaded = false;
  let googleAuthLoaded = false;
  
  const loadRecaptcha = (callback: (token: string) => void) => {
    const recaptchaScript = document.createElement('script');
    recaptchaScript.src = "https://www.google.com/recaptcha/api.js?render=explicit";
    recaptchaScript.async = true;
    document.head.appendChild(recaptchaScript);
    
    recaptchaScript.onload = () => {
      const customWindow = window as CustomWindow;
      if (customWindow.grecaptcha) {
        customWindow.grecaptcha.ready(() => {
          recaptchaLoaded = true;
          if (customWindow.grecaptcha) {
            customWindow.grecaptcha.render('recaptcha-container', {
              'sitekey': import.meta.env.VITE_RECAPTCHA_SITE_KEY,
              'callback': (token: string) => {
                recaptchaToken = token;
                callback(token);
              },
              'expired-callback': () => {
                recaptchaToken = "";
                callback("");
              }
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
  
  const loadGoogleAuth = (
    buttonId: string, 
    isDarkMode: boolean, 
    handleCredentialResponse: (response: GoogleCredentialResponse) => void
  ) => {
    const googleScript = document.createElement('script');
    googleScript.src = "https://accounts.google.com/gsi/client";
    googleScript.async = true;
    googleScript.defer = true;
    document.head.appendChild(googleScript);
    
    googleScript.onload = () => {
      googleAuthLoaded = true;
      try {
        const customWindow = window as CustomWindow;
        if (customWindow.google) {
          // Get Google Client ID from environment variables
          const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID || "161144128362-3jdhmpm3kfr253crkmv23jfqa9ubs2o8.apps.googleusercontent.com";
          
          customWindow.google.accounts.id.initialize({
            client_id: clientId,
            callback: handleCredentialResponse,
            ux_mode: 'redirect',
            redirect_uri: 'http://localhost:3000/google-callback'
          });
          
          const buttonElement = document.getElementById(buttonId);
          if (buttonElement) {
            customWindow.google.accounts.id.renderButton(
              buttonElement, 
              { 
                theme: isDarkMode ? 'filled_black' : 'outline',
                size: 'large',
                width: '100%',
                text: 'continue_with'
              }
            );
          }
        }
      } catch (error) {
        console.error('Failed to initialize Google Sign-In:', error);
      }
    };
    
    return () => {
      if (document.head.contains(googleScript)) {
        document.head.removeChild(googleScript);
      }
    };
  };
  
  return {
    loadRecaptcha,
    loadGoogleAuth,
    getRecaptchaToken: () => recaptchaToken,
    isRecaptchaLoaded: () => recaptchaLoaded,
    isGoogleAuthLoaded: () => googleAuthLoaded
  };
} 