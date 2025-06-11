import { getAuthToken } from '../utils/auth';
import appConfig from '../config/appConfig';
import { createLoggerWithPrefix } from '../utils/logger';

const API_BASE_URL = appConfig.api.baseUrl;
const logger = createLoggerWithPrefix('notifications-api');

export async function getNotifications() {
  try {
    logger.debug('Fetching notifications from API');
    const response = await fetch(`${API_BASE_URL}/notifications`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${getAuthToken()}`
      },
      credentials: 'include'
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      const errorMessage = errorData.message || 
        `Error ${response.status}: ${response.statusText}`;
      logger.error(`Failed to fetch notifications: ${errorMessage}`);
      throw new Error(errorMessage);
    }

    const data = await response.json();
    logger.info('Successfully fetched notifications', { count: data.notifications?.length || 0 });

    return data.notifications || [];
  } catch (error) {
    logger.error('Error fetching notifications:', error);
    throw error;
  }
}

export async function getUserInteractionNotifications() {
  try {
    logger.debug('Fetching user interaction notifications from API');
    const response = await fetch(`${API_BASE_URL}/notifications/interactions`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${getAuthToken()}`
      },
      credentials: 'include'
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      const errorMessage = errorData.message || 
        `Error ${response.status}: ${response.statusText}`;
      logger.error(`Failed to fetch interaction notifications: ${errorMessage}`);
      throw new Error(errorMessage);
    }

    const data = await response.json();
    logger.info('Successfully fetched interaction notifications', { 
      likes: data.interactions?.likes?.length || 0,
      bookmarks: data.interactions?.bookmarks?.length || 0,
      replies: data.interactions?.replies?.length || 0,
      follows: data.interactions?.follows?.length || 0
    });

    return data.interactions || {
      likes: [],
      bookmarks: [],
      replies: [],
      follows: []
    };
  } catch (error) {
    logger.error('Error fetching interaction notifications:', error);
    throw error;
  }
}

export async function getMentions() {
  try {
    logger.debug('Fetching mentions from API');
    const response = await fetch(`${API_BASE_URL}/notifications/mentions`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${getAuthToken()}`
      },
      credentials: 'include'
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      const errorMessage = errorData.message || 
        `Error ${response.status}: ${response.statusText}`;
      logger.error(`Failed to fetch mentions: ${errorMessage}`);
      throw new Error(errorMessage);
    }

    const data = await response.json();
    logger.info('Successfully fetched mentions', { count: data.mentions?.length || 0 });

    return data.mentions || [];
  } catch (error) {
    logger.error('Error fetching mentions:', error);
    throw error;
  }
}

export async function markNotificationAsRead(notificationId: string) {
  try {
    logger.debug('Marking notification as read', { notificationId });
    const response = await fetch(`${API_BASE_URL}/notifications/${notificationId}/read`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${getAuthToken()}`
      },
      credentials: 'include'
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      const errorMessage = errorData.message || 
        `Error ${response.status}: ${response.statusText}`;
      logger.error(`Failed to mark notification as read: ${errorMessage}`);
      throw new Error(errorMessage);
    }

    const data = await response.json();
    logger.info('Successfully marked notification as read', { notificationId });

    return data;
  } catch (error) {
    logger.error('Error marking notification as read:', error);
    throw error;
  }
}