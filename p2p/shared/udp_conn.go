package shared

import (
	"encoding/base64"
	"errors"
	"net"
)

type UDPPayload struct {
	Bytes []byte
	Addr  *net.UDPAddr
}

type UDPConn struct {
	sendChan chan *UDPPayload
	addr     *net.UDPAddr
	secret   string
}

func convertSecret(secretText string) ([32]byte, error) {
	// Ensure secret has been set
	var secret [32]byte
	if secretText == "" {
		return secret, errors.New("secret has not been set")
	}

	// Decode to byte slice
	bytes, err := base64.StdEncoding.DecodeString(secretText)
	if err != nil {
		return secret, errors.New("could not decode secret")
	}

	// copy byte slice into byte array
	copy(secret[:], bytes)

	return secret, nil
}

func (conn *UDPConn) Send(message *Message) error {
	bytes, err := MessageOut(conn, message)
	if err != nil {
		return err
	}

	conn.sendChan <- &UDPPayload{Bytes: bytes, Addr: conn.addr}

	return err
}

func (conn *UDPConn) Protocol() string {
	return "UDP"
}

func (conn *UDPConn) GetAddr() net.Addr {
	return conn.addr
}

func (conn *UDPConn) GetSecret() ([32]byte, error) {
	return convertSecret(conn.secret)
}

func (conn *UDPConn) SetSecret(secret [32]byte) {
	conn.secret = base64.StdEncoding.EncodeToString(secret[:])
}

func NewUDPConn(sendChan chan *UDPPayload, addr *net.UDPAddr) *UDPConn {
	return &UDPConn{
		sendChan: sendChan,
		addr:     addr,
	}
}
