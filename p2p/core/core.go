package core

import (
	"fmt"
	"log"
	"sync"

	"p2p/hole_punching/client"
	"p2p/shared"
)

type Core struct {
	client *client.Client
	mutex  sync.Mutex
}

func NewCore(addrStr string, username string) *Core {
	client, err := client.NewClient(username, addrStr)
	if err != nil {
		log.Fatal(err)
	}

	client.OnRegistered(registeredCallback)
	client.OnConnecting(connectingCallback)
	client.OnConnected(connectedCallback)
	client.OnMessage(messageCallback)

	return &Core{
		client: client,
	}
}

func (core *Core) SetPeerID(peerID string) {
	core.mutex.Lock()
	core.client.SetOtherPeer(&shared.Peer{ID: peerID})
	core.mutex.Unlock()
}

func (core *Core) Start() error {
	if err := core.client.Start(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (core *Core) Stop() {
	core.client.Stop()
}

func (core *Core) SendMessage(text string) {
	currentPeer := core.client.GetCurrentPeer()
	otherPeerConn := core.client.GetOtherPeerConn()

	if otherPeerConn == nil {
		log.Println("No Peer connected yet!")
		return
	}

	otherPeerConn.Send(&shared.Message{
		Type:    "message",
		PeerID:  currentPeer.ID,
		Content: text,
	})
}

// MARK: - Private

func registeredCallback(client *client.Client) {
	for {
		if otherPeer := client.GetOtherPeer(); otherPeer != nil {
			fmt.Printf("Establishing connection with peer %s...", otherPeer.ID)

			client.GetRDVServerConn().Send(&shared.Message{
				Type:    "establish",
				PeerID:  client.GetCurrentPeer().ID,
				Content: otherPeer.ID,
			})
			return
		}
	}
}

func connectingCallback(client *client.Client) {
	peer := client.GetOtherPeer()
	peerConn := client.GetOtherPeerConn()

	fmt.Println("Connecting to peer...")
	fmt.Printf("Username: %s\n", peer.Username)
	fmt.Printf("ID: %s\n", peer.ID)
	fmt.Printf("Address: %s\n\n", peerConn.GetAddr())
}

func connectedCallback(client *client.Client) {
	peer := client.GetOtherPeer()

	fmt.Printf("Connected to %s!\n", peer.Username)
}

func messageCallback(client *client.Client, text string) {
	fmt.Printf("Received sent a message: %s", text)
}
