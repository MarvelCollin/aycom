export interface IUser {
  id: string;
  name: string;
  username: string;
  profile_picture_url?: string;
  is_verified?: boolean;
  is_following?: boolean;
}

export interface IUserProfile {
  id: string;
  username: string;
  name: string;
  bio?: string;
  profile_picture_url?: string;
  banner_url?: string;
  is_verified: boolean;
  follower_count: number;
  following_count: number;
  created_at: string;
  location?: string;
  website?: string;
  email?: string;
  date_of_birth?: string;
  gender?: string;
}

export interface IUserUpdateRequest {
  name?: string;
  username?: string;
  email?: string;
  bio?: string;
  location?: string;
  website?: string;
  profile_picture_url?: string;
  banner_url?: string;
  date_of_birth?: string;
}

export interface IUserRegistrationRequest {
  name: string;
  username: string;
  email: string;
  password: string;
  confirm_password: string;
  date_of_birth: string;
  gender: string;
  security_question: string;
  security_answer: string;
}

export interface IUserVerificationRequest {
  email: string;
  verification_code: string;
}

export interface IUserLoginRequest {
  email: string;
  password: string;
}

export interface IUserLoginResponse {
  user: IUser;
  token: string;
}

export interface IPasswordResetRequest {
  email: string;
  security_answer: string;
  new_password: string;
  confirm_password: string;
} 