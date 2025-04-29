export interface GoogleCredentialResponse {
  credential: string;
  clientId: string;
  select_by: string;
}

export interface CustomWindow extends Window {
  grecaptcha?: {
    ready: (callback: () => void) => void;
    render: (container: string | HTMLElement, options: any) => number;
    reset: (id: number) => void;
    getResponse: (id: number) => string;
  };
  google?: {
    accounts: {
      id: {
        initialize: (config: any) => void;
        renderButton: (element: HTMLElement, options: any) => void;
      };
    };
  };
  handleGoogleCredentialResponse?: (response: GoogleCredentialResponse) => void;
} 