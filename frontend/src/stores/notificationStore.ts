import { writable } from 'svelte/store';

export interface Notification {
  id: string;
  type: 'like' | 'reply' | 'repost' | 'follow' | 'mention';
  actor_id: string;
  actor_username: string;
  actor_profile_pic: string;
  target_id: string;
  timestamp: string;
  read: boolean;
  content: string;
}

function createNotificationStore() {
  const { subscribe, update, set } = writable<Notification[]>([]);

  return {
    subscribe,

    addNotification: (notification: Notification) => {
      update(notifications => {

        const exists = notifications.some(n => n.id === notification.id);
        if (!exists) {
          return [notification, ...notifications];
        }
        return notifications;
      });
    },

    markAsRead: (id: string) => {
      update(notifications => {
        return notifications.map(notification => {
          if (notification.id === id) {
            return { ...notification, read: true };
          }
          return notification;
        });
      });
    },

    markAllAsRead: () => {
      update(notifications => {
        return notifications.map(notification => ({ ...notification, read: true }));
      });
    },

    removeNotification: (id: string) => {
      update(notifications => {
        return notifications.filter(notification => notification.id !== id);
      });
    },

    reset: () => {
      set([]);
    }
  };
}

export const notificationStore = createNotificationStore();