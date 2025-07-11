import type { IUserRegistrationRequest } from "./IUser";

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

export type IUserRegistration = IUserRegistrationRequest;

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

export type IRegisterCredentials = Pick<IUserRegistrationRequest, "username" | "email" | "password" | "confirm_password">;

export interface IAuthResponse {
  success: boolean;
  message?: string;
  access_token: string;
  refresh_token: string;
  user_id: string;
  token_type: string;
  expires_in: number;
  user?: any;
}

export interface IRefreshTokenRequest {
  refresh_token: string;
}

export interface IOAuthConfigResponse {
  success: boolean;
  data: {
    providers: Array<{
      name: string;
      client_id: string;
      auth_url: string;
      scopes: string[];
    }>;
  };
}

export interface IVerifyEmailRequest {
  email: string;
  verification_code: string;
}

export interface IVerifyEmailResponse {
  success: boolean;
  message: string;
  access_token?: string;
  refresh_token?: string;
  user_id?: string;
  expires_in?: number;
  token_type?: string;
}

export interface IResendVerificationRequest {
  email: string;
}

export interface IResendVerificationResponse {
  success: boolean;
  message: string;
}

export interface IGoogleLoginRequest {
  token_id: string;
}

export interface IForgotPasswordRequest {
  email: string;
}

export interface IForgotPasswordResponse {
  success: boolean;
  data: {
    message: string;
    security_question?: string;
    email?: string;
  };
}

export interface IVerifySecurityAnswerRequest {
  email: string;
  security_answer: string;
}

export interface IVerifySecurityAnswerResponse {
  success: boolean;
  data: {
    message: string;
    token: string;
    email: string;
    expires: string;
  };
}

export interface IResetPasswordRequest {
  token: string;
  email: string;
  new_password: string;
}

export interface IResetPasswordResponse {
  success: boolean;
  data: {
    message: string;
  };
}