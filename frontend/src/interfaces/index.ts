/**
 * Index file to export all interfaces
 */

// Common interfaces
export * from './ICommon';

// Auth interfaces
export * from './IAuth';

// User interfaces
export * from './IUser';

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

// Media interfaces - resolve ambiguous exports
export { 
  type IMediaUploadResponse, 
  type IMediaSearchResponse 
} from './IMedia';
// Re-export IMedia specifically from IMedia.ts
export { type IMedia as IMediaType } from './IMedia';

// Notification interfaces
export * from './INotification';

// Trend interfaces - resolve ambiguous exports
export { type ITrendsResponse } from './ITrend';
// Re-export ITrend specifically from ITrend.ts
export { type ITrend as ITrendItem } from './ITrend';

// Bookmark interfaces
export * from './IBookmark';

// AI interfaces
export * from './IAI';

// Search interfaces
export * from './ISearch';

// Toast interfaces
export * from './IToast'; 