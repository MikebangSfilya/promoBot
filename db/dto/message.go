package dto

// Message represents a reference to a message, i.e. the primary key of the Suggestions table where UID == ChatID.
// In most cases, this is the message in the private chat with some user.
type Message struct {
	ChatID    int64
	MessageID int
}

func NewMessage(chatID int64, messageID int) *Message {
	return &Message{
		ChatID:    chatID,
		MessageID: messageID,
	}
}
