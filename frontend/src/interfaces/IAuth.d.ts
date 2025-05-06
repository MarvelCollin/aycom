// Authentication related interfaces

// Google authentication interfaces
export interface IGoogleAccountsId {
  initialize: (config: any) => void;
  renderButton: (element: HTMLElement, options: any) => void;
}

export interface IGoogleAccounts {
  id: IGoogleAccountsId;
}

export interface IGoogle {
  accounts: IGoogleAccounts;
}

// Type for Google credential response
export interface IGoogleCredentialResponse {
  credential: string;
}

// reCAPTCHA interface
export interface IRecaptchaInstance {
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
export interface ICustomWindow extends Window {
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
  handleGoogleCredentialResponse?: (response: IGoogleCredentialResponse) => void;
}

// User registration data interface
export interface IUserRegistration {
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
  profile_picture_url?: string;
  banner_url?: string;
}

// Date of birth interface
export interface IDateOfBirth {
  month: string;
  day: string;
  year: string;
}

// Token response from the server
export interface ITokenResponse {
  access_token: string;
  refresh_token: string;
  user_id: string;
  token_type?: string;
  expires_in?: number;
}

// AuthStore type
export interface IAuthStore {
  userId: string | null;
  isAuthenticated: boolean;
  accessToken: string | null;
  refreshToken: string | null;
  username?: string | null;
  displayName?: string | null;
}

export interface ILoginCredentials {
  username: string;
  password: string;
}

export interface IRegisterCredentials {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
} 