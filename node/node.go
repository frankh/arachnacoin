package node

import (
	"bytes"
	"log"
	"net"
	"strings"
	"time"
)

var broadcastAddress, _ = net.ResolveUDPAddr("udp", "224.0.0.1:31042")
var broadcastPacket = []byte("Arachnacoin")
var broadcastInterval = 3 * time.Second
var connectedPeers = make(map[string]*net.Conn)
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
		connectedPeers[remoteIp] = &conn
	}

}

func ConnectToPeer(peerIp string) {
	conn, err := net.Dial("tcp", peerIp+":31042")
	if err != nil {
		log.Printf("Failed to connect to peer %s", peerIp)
		return
	}
	if AddrToIp(conn.RemoteAddr()) == AddrToIp(conn.LocalAddr()) {
		log.Printf("Ignoring connection from localhost")
		localIp = AddrToIp(conn.LocalAddr())
		return
	}
	connectedPeers[peerIp] = &conn
	log.Printf("Connected to peer %s", peerIp)
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
			log.Printf("Connecting to peer %s", peerIp)
			ConnectToPeer(peerIp)
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
