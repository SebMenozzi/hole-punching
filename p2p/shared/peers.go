package shared

import (
	"encoding/base64"
	"net"
	"strconv"
)

type Endpoint struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

func (endpoint Endpoint) String() string {
	return endpoint.IP + ":" + strconv.Itoa(endpoint.Port)
}

type Peer struct {
	ID         string       `json:"id,omitempty"`
	Username   string       `json:"username,omitempty"`
	Endpoint   Endpoint     `json:"endpoint,omitempty"`
	PublicKey  string       `json:"publicKey,omitempty"`
	PrivateKey [32]byte     `json:"-"`
	Addr       *net.UDPAddr `json:"-"`
}

func (peer *Peer) GetPublicKey() ([32]byte, error) {
	var key [32]byte
	bytes, err := base64.StdEncoding.DecodeString(peer.PublicKey)
	if err != nil {
		return key, err
	}

	copy(key[:], bytes)

	return key, nil
}

func (peer *Peer) SetPublicKey(key [32]byte) {
	peer.PublicKey = base64.StdEncoding.EncodeToString(key[:])
}

type Peers map[string]*Peer
