export interface User {
  id: string;
  username: string;
  name: string;
  profile_picture_url: string | null;
  is_verified?: boolean;
}

export interface ApiUserResponse {
  users?: Array<{
    id: string;
    username?: string;
    name?: string;
    profile_picture_url?: string;
    is_verified?: boolean;
  }>;
  data?: {
    users?: Array<{
      id: string;
      username?: string;
      name?: string;
      profile_picture_url?: string;
      is_verified?: boolean;
    }>;
  };
}

export interface Participant {
  id: string;
  username: string;
  name: string;
  profile_picture_url: string | null;
  is_verified: boolean;
}

export interface LastMessage {
  content: string;
  timestamp: string | number;
  sender_id: string;
  sender_name: string;
}

export interface Message {
  id: string;
  content: string;
  timestamp: string;
  sender_id: string;
  sender_name: string;
  sender_profile_picture?: string;
  is_own: boolean;
  is_read: boolean;
  is_deleted: boolean;
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
  profile_picture_url: string | null;
  participants: Participant[];
  last_message?: LastMessage;
  messages: Message[];
  unread_count: number;
}

export interface CreateChatResponse {
  chat: {
    id: string;
    name?: string;
    is_group_chat?: boolean;
    participants?: Array<{
      id: string;
      username?: string;
      name?: string;
      profile_picture_url?: string | null;
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