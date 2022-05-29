package message

import (
	"net"

	"github.com/gobwas/ws"
)

type DataSignUp struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (p Controller) SignUp(sessionUUID string, conn net.Conn, op ws.OpCode, write Write, message *Message) IError {
	user, err := p.r.UserCreate(message.SignUp.Username, message.SignUp.Password)
	if err != nil {
		return newError(0, err)
	}

	err = write(conn, op, &Message{
		Type: DataTypeAuthorized,
		Authorized: &DataAuthorized{
			UserUUID:  user.ID,
			AccessKey: user.AccessKey,
		},
	})
	if err != nil {
		return newError(0, err)
	}

	return nil

}
