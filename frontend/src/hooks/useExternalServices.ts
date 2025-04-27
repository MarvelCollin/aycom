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
              'sitekey': import.meta.env.GOOGLE_SECRET || import.meta.env.VITE_RECAPTCHA_SITE_KEY,
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
          customWindow.google.accounts.id.initialize({
            client_id: import.meta.env.GOOGLE_CLIENT || import.meta.env.VITE_GOOGLE_CLIENT_ID,
            callback: handleCredentialResponse
          });
          
          const buttonElement = document.getElementById(buttonId);
          if (buttonElement) {
            customWindow.google.accounts.id.renderButton(
              buttonElement, 
              { 
                theme: isDarkMode ? 'filled_black' : 'outline',
                size: 'large',
                width: '100%',
                text: 'signup_with'
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