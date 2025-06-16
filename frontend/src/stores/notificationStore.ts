import { writable, derived } from "svelte/store";
import { toastStore } from "./toastStore";
import { createLoggerWithPrefix } from "../utils/logger";

const logger = createLoggerWithPrefix("NotificationStore");

export interface Notification {
  id: string;
  type: "like" | "reply" | "repost" | "follow" | "mention";
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

  const unreadCount = derived({ subscribe }, $notifications => {
    return $notifications.filter(notification => !notification.read).length;
  });

  const showNotificationToast = (notification: Notification) => {
    const message = getNotificationMessage(notification);
    toastStore.showToast(message, "info");
  };

  const getNotificationMessage = (notification: Notification): string => {
    switch (notification.type) {
      case "like":
        return `${notification.actor_username} liked your post`;
      case "reply":
        return `${notification.actor_username} replied to your post`;
      case "repost":
        return `${notification.actor_username} reposted your post`;
      case "follow":
        return `${notification.actor_username} followed you`;
      case "mention":
        return `${notification.actor_username} mentioned you`;
      default:
        return `New notification from ${notification.actor_username}`;
    }
  };

  return {
    subscribe,
    unreadCount: { subscribe: unreadCount.subscribe },

    addNotification: (notification: Notification) => {
      logger.debug("Adding notification", notification);

      update(notifications => {

        const exists = notifications.some(n => n.id === notification.id);
        if (!exists) {

          showNotificationToast(notification);
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