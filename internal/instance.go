// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package internal

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"
)

func NewInstance() *Instance {
	return &Instance{
		Clients: make(map[string]*Client),
		udpbuf:  make([]byte, 1024),
		buffers: &sync.Pool{
			New: func() interface{} { return new(bytes.Buffer) },
		},
	}
}

type Instance struct {
	sync.Mutex
	listener *net.UDPConn
	Clients  map[string]*Client
	udpbuf   []byte
	buffers  *sync.Pool
}

func (i *Instance) Listen(addrStr string) error {
	addr, err := net.ResolveUDPAddr("udp", addrStr)
	if err != nil {
		return err
	}
	listener, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	i.listener = listener
	go i.RunTimer()
	i.RunPacketRead()
	return nil
}

func (i *Instance) RunPacketRead() {
	for {
		addr, data := i.readPacket()
		i.Lock()
		if len(data) == 58 && data[2] == 0x06 {
			fmt.Printf("New client from %v\n", addr.String())
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

func (i *Instance) RunTimer() {
	timer := time.Tick(1 * time.Second)
	for {
		i.Lock()
		for k, client := range i.Clients {
			if !client.Active() {
				fmt.Printf("Removing inactive client %v\n", client.Addr.String())
				delete(i.Clients, k)
			}
		}
		i.Unlock()
		<-timer
	}
}

func (i *Instance) readPacket() (*net.UDPAddr, []byte) {
	len, addr, _ := i.listener.ReadFromUDP(i.udpbuf)
	return addr, i.udpbuf[:len]
}
