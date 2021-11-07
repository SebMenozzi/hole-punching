package shared

import "net"

type Conn interface {
	Send(*Message) error
	Protocol() string
	GetAddr() net.Addr
	GetSecret() ([32]byte, error)
	SetSecret([32]byte)
}

type Conns map[string]Conn
