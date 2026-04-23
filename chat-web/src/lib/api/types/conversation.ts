// participant.ts
export interface Participant {
  user_uuid: string;   // corresponds to uuid.UUID
  username: string;
  isAdmin?: boolean; // optional for new private chats, required for group chats
}

// conversation-request.ts
// import type { Participant } from './participant';

export interface ConversationRequest {
  name?: string | null;              // optional
  type: 'private' | 'group';         // matches validate:"oneof=private group"
  participants?: Participant[];      // optional, can enforce min/max length at runtime if needed
}

// conversation-response.ts
export interface ConversationResponse {
  conversation_id: number;           // corresponds to int64
  name?: string | null;              // optional
  type: 'private' | 'group';
  created_at: Date;                  // time.Time
  created_by: string;              // corresponds to uuid.UUID
  last_message_at?: Date | null;     // optional
}

export interface PrivateConversationResponse {
  conversation_id: number;           // corresponds to int64
  type: 'private' | 'group';
}