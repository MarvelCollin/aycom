export interface User {
  id: string;
  username: string;
  displayName: string;
  avatar: string | null;
  isVerified?: boolean;
}

export interface ApiUserResponse {
  users?: Array<{
    id: string;
    username?: string;
    name?: string;
    display_name?: string;
    profile_picture_url?: string;
    is_verified?: boolean;
  }>;
  data?: {
    users?: Array<{
      id: string;
      username?: string;
      name?: string;
      display_name?: string;
      profile_picture_url?: string;
      is_verified?: boolean;
    }>;
  };
}

export interface Participant {
  id: string;
  username: string;
  displayName: string;
  avatar: string | null;
  isVerified: boolean;
}

export interface LastMessage {
  content: string;
  timestamp: string | number;
  senderId: string;
  senderName: string;
}

export interface Message {
  id: string;
  content: string;
  timestamp: string;
  senderId: string;
  senderName: string;
  senderAvatar?: string;
  isOwn: boolean;
  isRead: boolean;
  isDeleted: boolean;
  attachments: Attachment[];
}

export interface Attachment {
  id: string;
  type: 'image' | 'gif' | 'video';
  url: string;
  thumbnail?: string;
}

export interface Chat {
  id: string;
  type: 'individual' | 'group';
  name: string;
  avatar: string | null;
  participants: Participant[];
  lastMessage?: LastMessage;
  messages: Message[];
  unreadCount: number;
}

export interface CreateChatResponse {
  chat: {
    id: string;
    name?: string;
    is_group_chat?: boolean;
    participants?: Array<{
      id: string;
      username?: string;
      display_name?: string;
      avatar?: string | null;
      is_verified?: boolean;
    }>;
  };
}

export interface ChatMessage {
  type: string;
  content: string;
  user_id: string;
  chat_id: string;
  timestamp: Date | number;
  message_id: string;
} 