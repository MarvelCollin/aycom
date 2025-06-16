import { onDestroy } from "svelte";
import type { IGoogleCredentialResponse, ICustomWindow } from "../interfaces/IAuth";

export function useExternalServices() {
  let recaptchaWidgetId: number | null = null;
  let googleAuthLoaded = false;

  const getRecaptchaSiteKey = (): string =>
    import.meta.env.VITE_RECAPTCHA_SITE_KEY || "6Ld6UysrAAAAAPW3XRLe-M9bGDgOPJ2kml1yCozA";

  const getGoogleClientId = (): string => {
    const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID || "161144128362-3jdhmpm3kfr253crkmv23jfqa9ubs2o8.apps.googleusercontent.com";
    console.log("Using Google Client ID:", clientId);
    return clientId;
  };

  const getGoogleRedirectUri = (): string => {
    const redirectUri = import.meta.env.VITE_GOOGLE_REDIRECT_URI || "http://localhost:3000/register";
    console.log("Using Google Redirect URI:", redirectUri);
    return redirectUri;
  };

  const loadRecaptcha = (
    callback: (token: string) => void,
    containerId: string | HTMLElement = "recaptcha-container"
  ): (() => void) => {
    if ((window as ICustomWindow).grecaptcha) {
      initializeRecaptcha(callback, containerId);
      return () => {};
    }

    (window as any).CaptchaCallback = () => {
      initializeRecaptcha(callback, containerId);
    };

    const script = document.createElement("script");
    script.src = "https://www.google.com/recaptcha/api.js?onload=CaptchaCallback&render=explicit";
    script.async = true;
    script.defer = true;

    document.head.appendChild(script);

    return () => {
      try {
        document.head.removeChild(script);
      } catch (e) {
        console.error("Error removing reCAPTCHA script:", e);
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
      const container = typeof containerId === "string"
        ? document.getElementById(containerId)
        : containerId;

      if (!container) {
        console.error(`reCAPTCHA container with ID "${containerId}" not found`);
        return;
      }

      recaptchaWidgetId = customWindow.grecaptcha.render(container, {
        sitekey: getRecaptchaSiteKey(),
        theme: "dark",
        callback: (token: string) => {
          callback(token);
        }
      });
    } catch (error) {
      console.error("Error initializing reCAPTCHA:", error);
    }
  };

  const getRecaptchaToken = (): string => {
    const customWindow = window as ICustomWindow;
    if (!customWindow.grecaptcha || recaptchaWidgetId === null) {
      return "";
    }

    try {
      return customWindow.grecaptcha.getResponse(recaptchaWidgetId) || "";
    } catch (error) {
      console.error("Error getting reCAPTCHA token:", error);
      return "";
    }
  };

  const loadGoogleAuth = (
    buttonId: string,
    isDarkMode: boolean,
    callback: (response: IGoogleCredentialResponse) => void
  ): (() => void) => {
    // Store the callback in the window object
    (window as ICustomWindow).handleGoogleCredentialResponse = (response) => {
      console.log("Google credential response received:", response);
      callback(response);
    };

    // Check if Google API is already loaded
    if ((window as ICustomWindow).google?.accounts) {
      console.log("Google accounts API already loaded, initializing...");
      initializeGoogleAuth(buttonId, isDarkMode);
      return () => {
        delete (window as ICustomWindow).handleGoogleCredentialResponse;
      };
    }

    console.log("Loading Google accounts API...");
    const script = document.createElement("script");
    script.src = "https://accounts.google.com/gsi/client";
    script.async = true;
    script.defer = true;

    script.onload = () => {
      console.log("Google accounts API loaded successfully");
      googleAuthLoaded = true;
      initializeGoogleAuth(buttonId, isDarkMode);
    };

    script.onerror = (error) => {
      console.error("Failed to load Google accounts API:", error);
      const buttonElement = document.getElementById(buttonId);
      if (buttonElement) {
        buttonElement.innerHTML = "<div class=\"p-2 text-center text-red-500\">Failed to load Google Sign-In</div>";
      }
    };

    document.head.appendChild(script);

    return () => {
      try {
        document.head.removeChild(script);
      } catch (e) {
        console.error("Error removing Google Sign-In script:", e);
      }

      delete (window as ICustomWindow).handleGoogleCredentialResponse;
    };
  };

  const initializeGoogleAuth = (buttonId: string, isDarkMode: boolean) => {
    const customWindow = window as ICustomWindow;
    if (!customWindow.google?.accounts) {
      console.error("Google accounts API not available");
      return;
    }

    try {
      const buttonElement = document.getElementById(buttonId);
      if (!buttonElement) {
        console.error(`Element with ID "${buttonId}" not found`);
        return;
      }

      const clientId = getGoogleClientId();
      if (!clientId) {
        console.error("Google Client ID not provided");
        buttonElement.innerHTML = "<div class=\"p-2 text-center text-red-500\">Missing Google Client ID</div>";
        return;
      }

      console.log("Initializing Google Sign-In with client ID and callback");
      customWindow.google.accounts.id.initialize({
        client_id: clientId,
        callback: customWindow.handleGoogleCredentialResponse,
        auto_select: false,
        cancel_on_tap_outside: true
      });

      console.log("Rendering Google Sign-In button");
      customWindow.google.accounts.id.renderButton(buttonElement, {
        theme: isDarkMode ? "filled_black" : "outline",
        size: "large",
        type: "standard",
        shape: "pill",
        text: "continue_with",
        logo_alignment: "left",
        width: buttonElement.offsetWidth
      });
    } catch (error) {
      console.error("Error initializing Google Sign-In:", error);
      const buttonElement = document.getElementById(buttonId);
      if (buttonElement) {
        buttonElement.innerHTML = "<div class=\"p-2 text-center text-red-500\">Error initializing Google Sign-In</div>";
      }
    }
  };

  return {
    loadRecaptcha,
    getRecaptchaToken,
    loadGoogleAuth,
    isGoogleAuthLoaded: () => googleAuthLoaded
  };
}