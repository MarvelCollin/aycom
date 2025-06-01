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
  is_admin?: boolean;
  is_banned?: boolean;
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
  subscribe_to_newsletter?: boolean;
  recaptcha_token?: string;
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

/**
 * User API response interfaces
 */
export interface IUserResponse {
  success: boolean;
  data: IUserProfile;
}

export interface IUsersResponse {
  success: boolean;
  data: {
    users: IUser[];
    pagination: {
      total_count: number;
      current_page: number;
      per_page: number;
      total_pages: number;
      has_more?: boolean;
    };
  };
}

export interface IUsernameCheckResponse {
  success: boolean;
  data: {
    available: boolean;
  };
}

export interface IMediaUploadRequest {
  file: File;
}

export interface IMediaUpdateRequest {
  url: string;
}

export interface IMediaUpdateResponse {
  success: boolean;
  data: {
    message: string;
    url: string;
  };
}

export interface IFollowResponse {
  success: boolean;
  data: {
    message: string;
    was_already_following?: boolean;
    is_now_following?: boolean;
  };
}

export interface IFollowStatusResponse {
  success: boolean;
  data: {
    is_following: boolean;
  };
}

export interface IBlockResponse {
  success: boolean;
  data: {
    message: string;
  };
}

export interface IReportUserRequest {
  reason: string;
}

export interface IReportUserResponse {
  success: boolean;
  data: {
    message: string;
  };
}

export interface IBlockedUsersResponse {
  success: boolean;
  data: {
    blocked_users: IUser[];
    pagination: {
      total_count: number;
      current_page: number;
      per_page: number;
    };
  };
}

export interface IAdminStatusUpdateRequest {
  is_admin: boolean;
} 