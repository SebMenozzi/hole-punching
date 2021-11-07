package client

import (
	"encoding/base64"
	"encoding/hex"
	"net"
	"sync"
	"time"

	"p2p/crypto"
	"p2p/hole_punching/server"
	"p2p/shared"
)

type Client struct {
	rdvServer *server.Server
	addr      *net.UDPAddr

	// Current peer
	currentPeer *shared.Peer
	// Other peer (peer to connect with)
	otherPeer *shared.Peer

	// Server conn
	rdvServerConn shared.Conn
	// Other peer conn
	otherPeerConn shared.Conn

	mutex *sync.Mutex

	registeredCallback func(client *Client)
	connectingCallback func(client *Client)
	connectedCallback  func(client *Client)
	messageCallback    func(client *Client, text string)
}

func (client *Client) Connect() {
	otherPeerConn := client.GetOtherPeerConn()
	currentPeer := client.GetCurrentPeer()

	// Sends 5 connect messages
	for i := 0; i < 5; i += 1 {
		client.connectedCallback(client)

		otherPeerConn.Send(&shared.Message{
			Type:   "connect",
			PeerID: currentPeer.ID,
		})

		time.Sleep(3 * time.Second)
	}
}

func (client *Client) Start() error {
	rdvServer := client.GetRDVServer()

	// Add rendez-vous server connection
	serverConn, err := rdvServer.CreateConn(client.addr)
	if err != nil {
		return err
	}

	client.SetRDVServerConn(serverConn)

	// Get public key
	pubKey, err := client.GetCurrentPeer().GetPublicKey()
	if err != nil {
		return err
	}

	// Start rendez-vous server
	go rdvServer.Listen()

	// Send greeting message to server
	serverConn.Send(&shared.Message{
		Type:    "greeting",
		Content: base64.StdEncoding.EncodeToString(pubKey[:]),
	})

	return nil
}

func NewClient(
	username string,
	addrStr string,
) (*Client, error) {
	clientAddr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		return nil, err
	}

	// Create UDP server
	rdvServer, err := server.NewServer(shared.GenPort())
	if err != nil {
		return nil, err
	}

	// Create current peer
	currentPeer := &shared.Peer{
		Username: username,
	}

	// Create public and private keys
	var pubKey [32]byte
	currentPeer.PrivateKey, pubKey, err = crypto.GenKeyPair()
	if err != nil {
		return nil, err
	}

	currentPeer.SetPublicKey(pubKey)

	// Create client ID: SHA-2 + HMAC hash of public key
	currentPeer.ID = hex.EncodeToString(crypto.Hash("Hashing client public key for client id", pubKey[:]))

	client := &Client{
		rdvServer:          rdvServer,
		addr:               clientAddr,
		currentPeer:        currentPeer,
		otherPeer:          nil,
		mutex:              &sync.Mutex{},
		registeredCallback: func(*Client) {},
		connectingCallback: func(*Client) {},
		connectedCallback:  func(*Client) {},
		messageCallback:    func(*Client, string) {},
	}

	rdvServer.OnMessage(createMessageCallback(client))

	return client, nil
}

func (client *Client) GetCurrentPeer() *shared.Peer {
	return client.currentPeer
}

func (client *Client) GetOtherPeer() *shared.Peer {
	return client.otherPeer
}

func (client *Client) SetOtherPeer(peer *shared.Peer) {
	client.otherPeer = peer
}

func (client *Client) GetOtherPeerConn() shared.Conn {
	return client.otherPeerConn
}

func (client *Client) SetOtherPeerConn(conn shared.Conn) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.otherPeerConn = conn
}

func (client *Client) GetRDVServerConn() shared.Conn {
	return client.rdvServerConn
}

func (client *Client) SetRDVServerConn(conn shared.Conn) {
	client.rdvServerConn = conn
}

func (client *Client) GetRDVServer() *server.Server {
	return client.rdvServer
}

func (client *Client) OnRegistered(callback func(client *Client)) {
	client.registeredCallback = callback
}

func (client *Client) OnConnecting(callback func(client *Client)) {
	client.connectingCallback = callback
}

func (client *Client) OnConnected(callback func(client *Client)) {
	client.connectedCallback = callback
}

func (client *Client) OnMessage(callback func(client *Client, text string)) {
	client.messageCallback = callback
}

func (client *Client) Stop() {
	client.rdvServer.Stop()
}
