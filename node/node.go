package node

import (
	"bytes"
	"encoding/json"
	"github.com/frankh/arachnacoin/block"
	"github.com/frankh/arachnacoin/store"
	"log"
	"net"
	"strings"
	"time"
)

type Message struct {
	Type string `json:"type"`
}

type MessageBlock struct {
	Type  string      `json:"type"`
	Block block.Block `json:"block"`
}

type MessageChain struct {
	Type   string        `json:"type"`
	Blocks []block.Block `json:"blocks"`
}

var broadcastAddress, _ = net.ResolveUDPAddr("udp", "224.0.0.1:31042")
var broadcastPacket = []byte("Arachnacoin")
var broadcastInterval = 3 * time.Second
var connectedPeers = make(map[string]net.Conn)
var localIp string

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func AddrToIp(addr net.Addr) string {
	return strings.Split(addr.String(), ":")[0]
}

func PeerServer() {
	log.Printf("Listening for peer connections on 31042")
	ln, err := net.Listen("tcp", ":31042")
	checkErr(err)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		remoteIp := AddrToIp(conn.RemoteAddr())

		if AddrToIp(conn.RemoteAddr()) == AddrToIp(conn.LocalAddr()) {
			return
		}

		log.Printf("Accepted connection from peer %s", remoteIp)
		connectedPeers[remoteIp] = conn
		go handlePeerConnection(remoteIp, conn)
	}
}

func handlePeerConnection(peer string, conn net.Conn) {
	var message Message
	rawMessage := make([]byte, 10000)

	for {
		n, err := conn.Read(rawMessage)
		if err != nil {
			log.Printf("Disconnecting from peer: %s", err)
			delete(connectedPeers, peer)
			conn.Close()
			return
		}
		jsonMessage := make([]byte, n)
		copy(jsonMessage, rawMessage)

		err = json.Unmarshal(jsonMessage, &message)
		if err != nil {
			panic(err)
			continue
		}

		switch message.Type {
		case "block":
			log.Printf("Received block from peer")
			var messageBlock MessageBlock
			err = json.Unmarshal(jsonMessage, &messageBlock)
			if err != nil {
				log.Printf("Bad block...ignoring")
				continue
			}
			receiveBlock(messageBlock.Block)
		default:
			continue
		}

	}

}

func receiveBlock(b block.Block) {
	if b.Height <= store.FetchHighestBlock().Height {
		return
	}

	if store.ValidateBlock(b) {
		log.Printf("Saved block of height %d", b.Height)
		store.StoreBlock(b)
		BroadcastLatestBlock()
	} else {
		log.Printf("Could not validate block...requesting chain")
	}
}

func BroadcastLatestBlock() {
	b := store.FetchHighestBlock()
	// Don't broadcast genesis...
	if b.Height == 0 {
		return
	}

	for peer, conn := range connectedPeers {
		SendBlockToPeer(b, peer, conn)
	}
}

func SendBlockToPeer(b block.Block, peer string, conn net.Conn) {
	message := MessageBlock{
		"block",
		b,
	}

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}

	log.Printf("Sending block %d to %s", b.Height, peer)
	conn.Write(jsonMessage)
}

func ConnectToPeer(peerIp string) {
	conn, err := net.Dial("tcp", peerIp+":31042")
	if err != nil {
		log.Printf("Failed to connect to peer %s: %s", peerIp, err)
		return
	}
	if AddrToIp(conn.RemoteAddr()) == AddrToIp(conn.LocalAddr()) {
		localIp = AddrToIp(conn.LocalAddr())
		return
	}
	connectedPeers[peerIp] = conn
	log.Printf("Connected to peer %s", peerIp)
	BroadcastLatestBlock()
	handlePeerConnection(peerIp, conn)
}

func ListenForPeers() {
	log.Printf("Listening for udp broadcasts on 31042")
	inConn, err := net.ListenMulticastUDP("udp", nil, broadcastAddress)
	checkErr(err)

	buf := make([]byte, len(broadcastPacket))
	for {
		_, addr, err := inConn.ReadFrom(buf)
		if err != nil {
			continue
		}

		peerIp := AddrToIp(addr)
		// Don't try and connect to already connected peers
		if peerIp == localIp || connectedPeers[peerIp] != nil {
			continue
		}

		if bytes.Compare(buf, broadcastPacket) == 0 {
			go ConnectToPeer(peerIp)
		}
	}
}

func BroadcastForPeers() {
	log.Printf("Sending udp broadcasts on 31042")
	outConn, err := net.DialUDP("udp", nil, broadcastAddress)
	checkErr(err)

	for {
		time.Sleep(broadcastInterval)
		outConn.Write([]byte(broadcastPacket))
	}
}
