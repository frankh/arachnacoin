package node

import (
	"bufio"
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

type MessageChainRequest struct {
	Type   string `json:"type"`
	Latest string `json:"latest"`
}

type Peer struct {
	Address string
	Conn    net.Conn
}

var broadcastAddress, _ = net.ResolveUDPAddr("udp", "224.0.0.1:31042")
var broadcastPacket = []byte("Arachnacoin")
var broadcastInterval = 3 * time.Second
var connectedPeers = make(map[string]Peer)
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
	ln, err := net.Listen("tcp", "0.0.0.0:31042")
	checkErr(err)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		remoteIp := AddrToIp(conn.RemoteAddr())

		if AddrToIp(conn.RemoteAddr()) == AddrToIp(conn.LocalAddr()) {
			log.Printf("Ignored connecting to self")
			continue
		}

		log.Printf("Accepted connection from peer %s", remoteIp)
		peer := Peer{remoteIp, conn}
		connectedPeers[remoteIp] = peer
		go handlePeerConnection(peer)
	}
}

func handlePeerConnection(peer Peer) {
	var message Message

	for {
		jsonMessage, err := bufio.NewReader(peer.Conn).ReadBytes('\n')
		if err != nil {
			log.Printf("Disconnecting from peer: %s", err)
			delete(connectedPeers, peer.Address)
			peer.Conn.Close()
			return
		}

		err = json.Unmarshal(jsonMessage, &message)
		if err != nil {
			log.Printf("%s", jsonMessage)
			continue
		}

		switch message.Type {
		case "block":
			// log.Printf("Received block from %s", peer.Address)
			var messageBlock MessageBlock
			err = json.Unmarshal(jsonMessage, &messageBlock)
			if err != nil {
				log.Printf("Bad block...ignoring")
				continue
			}
			receiveBlock(peer, messageBlock.Block)
		case "chain":
			log.Printf("Received chain request from %s", peer.Address)
			var messageChain MessageChainRequest
			err = json.Unmarshal(jsonMessage, &messageChain)
			if err != nil {
				log.Printf("Bad chain request...ignoring")
				continue
			}
			handleChainRequest(peer, messageChain.Latest)
		default:
			log.Printf("Ignoring unknown message from peer %s", peer.Address)
			continue
		}

	}

}

func receiveBlock(peer Peer, b block.Block) {
	if store.FetchBlock(b.HashString()) != nil {
		// log.Printf("Already have this block")
		return
	}

	if store.ValidateBlock(b) {
		oldHeight := store.FetchHighestBlock().Height
		log.Printf("Saved block of height %d", b.Height)
		store.StoreBlock(b)
		if b.Height > oldHeight {
			BroadcastLatestBlock()
		}
	} else {
		requestBlockChain(peer, b)
	}
}

func requestBlockChain(peer Peer, b block.Block) {
	message := MessageChainRequest{
		"chain",
		b.HashString(),
	}

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("Couldn't read message %x", jsonMessage)
		return
	}

	log.Printf("Requesting chain up to %d from %s", b.Height, peer.Address)
	peer.Conn.Write(jsonMessage)
	peer.Conn.Write([]byte{'\n'})
}

func handleChainRequest(peer Peer, latest string) {
	b := store.FetchBlock(latest)
	if b == nil {
		log.Printf("Cannot handle chain request, block not found")
		return
	}

	hashChain := store.GetBlockHashChain(b)
	if hashChain == nil {
		log.Printf("Cannot handle chain request, broken chain")
		return
	}

	for i, _ := range hashChain {
		// Send last block in hash chain (earliest) first
		b = store.FetchBlock(hashChain[len(hashChain)-i-1])
		if b != nil {
			SendBlockToPeer(*b, peer)
		}
	}

}

func BroadcastLatestBlock() {
	b := store.FetchHighestBlock()
	// Don't broadcast genesis...
	if b.Height == 0 {
		return
	}

	for _, peer := range connectedPeers {
		SendBlockToPeer(b, peer)
	}
}

func SendBlockToPeer(b block.Block, peer Peer) {
	message := MessageBlock{
		"block",
		b,
	}

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Printf("Couldn't read message %x", jsonMessage)
		return
	}

	// log.Printf("Sending block %d to %s", b.Height, peer.Address)
	peer.Conn.Write(jsonMessage)
	peer.Conn.Write([]byte{'\n'})
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
	peer := Peer{peerIp, conn}
	connectedPeers[peerIp] = peer
	log.Printf("Connected to peer %s", peerIp)
	BroadcastLatestBlock()
	handlePeerConnection(peer)
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
		if peerIp == localIp || connectedPeers[peerIp].Conn != nil {
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
