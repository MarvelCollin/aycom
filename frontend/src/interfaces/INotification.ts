/**
 * Notification-related interfaces
 */

import type { IPagination } from "./ICommon";

/**
 * Notification type enum
 */
export enum NotificationType {
  MESSAGE = "message",
  LIKE = "like",
  FOLLOW = "follow",
  REPLY = "reply",
  MENTION = "mention",
  COMMUNITY = "community",
  SYSTEM = "system"
}

/**
 * Notification interface
 */
export interface INotification {
  id: string;
  user_id: string;
  type: NotificationType;
  content: string;
  data?: any;
  read: boolean;
  created_at: string;
}

/**
 * Notifications response
 */
export interface INotificationsResponse {
  success: boolean;
  data: {
    notifications: INotification[];
    pagination: IPagination;
  };
}

/**
 * Mark notification as read request
 */
export interface IMarkNotificationReadRequest {
  notification_id: string;
}

/**
 * Mark notification as read response
 */
export interface IMarkNotificationReadResponse {
  success: boolean;
  data: {
    message: string;
  };
}

/**
 * Mark all notifications as read response
 */
export interface IMarkAllNotificationsReadResponse {
  success: boolean;
  data: {
    message: string;
  };
}

/**
 * Delete notification response
 */
export interface IDeleteNotificationResponse {
  success: boolean;
  data: {
    message: string;
  };
}

/**
 * Notification preferences
 */
export interface INotificationPreferences {
  likes: boolean;
  comments: boolean;
  follows: boolean;
  mentions: boolean;
  direct_messages: boolean;
}

/**
 * Notification preferences response
 */
export interface INotificationPreferencesResponse {
  success: boolean;
  data: {
    preferences: INotificationPreferences;
  };
}

/**
 * Update notification preferences request
 */
export interface IUpdateNotificationPreferencesRequest {
  likes?: boolean;
  comments?: boolean;
  follows?: boolean;
  mentions?: boolean;
  direct_messages?: boolean;
}

/**
 * Update notification preferences response
 */
export interface IUpdateNotificationPreferencesResponse {
  success: boolean;
  data: {
    message: string;
    preferences: INotificationPreferences;
  };
}