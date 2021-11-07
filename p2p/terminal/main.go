package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"p2p/hole_punching/client"
	"p2p/shared"
)

func main() {
	fmt.Println("- Terminal Client - ")

	// Get username from user
	var username string
	for username == "" || len(username) > 32 {
		fmt.Println("Username (<= 32 characters)")
		fmt.Print("> ")
		if _, err := fmt.Scanln(&username); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("Nice to meet you %s!\n", username)

	client, err := client.NewClient(username, "127.0.0.1:9001")
	if err != nil {
		log.Fatal(err)
	}

	client.OnRegistered(registeredCallback)
	client.OnConnecting(connectingCallback)
	client.OnConnected(connectedCallback)
	client.OnMessage(messageCallback)

	client.Start()

	exit := make(chan os.Signal)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println(<-exit)

	client.Stop()
}

// MARK: - Private

func registeredCallback(client *client.Client) {
	var peerID string
	for peerID == "" {
		fmt.Println("PeerID")
		fmt.Print("> ")
		if _, err := fmt.Scanln(&peerID); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("Establishing connection with peer %s...\n", peerID)

	client.SetOtherPeer(&shared.Peer{ID: peerID})
	client.GetRDVServerConn().Send(&shared.Message{
		Type:    "establish",
		PeerID:  client.GetCurrentPeer().ID,
		Content: peerID,
	})
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

	fmt.Printf("Connected to %s over an encrypted channel\n", peer.Username)

	go func() {
		for {
			var text string
			for text == "" {
				fmt.Println("Text")
				fmt.Print("> ")
				if _, err := fmt.Scanln(&text); err != nil {
					log.Fatal(err)
				}
			}

			currentPeer := client.GetCurrentPeer()
			otherPeerConn := client.GetOtherPeerConn()

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
	}()
}

func messageCallback(client *client.Client, text string) {
	fmt.Printf("Received sent a message: %s", text)
}
