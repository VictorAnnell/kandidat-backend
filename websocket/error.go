package websocket

type IError interface {
	Error() (uint32, error)
}

const (
	errCode uint32 = iota
	errCodeJSUnmarshal
)
