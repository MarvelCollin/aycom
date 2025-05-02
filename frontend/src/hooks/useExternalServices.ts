import { onDestroy } from 'svelte';
import type { IGoogleCredentialResponse, ICustomWindow } from '../interfaces/IAuth';

export function useExternalServices() {
  let recaptchaWidgetId: number | null = null;
  let googleAuthLoaded = false;
  
  const getRecaptchaSiteKey = (): string => 
    import.meta.env.VITE_RECAPTCHA_SITE_KEY || '';
    
  const getGoogleClientId = (): string => 
    import.meta.env.VITE_GOOGLE_CLIENT_ID || '';
    
  const getGoogleRedirectUri = (): string => 
    import.meta.env.VITE_GOOGLE_REDIRECT_URI || '';
  
  const loadRecaptcha = (
    callback: (token: string) => void, 
    containerId: string | HTMLElement = 'recaptcha-container'
  ): (() => void) => {
    if ((window as ICustomWindow).grecaptcha) {
      initializeRecaptcha(callback, containerId);
      return () => {};
    }
    
    (window as any).CaptchaCallback = () => {
      initializeRecaptcha(callback, containerId);
    };
    
    const script = document.createElement('script');
    script.src = 'https://www.google.com/recaptcha/api.js?onload=CaptchaCallback&render=explicit';
    script.async = true;
    script.defer = true;
    
    document.head.appendChild(script);
    
    return () => {
      try {
        document.head.removeChild(script);
      } catch (e) {
        console.error('Error removing reCAPTCHA script:', e);
      }
      
      if ((window as ICustomWindow).grecaptcha && recaptchaWidgetId !== null) {
        (window as ICustomWindow).grecaptcha?.reset(recaptchaWidgetId);
      }
      
      delete (window as any).CaptchaCallback;
    };
  };

  const initializeRecaptcha = (
    callback: (token: string) => void, 
    containerId: string | HTMLElement
  ) => {
    const customWindow = window as ICustomWindow;
    if (!customWindow.grecaptcha) return;
    
    try {
      const container = typeof containerId === 'string' 
        ? document.getElementById(containerId) 
        : containerId;
        
      if (!container) {
        console.error(`reCAPTCHA container with ID "${containerId}" not found`);
        return;
      }
      
      recaptchaWidgetId = customWindow.grecaptcha.render(container, {
        sitekey: getRecaptchaSiteKey(),
        theme: 'dark',
        callback: (token: string) => {
          callback(token);
        }
      });
    } catch (error) {
      console.error('Error initializing reCAPTCHA:', error);
    }
  };
  
  const getRecaptchaToken = (): string => {
    const customWindow = window as ICustomWindow;
    if (!customWindow.grecaptcha || recaptchaWidgetId === null) {
      return '';
    }
    
    return customWindow.grecaptcha.getResponse(recaptchaWidgetId);
  };
  
  const loadGoogleAuth = (
    buttonId: string,
    isDarkMode: boolean,
    callback: (response: IGoogleCredentialResponse) => void
  ): (() => void) => {
    (window as ICustomWindow).handleGoogleCredentialResponse = callback;
    
    if ((window as ICustomWindow).google?.accounts) {
      initializeGoogleAuth(buttonId, isDarkMode);
      return () => {};
    }
    
    const script = document.createElement('script');
    script.src = 'https://accounts.google.com/gsi/client';
    script.async = true;
    script.defer = true;
    
    script.onload = () => {
      googleAuthLoaded = true;
      initializeGoogleAuth(buttonId, isDarkMode);
    };
    
    document.head.appendChild(script);
    
    return () => {
      try {
        document.head.removeChild(script);
      } catch (e) {
        console.error('Error removing Google Sign-In script:', e);
      }
      
      delete (window as ICustomWindow).handleGoogleCredentialResponse;
    };
  };
  
  const initializeGoogleAuth = (buttonId: string, isDarkMode: boolean) => {
    const customWindow = window as ICustomWindow;
    if (!customWindow.google?.accounts) return;
    
    try {
      const buttonElement = document.getElementById(buttonId);
      if (!buttonElement) {
        console.error(`Element with ID "${buttonId}" not found`);
        return;
      }
      
      const clientId = getGoogleClientId();
      if (!clientId) {
        console.error('Google Client ID not provided');
        return;
      }
      
      customWindow.google.accounts.id.initialize({
        client_id: clientId,
        callback: customWindow.handleGoogleCredentialResponse,
        auto_select: false,
        cancel_on_tap_outside: true
      });
      
      customWindow.google.accounts.id.renderButton(buttonElement, {
        theme: isDarkMode ? 'filled_black' : 'outline',
        size: 'large',
        type: 'standard',
        shape: 'pill',
        text: 'continue_with',
        logo_alignment: 'left',
        width: buttonElement.offsetWidth
      });
    } catch (error) {
      console.error('Error initializing Google Sign-In:', error);
    }
  };
  
  return {
    loadRecaptcha,
    getRecaptchaToken,
    loadGoogleAuth,
    isGoogleAuthLoaded: () => googleAuthLoaded
  };
} 