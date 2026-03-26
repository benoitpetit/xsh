// Package models provides data models for direct messages.
package models

// DMParticipant represents a participant in a DM conversation
type DMParticipant struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Handle string `json:"handle"`
}

// DMConversation represents a DM conversation
type DMConversation struct {
	ID              string          `json:"id"`
	Type            string          `json:"type"` // "one_to_one" or "group"
	Participants    []DMParticipant `json:"participants"`
	LastMessage     string          `json:"last_message"`
	LastMessageTime string          `json:"last_message_time"`
	Unread          bool            `json:"unread"`
}
