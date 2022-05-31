package message

import (
	"net"

	"github.com/VictorAnnell/kandidat-backend/rediscli"
	"github.com/gobwas/ws"
)

type DataChannelMessages struct {
	RecipientUUID    string              `json:"recipientUUID,omitempty"`
	Offset           int64               `json:"offset,omitempty"`
	Limit            int64               `json:"limit,omitempty"`
	Messages         []*rediscli.Message `json:"messages,omitempty"`
	MessagesTotal    int64               `json:"messagesTotal,omitempty"`
	MessagesRecieved int                 `json:"messagesRecieved,omitempty"`
}

func (p Controller) ChannelMessages(sessionUUID string, conn net.Conn, op ws.OpCode, writer Write, message *Message) IError {

	channelUUID, err := p.r.GetChannelUUID(message.UserID, message.ChannelMessages.RecipientUUID)
	if err != nil {
		return newError(404, err)
	}

	channelMessages, err := p.r.ChannelMessages(channelUUID, message.ChannelMessages.Offset, message.ChannelMessages.Limit)
	if err != nil {
		return newError(0, err)
	}

	messagesCount, err := p.r.ChannelMessagesCount(channelUUID)
	if err != nil {
		return newError(0, err)
	}

	err = writer(conn, op, &Message{
		Type: DataTypeChannelMessages,
		ChannelMessages: &DataChannelMessages{
			MessagesTotal:    messagesCount,
			MessagesRecieved: len(channelMessages),
			Messages:         channelMessages,
		},
	})
	if err != nil {
		return newError(0, err)
	}

	return nil
}
