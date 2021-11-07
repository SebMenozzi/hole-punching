package client

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"p2p/crypto"

	"p2p/shared"

	"github.com/mitchellh/mapstructure"
)

func createMessageCallback(client *Client) func(shared.Conns, shared.Conn, *shared.Message) {
	return func(conns shared.Conns, conn shared.Conn, message *shared.Message) {
		// Ensure there was no error during registration
		res, err := route(client, conns, conn, message)
		if err != nil {
			fmt.Println(err)
		}

		if res != nil {
			conn.Send(res)
		}
	}
}

func route(client *Client, conns shared.Conns, conn shared.Conn, message *shared.Message) (*shared.Message, error) {
	switch message.Type {
	case "greeting":
		return greetingHandler(client, conn, message)
	case "register":
		return registerHandler(client, conn, message)
	case "establish":
		return establishHandler(client, conn, message)
	case "connect":
		return connectHandler(client, conn, message)
	case "message":
		return messageHandler(client, conn, message)
	}

	return nil, nil
}

func greetingHandler(client *Client, serverConn shared.Conn, message *shared.Message) (*shared.Message, error) {
	currentPeer := client.GetCurrentPeer()

	// Quit the client if greeting fails
	if message.Error != "" {
		return nil, errors.New(message.Error)
	}

	// Ensure that server sent back a public key string
	str, ok := message.Content.(string)
	if !ok {
		return nil, errors.New("expected to receive public key with greeting")
	}

	// Get server public key
	bytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}
	var pubKey [32]byte
	copy(pubKey[:], bytes)

	// Get self public keySent
	serverPubKey, err := currentPeer.GetPublicKey()
	if err != nil {
		return nil, err
	}

	// Create and store secret
	serverConn.SetSecret(crypto.GenSharedSecret(currentPeer.PrivateKey, pubKey))

	// Send register message to server
	return &shared.Message{
		Type:   "register",
		PeerID: currentPeer.ID,
		Content: shared.Registration{
			Username:  currentPeer.Username,
			PublicKey: base64.StdEncoding.EncodeToString(serverPubKey[:]),
		},
	}, nil
}

func registerHandler(client *Client, serverConn shared.Conn, message *shared.Message) (*shared.Message, error) {
	// Quit the client if registration fails
	if message.Error != "" {
		return nil, errors.New(message.Error)
	}

	client.registeredCallback(client)

	return nil, nil
}

func establishHandler(client *Client, serverConn shared.Conn, message *shared.Message) (*shared.Message, error) {
	if message.Error != "" {
		return nil, errors.New(message.Error)
	}

	var peer shared.Peer
	err := mapstructure.Decode(message.Content, &peer)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	client.SetOtherPeer(&shared.Peer{
		ID:       peer.ID,
		Username: peer.Username,
	})

	var addr net.Addr
	switch serverConn.Protocol() {
	case "UDP":
		addr, err = net.ResolveUDPAddr("udp", peer.Endpoint.String())
	case "TCP":
		addr, err = net.ResolveTCPAddr("tcp", peer.Endpoint.String())
	default:
		addr, err = nil, fmt.Errorf("unknown Conn protocol %s", serverConn.Protocol())
	}

	if err != nil {
		return nil, err
	}

	if client.GetOtherPeerConn() != nil && client.GetOtherPeerConn().GetAddr().String() != addr.String() {
		return nil, nil
	}

	go func() {
		otherPeerConn, err := client.GetRDVServer().CreateConn(addr)
		if err != nil {
			return
		}

		client.SetOtherPeerConn(otherPeerConn)

		go client.Connect()

		// Call the callback after the connection
		client.connectingCallback(client)
	}()

	return nil, nil
}

func connectHandler(client *Client, peerConn shared.Conn, message *shared.Message) (*shared.Message, error) {
	otherPeerConn := client.GetOtherPeerConn()
	if otherPeerConn == nil {
		return nil, nil
	}

	if peerConn != otherPeerConn {
		// If addresses are the same then this is the correct peer but the listener has picked up the message and created a new conn
		if otherPeerConn.GetAddr().String() == peerConn.GetAddr().String() {
			otherPeerConn = peerConn
			client.SetOtherPeerConn(otherPeerConn)
		}
		return nil, errors.New("received connect message from unknown peer")
	}

	pubKey, err := client.GetCurrentPeer().GetPublicKey()
	if err != nil {
		return nil, err
	}

	return &shared.Message{
		Type:    "key",
		PeerID:  client.GetCurrentPeer().ID,
		Content: base64.StdEncoding.EncodeToString(pubKey[:]),
	}, nil
}

func messageHandler(client *Client, peerConn shared.Conn, message *shared.Message) (*shared.Message, error) {
	text, ok := message.Content.(string)
	if !ok {
		return nil, errors.New("message message must send some text in content field")
	}

	client.messageCallback(client, text)

	return nil, nil
}
