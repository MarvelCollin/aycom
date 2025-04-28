import { onDestroy } from 'svelte';
import type { GoogleCredentialResponse, CustomWindow } from '../interfaces/auth';
import type { IGoogleCredentialResponse } from '../interfaces/IAuth';

export function useExternalServices() {
  let recaptchaWidgetId: number | null = null;
  let googleAuthLoaded = false;
  
  // Direct implementation with hardcoded values as requested
  const RECAPTCHA_SITE_KEY = '6Lfj6CYrAAAAAHmB9dsCaAzKt06ebbXkBduL-bxe';
  const GOOGLE_CLIENT_ID = '161144128362-3jdhmpm3kfr253crkmv23jfqa9ubs2o8.apps.googleusercontent.com';
  const GOOGLE_REDIRECT_URI = 'http://localhost:3000/register';
  
  /**
   * Loads the reCAPTCHA script and initializes it
   * @param callback Function to call when the token is updated
   * @returns Cleanup function
   */
  const loadRecaptcha = (callback: (token: string) => void): (() => void) => {
    // Check if reCAPTCHA script is already loaded
    if (window.grecaptcha) {
      initializeRecaptcha(callback);
      return () => {};
    }
    
    // Add the reCAPTCHA script
    const script = document.createElement('script');
    script.src = 'https://www.google.com/recaptcha/api.js?render=explicit';
    script.async = true;
    script.defer = true;
    
    script.onload = () => {
      initializeRecaptcha(callback);
    };
    
    document.head.appendChild(script);
    
    // Return cleanup function
    return () => {
      // Remove the script
      try {
        document.head.removeChild(script);
      } catch (e) {
        console.error('Error removing reCAPTCHA script:', e);
      }
      
      // Reset reCAPTCHA if it exists
      if (window.grecaptcha && recaptchaWidgetId !== null) {
        window.grecaptcha.reset(recaptchaWidgetId);
      }
    };
  };
  
  /**
   * Initializes reCAPTCHA after the script is loaded
   * @param callback Function to call when the token is updated
   */
  const initializeRecaptcha = (callback: (token: string) => void) => {
    if (!window.grecaptcha) return;
    
    window.grecaptcha.ready(() => {
      recaptchaWidgetId = window.grecaptcha.render('recaptcha-container', {
        sitekey: import.meta.env.VITE_RECAPTCHA_SITE_KEY || '6LeIxAcTAAAAAJcZVRqyHh71UMIEGNQ_MXjiZKhI', // Test key if not provided
        theme: 'dark',
        callback: (token: string) => {
          callback(token);
        }
      });
    });
  };
  
  /**
   * Gets the current reCAPTCHA token
   * @returns The reCAPTCHA token
   */
  const getRecaptchaToken = (): string => {
    if (!window.grecaptcha || recaptchaWidgetId === null) {
      return '';
    }
    
    return window.grecaptcha.getResponse(recaptchaWidgetId);
  };
  
  /**
   * Loads the Google Sign-In script and initializes it
   * @param buttonId ID of the HTML element to render the Google button in
   * @param isDarkMode Whether to use dark mode
   * @param callback Function to call when a credential is received
   * @returns Cleanup function
   */
  const loadGoogleAuth = (
    buttonId: string,
    isDarkMode: boolean,
    callback: (response: IGoogleCredentialResponse) => void
  ): (() => void) => {
    // Set up the global callback function for Google
    window.handleGoogleCredentialResponse = callback;
    
    // Check if Google Sign-In script is already loaded
    if (window.google?.accounts) {
      initializeGoogleAuth(buttonId, isDarkMode);
      return () => {};
    }
    
    // Add the Google Sign-In script
    const script = document.createElement('script');
    script.src = 'https://accounts.google.com/gsi/client';
    script.async = true;
    script.defer = true;
    
    script.onload = () => {
      initializeGoogleAuth(buttonId, isDarkMode);
    };
    
    document.head.appendChild(script);
    
    // Return cleanup function
    return () => {
      // Remove the script
      try {
        document.head.removeChild(script);
      } catch (e) {
        console.error('Error removing Google Sign-In script:', e);
      }
      
      // Remove the global callback
      delete window.handleGoogleCredentialResponse;
    };
  };
  
  /**
   * Initializes Google Sign-In after the script is loaded
   * @param buttonId ID of the HTML element to render the Google button in
   * @param isDarkMode Whether to use dark mode
   */
  const initializeGoogleAuth = (buttonId: string, isDarkMode: boolean) => {
    if (!window.google?.accounts) return;
    
    const buttonElement = document.getElementById(buttonId);
    if (!buttonElement) {
      console.error(`Element with ID "${buttonId}" not found`);
      return;
    }
    
    window.google.accounts.id.initialize({
      client_id: import.meta.env.VITE_GOOGLE_CLIENT_ID || 'YOUR_CLIENT_ID_HERE',
      callback: window.handleGoogleCredentialResponse,
      auto_select: false,
      cancel_on_tap_outside: true
    });
    
    window.google.accounts.id.renderButton(buttonElement, {
      theme: isDarkMode ? 'filled_black' : 'outline',
      size: 'large',
      type: 'standard',
      shape: 'pill',
      text: 'continue_with',
      logo_alignment: 'left',
      width: buttonElement.offsetWidth
    });
  };
  
  return {
    loadRecaptcha,
    getRecaptchaToken,
    loadGoogleAuth,
    isGoogleAuthLoaded: () => googleAuthLoaded
  };
} 