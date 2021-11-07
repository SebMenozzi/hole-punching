package server

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"

	"p2p/crypto"
	"p2p/shared"
)

func createMessageCallback(server *Server, peers shared.Peers) func(conns shared.Conns, conn shared.Conn, message *shared.Message) {
	return func(conns shared.Conns, conn shared.Conn, message *shared.Message) {
		// Log request
		log.Printf("Request from client at %s over %s with type %s", conn.GetAddr(), conn.Protocol(), message.Type)

		// Route request to a handler
		res, err := route(server, peers, conns, conn, message)

		// Respond with error if there was one
		if err != nil {
			conn.Send(&shared.Message{
				Type:  message.Type,
				Error: err.Error(),
			})
			return
		}

		// Respond
		err = conn.Send(res)
		if err != nil {
			log.Print(err)
		}
	}
}

func route(server *Server, peers shared.Peers, conns shared.Conns, conn shared.Conn, message *shared.Message) (*shared.Message, error) {
	switch message.Type {
	case "greeting":
		return greetingHandler(server, conn, message)
	case "register":
		return registerHandler(peers, conn, message)
	case "establish":
		return establishHandler(peers, conns, message)
	default:
		return notFoundHandler(message)
	}
}

func greetingHandler(server *Server, conn shared.Conn, message *shared.Message) (*shared.Message, error) {
	// Ensure that public key was sent in greeting request
	str, ok := message.Content.(string)
	if !ok {
		return nil, fmt.Errorf("greeting request must contain client's public key")
	}

	// Get public key contained in content
	bs, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}

	// Create shared secret from private key and peer public key
	var clientPubKey [32]byte
	copy(clientPubKey[:], bs[:])
	conn.SetSecret(crypto.GenSharedSecret(server.privateKey, clientPubKey))

	// Send greeting response
	return &shared.Message{
		Type:    "greeting",
		Content: base64.StdEncoding.EncodeToString(server.publicKey[:]),
	}, nil
}

// Register the requesting peer in the server
func registerHandler(peers shared.Peers, conn shared.Conn, message *shared.Message) (*shared.Message, error) {
	// Map -> structure the content
	var registration shared.Registration
	err := mapstructure.Decode(message.Content, &registration)
	if err != nil {
		return nil, err
	}

	// Register peer
	endpoint := strings.Split(conn.GetAddr().String(), ":")
	if len(endpoint) != 2 {
		return nil, fmt.Errorf("address is not valid")
	}

	port, err := strconv.Atoi(endpoint[1])
	if err != nil {
		return nil, err
	}

	peers[message.PeerID] = &shared.Peer{
		ID:       message.PeerID,
		Username: registration.Username,
		Endpoint: shared.Endpoint{
			IP:   endpoint[0],
			Port: port,
		},
	}

	log.Printf("Registered peer: %s at addr %s", message.PeerID, conn.GetAddr().String())

	// Confirm registry to peer
	return &shared.Message{
		Type:    "register",
		Encrypt: true,
	}, nil
}

// Facilitate in the establishing of the p2p connection
func establishHandler(peers shared.Peers, conns shared.Conns, message *shared.Message) (*shared.Message, error) {
	// Make sure requesting peer has registered with server
	rp, ok := peers[message.PeerID]
	if !ok {
		return nil, fmt.Errorf("client is not registered with this server")
	}

	// Make sure that a valid payload was sent
	id, ok := message.Content.(string)
	if !ok {
		return nil, fmt.Errorf("request content is malformed")
	}

	// Make sure the other peer has registered with the server
	op, ok := peers[id]
	if !ok {
		return nil, fmt.Errorf("peer: %s has not registered with the server", id)
	}

	// Get conn for other peer
	conn, ok := conns[op.Endpoint.String()]
	if !ok {
		return nil, fmt.Errorf("could not resolve the peer: %s's conn", id)
	}

	// Send requesting peer's endpoint to other peer
	conn.Send(&shared.Message{
		Type:    "establish",
		Content: rp,
		Encrypt: true,
	})

	// Send requesting peer other peer's endpoint
	return &shared.Message{
		Type:    "establish",
		Content: op,
		Encrypt: true,
	}, nil
}

func notFoundHandler(message *shared.Message) (*shared.Message, error) {
	return nil, fmt.Errorf("request type %s undefined", message.Type)
}
