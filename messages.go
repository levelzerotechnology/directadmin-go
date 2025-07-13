package directadmin

import (
	"net/http"
	"time"
)

type Message struct {
	From      string    `json:"from"`
	FromName  string    `json:"fromName"`
	ID        int       `json:"id"`
	LegacyID  string    `json:"legacyID"`
	Message   string    `json:"message"`
	Subject   string    `json:"subject"`
	Timestamp time.Time `json:"timestamp"`
	Unread    bool      `json:"unread"`
}

// GetMessages (user) returns an array of the session user's backups.
func (c *UserContext) GetMessages() ([]*Message, error) {
	var messages []*Message

	if _, err := c.makeRequestNew(http.MethodGet, "messages", nil, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}
