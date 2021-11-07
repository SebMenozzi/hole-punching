package shared

import "net"

type Message struct {
	Type    string      `json:"type"`
	PeerID  string      `json:"peerID,omitempty"`
	Error   string      `json:"error,omitempty"`
	Content interface{} `json:"data,omitempty"`
	Encrypt bool        `json:"-"`
	addr    *net.UDPAddr
}

func (message *Message) GetAddr() *net.UDPAddr {
	return message.addr
}

func (message *Message) SetAddr(addr *net.UDPAddr) *Message {
	message.addr = addr
	return message
}

// Message type registration
type Registration struct {
	Username  string `json:"username"`
	PublicKey string `json:"publicKey"`
}
