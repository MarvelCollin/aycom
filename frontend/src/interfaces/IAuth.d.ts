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

export interface IGoogleCredentialResponse {
  credential: string;
}

export interface IRecaptchaInstance {
  ready: (callback: () => void) => void;
  render: (container: string, options: any) => number;
}

declare global {
  interface Window {
    Cypress?: any;
  }
}

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
    getResponse: (widgetId: number) => string;
  };
  handleGoogleCredentialResponse?: (response: IGoogleCredentialResponse) => void;
}

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

export interface IDateOfBirth {
  month: string;
  day: string;
  year: string;
}

export interface ITokenResponse {
  access_token: string;
  refresh_token: string;
  user_id: string;
  token_type?: string;
  expires_in?: number;
}

export interface IAuthStore {
  user_id: string | null;
  is_authenticated: boolean;
  access_token: string | null;
  refresh_token: string | null;
  username?: string | null;
  name?: string | null;
  is_admin?: boolean;
}

export interface ILoginCredentials {
  email: string;
  password: string;
}

export interface IRegisterCredentials {
  username: string;
  email: string;
  password: string;
  confirm_password: string;
}