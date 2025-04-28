// Authentication related interfaces

// Google authentication interfaces
export interface GoogleAccountsId {
  initialize: (config: any) => void;
  renderButton: (element: HTMLElement, options: any) => void;
}

export interface GoogleAccounts {
  id: GoogleAccountsId;
}

export interface Google {
  accounts: GoogleAccounts;
}

// Type for Google credential response
export interface GoogleCredentialResponse {
  credential: string;
}

// reCAPTCHA interface
export interface RecaptchaInstance {
  ready: (callback: () => void) => void;
  render: (container: string, options: any) => number;
}

// Cypress type declaration
declare global {
  interface Window {
    Cypress?: any;
  }
}

// Custom window interface that includes Google and reCAPTCHA properties
export interface CustomWindow extends Window {
  google?: {
    accounts: {
      id: {
        initialize: (config: any) => void;
        renderButton: (element: HTMLElement, options: any) => void;
      }
    }
  };
  grecaptcha?: {
    ready: (callback: () => void) => void;
    render: (container: string | HTMLElement, parameters: any) => number;
    execute: (siteKey: string, options?: { action: string }) => Promise<string>;
    reset: (widgetId?: number) => void;
  };
  handleGoogleCredentialResponse?: (response: GoogleCredentialResponse) => void;
}

// User registration data interface
export interface UserRegistration {
  name: string;
  email: string;
  username: string;
  password: string;
  confirm_password: string;
  gender: string;
  date_of_birth: string;
  security_question: string;
  security_answer: string;
  subscribe_to_newsletter: boolean;
  recaptcha_token: string;
}

// Date of birth interface
export interface DateOfBirth {
  month: string;
  day: string;
  year: string;
}

// Token response from the server
export interface TokenResponse {
  access_token: string;
  refresh_token: string;
  user_id: string;
  token_type?: string;
  expires_in?: number;
}

// AuthStore type
export interface AuthStore {
  isAuthenticated: boolean;
  userId: string | null;
  accessToken: string | null;
  refreshToken: string | null;
} 