export interface IGoogleCredentialResponse {
  credential: string;
  clientId: string;
  select_by: string;
}

export interface IAuthStore {
  userId: string | null;
  isAuthenticated: boolean;
  accessToken: string | null;
  refreshToken: string | null;
} 