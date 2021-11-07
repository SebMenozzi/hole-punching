package server

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"p2p/crypto"
	"p2p/shared"
)

type Server struct {
	conn            *net.UDPConn
	publicKey       [32]byte
	privateKey      [32]byte
	conns           shared.Conns
	sendChan        chan *shared.UDPPayload
	messageCallback func(shared.Conns, shared.Conn, *shared.Message)
	exit            chan bool
	wg              *sync.WaitGroup
}

func (server *Server) sender() {
	server.wg.Add(1)
	defer server.wg.Done()

	for {
		select {
		case <-server.exit:
			log.Print("Exiting UDP sender")
			return
		case payload := <-server.sendChan:
			_, err := server.conn.WriteToUDP(payload.Bytes, payload.Addr)
			if err != nil {
				log.Print(err)
			}
		}
	}
}

func (server *Server) serve(b []byte, conn shared.Conn) {
	defer server.wg.Done()

	message, err := shared.MessageIn(conn, b)
	if err != nil {
		conn.Send(&shared.Message{
			Error: "Malformed payload was sent",
		})
		return
	}

	go server.messageCallback(server.conns, conn, message)
}

func (server *Server) receiver() {
	server.wg.Add(1)
	defer server.wg.Done()

	for {
		select {
		case <-server.exit:
			log.Print("Exiting UDP receiver")
			server.conn.Close()
			return
		default:
		}

		buffer := make([]byte, 2048)
		server.conn.SetDeadline(time.Now().Add(time.Second))
		n, addr, err := server.conn.ReadFromUDP(buffer)
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}

			delete(server.conns, addr.String())
			log.Print(err)
			return
		}

		conn, ok := server.conns[addr.String()]
		if !ok {
			conn = shared.NewUDPConn(server.sendChan, addr)
			server.conns[addr.String()] = conn
		}

		// Process message
		server.wg.Add(1)

		go server.serve(buffer[:n], conn)
	}
}

func (server *Server) CreateConn(addr net.Addr) (shared.Conn, error) {
	if addr == nil {
		return nil, errors.New("conns addr must not be nil")
	}

	udpAddr, ok := addr.(*net.UDPAddr)
	if !ok {
		return nil, errors.New("could not assert net.Addr to *net.UDPAddr")
	}

	conn := shared.NewUDPConn(server.sendChan, udpAddr)
	server.conns[addr.String()] = conn

	return conn, nil
}

func (server *Server) OnMessage(callback func(conns shared.Conns, conn shared.Conn, message *shared.Message)) {
	server.messageCallback = callback
}

func (server *Server) Stop() {
	close(server.exit)
	server.wg.Wait()

	log.Print("UDP server exited")
}

func (server *Server) Listen() {
	go server.sender()

	server.receiver()
}

func NewServer(addrStr string) (*Server, error) {
	// Create UDP addr
	addr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		return nil, err
	}

	// Create UDP conn
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	// Generate public and private key
	privateKey, publicKey, err := crypto.GenKeyPair()
	if err != nil {
		return nil, err
	}

	server := &Server{
		conn:            conn,
		publicKey:       publicKey,
		privateKey:      privateKey,
		conns:           make(shared.Conns),
		sendChan:        make(chan *shared.UDPPayload, 100),
		messageCallback: func(conns shared.Conns, conn shared.Conn, message *shared.Message) {},
		exit:            make(chan bool),
		wg:              &sync.WaitGroup{},
	}

	udpPeers := make(shared.Peers)
	server.OnMessage(createMessageCallback(server, udpPeers))

	return server, nil
}
