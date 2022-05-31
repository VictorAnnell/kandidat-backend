package message

import (
	"net"
	"time"

	"github.com/VictorAnnell/kandidat-backend/rediscli"
	"github.com/gobwas/ws"
	"github.com/google/uuid"
)

type DataChannelMessage struct {
	UUID          string         `json:"UUID"`
	Sender        *rediscli.User `json:"Sender,omitempty"`
	SenderID      string         `json:"SenderID"`
	Recipient     *rediscli.User `json:"Recipient,omitempty"`
	RecipientUUID string         `json:"RecipientUUID"`
	Message       string         `json:"Message"`
	CreatedAt     time.Time      `json:"CreatedAt"`
}

func (p Controller) ChannelMessage(sessionUUID string, conn net.Conn, op ws.OpCode, writer Write, message *Message) IError {

	channelMessage := &rediscli.Message{
		UUID:          uuid.NewString(),
		SenderID:      message.UserID,
		RecipientUUID: message.ChannelMessage.RecipientUUID,
		Message:       message.ChannelMessage.Message,
		CreatedAt:     time.Now(),
	}

	channelUUID, err := p.r.ChannelMessage(channelMessage)
	if err != nil {
		return nil
	}

	channelSessionsSendMessage(message.UserID, channelUUID, writer, message)

	return nil
}
