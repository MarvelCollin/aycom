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

/**
 * Extended chat participant interface to accommodate legacy data structure
 */
export interface ChatParticipant extends Partial<Participant> {
  id?: string;
  user_id?: string;
  username?: string;
  display_name?: string;
  avatar_url?: string;
  profile_picture_url?: string | null;
  name?: string;
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

/**
 * Extended chat message interface to accommodate legacy data structure
 */
export interface ChatMessage {
  id?: string;
  message_id?: string;
  sender_id?: string;
  chat_id?: string;
  user_id?: string;
  content: string;
  timestamp: number | string;
  is_read?: boolean;
  is_deleted?: boolean;
  is_edited?: boolean;
  user?: {
    id: string;
    username?: string;
    display_name?: string;
    avatar_url?: string;
    name?: string;
  };
  type?: string;
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

/**
 * Interface for message with user data as used in the store
 */
export interface MessageWithUser {
  type: string;
  content: string;
  user_id: string;
  chat_id: string;
  timestamp: Date | number;
  message_id: string;
  user: {
    id: string;
    username: string;
    name: string;
    profile_picture_url: string | null;
  };
  is_read?: boolean;
  is_deleted?: boolean;
  is_edited?: boolean;
  attachments?: Attachment[];
}

/**
 * Chat API response interfaces
 */
export interface IChatResponse {
  success: boolean;
  data: {
    chat: Chat;
  };
}

export interface IChatsResponse {
  success: boolean;
  data: {
    chats: Chat[];
    pagination: {
      total_count: number;
      current_page: number;
      per_page: number;
      total_pages: number;
      has_more?: boolean;
    };
  };
}

export interface IParticipantsResponse {
  success: boolean;
  data: {
    participants: Participant[];
  };
}

export interface ICreateChatRequest {
  type: 'individual' | 'group';
  name?: string;
  participants: string[];
}

export interface IAddParticipantRequest {
  user_id: string;
  is_admin?: boolean;
}

export interface IParticipantResponse {
  success: boolean;
  data: {
    participant: Participant;
  };
}

export interface ISendMessageRequest {
  content: string;
  media_url?: string;
  media_type?: string;
  reply_to_message_id?: string;
}

export interface IMessageResponse {
  success: boolean;
  data: {
    message: {
      id: string;
      chat_id: string;
      sender_id: string;
      content: string;
      media_url?: string;
      media_type?: string;
      sent_at: string;
      unsent: boolean;
      reply_to_message_id?: string;
    };
  };
}

export interface IMessagesResponse {
  success: boolean;
  data: {
    messages: Array<{
      id: string;
      chat_id: string;
      sender_id: string;
      content: string;
      media_url?: string;
      media_type?: string;
      sent_at: string;
      unsent: boolean;
      deleted_for_sender: boolean;
      deleted_for_all: boolean;
      reply_to_message_id?: string;
    }>;
    pagination: {
      total_count: number;
      current_page: number;
      per_page: number;
      total_pages: number;
      has_more?: boolean;
    };
  };
}

export interface IChatHistoryResponse {
  success: boolean;
  data: {
    history: Array<{
      chat_id: string;
      name: string;
      is_group: boolean;
      last_message?: {
        content: string;
        timestamp: string;
        sender_id: string;
        sender_name: string; 
      };
      unread_count: number;
    }>;
    pagination: {
      total_count: number;
      current_page: number;
      per_page: number;
      total_pages: number;
      has_more?: boolean;
    };
  };
} 