/**
 * Index file to export all interfaces
 */

// Common interfaces
export * from './ICommon';

// Auth interfaces
export * from './IAuth';

// Media interfaces - these need to be exported first to avoid conflicts
export * from './IMedia';

// User interfaces - export everything except IMediaUpdateResponse which is already exported from IMedia.ts
export {
  type IUser,
  type IUserProfile,
  type IUserUpdateRequest,
  type IUserRegistrationRequest,
  type IUserVerificationRequest,
  type IUserLoginRequest,
  type IUserLoginResponse,
  type IPasswordResetRequest,
  type IUserResponse,
  type IUsersResponse,
  type IUsernameCheckResponse,
  type IMediaUploadRequest,
  type IFollowResponse,
  type IFollowStatusResponse,
  type IBlockResponse,
  type IReportUserRequest,
  type IReportUserResponse,
  type IBlockedUsersResponse,
  type IAdminStatusUpdateRequest
} from './IUser';

// Trend interfaces - these need to be exported first to avoid conflicts
export * from './ITrend';

// Social media interfaces
export * from './ISocialMedia';

// Chat interfaces
export * from './IChat';

// Community interfaces
export * from './ICommunity';

// Category interfaces
export * from './ICategory';

// Admin interfaces
export * from './IAdmin';

// Notification interfaces
export * from './INotification';

// Bookmark interfaces
export * from './IBookmark';

// AI interfaces
export * from './IAI';

// Search interfaces
export * from './ISearch';

// Toast interfaces
export * from './IToast';

// Extended tweet utilities
export * from './ITweet.extended'; 