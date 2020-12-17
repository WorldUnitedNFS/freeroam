// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package freeroam

import (
	"bytes"
	"log"
	"net"
	"sync"
	"time"
)

func NewServer() *Server {
	return &Server{
		Clients: make(map[string]*Client),
		recvbuf: make([]byte, 1024),
		buffers: &sync.Pool{
			New: func() interface{} { return new(bytes.Buffer) },
		},
	}
}

type Server struct {
	sync.Mutex
	listener *net.UDPConn
	Clients  map[string]*Client
	recvbuf  []byte
	buffers  *sync.Pool
}

func (i *Server) Listen(addrStr string) error {
	addr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		return err
	}
	i.listener, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	go i.RunTimer()
	i.RunPacketRead()
	return nil
}

func (i *Server) RunPacketRead() {
	for {
		addr, data := i.readPacket()
		i.Lock()
		if len(data) == 58 && data[2] == 0x06 {
			log.Printf("New client from %v", addr.String())
			client := newClient(ClientConfig{
				CliTime: data[52:54],
				Addr:    addr,
				Conn:    i.listener,
				Buffers: i.buffers,
				Clients: i.Clients,
			})
			i.Clients[addr.String()] = client
			client.replyHandshake()
			i.Unlock()
			continue
		}
		client, ok := i.Clients[addr.String()]
		if ok {
			client.processPacket(data)
		}
		i.Unlock()
	}
}

func (i *Server) RunTimer() {
	for {
		i.Lock()
		for k, client := range i.Clients {
			if !client.Active() {
				log.Printf("Removing inactive client %v", client.Addr.String())
				delete(i.Clients, k)
			}
		}
		i.Unlock()
		time.Sleep(1 * time.Second)
	}
}

func (i *Server) readPacket() (*net.UDPAddr, []byte) {
	recvlen, addr, _ := i.listener.ReadFromUDP(i.recvbuf)
	return addr, i.recvbuf[:recvlen]
}
