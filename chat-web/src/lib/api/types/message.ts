export interface ChatMessage {
    id: string;
    type: string;
    conversationId?: number;
    from: string;
    fromName: string;
    to?: string;
    toName?: string;
    message: string;
    timestamp: number;
}

export interface MessageResponse {
    id: number | null; // will be undefined for new messages received via WS before they are saved to DB
    conversation_id: number;
    sender_uuid: string;
    sender_name: string;
    content: string;
    sent_at: string; // ISO 8601 timestamp
    edited_at?: string; // ISO 8601 timestamp
    timestamp?: number; // Unix timestamp in milliseconds
}

export function toMessageResponse(msg: ChatMessage): MessageResponse {
    // const id = Number(msg.id);
    // const conversationId = Number(msg.conversationId);

    // if (Number.isNaN(id) || Number.isNaN(conversationId)) {
    //     throw new Error('Invalid numeric id in ChatMessage');
    // }

    return {
        id : msg.timestamp,
        conversation_id: Number(msg.conversationId),
        sender_uuid: msg.from,
        sender_name: msg.fromName,
        content: msg.message,
        sent_at: new Date(msg.timestamp).toISOString(),
        timestamp: msg.timestamp
    };
}

/* Go Struct

type MessageResponse struct {
	ID             int64      `json:"id"`
	ConversationID int64      `json:"conversation_id"`
	SenderUserUuid *uuid.UUID `json:"sender_uuid,omitempty"`
	SenderUsername string     `json:"sender_name,omitempty"`
	Content        string     `json:"content"`
	SentAt         time.Time  `json:"sent_at"`
	EditedAt       *time.Time `json:"edited_at,omitempty"`
	Timestamp      *int64     `json:"timestamp"`
}

*/