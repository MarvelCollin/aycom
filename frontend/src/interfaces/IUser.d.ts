export interface IUser {
  id: string;
  name: string;
  username: string;
  profile_picture?: string;
  verified?: boolean;
  isFollowing?: boolean;
}

export interface IUserProfile {
  id: string;
  username: string;
  displayName: string;
  bio?: string;
  avatar?: string;
  banner?: string;
  profile_picture?: string;
  background_banner_url?: string;
  verified: boolean;
  followerCount: number;
  followingCount: number;
  joinDate: string;
  location?: string;
  website?: string;
  email?: string;
  dateOfBirth?: string;
  gender?: string;
}

export interface IUserUpdateRequest {
  name?: string;
  username?: string;
  email?: string;
  bio?: string;
  location?: string;
  website?: string;
  profile_picture?: string;
  banner?: string;
  birthday?: string;
}

export interface IUserRegistrationRequest {
  name: string;
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
  dob: string;
  gender: string;
  securityQuestion: string;
  securityAnswer: string;
}

export interface IUserVerificationRequest {
  email: string;
  verificationCode: string;
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
  securityAnswer: string;
  newPassword: string;
  confirmPassword: string;
} 