// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package freeroam

import (
	"bytes"
	"encoding/binary"
	"github.com/WorldUnitedNFS/freeroam/grid"
	"log"
	"net"
	"sync"
	"time"
)

func NewServer() *Server {
	worldGrid := grid.CreateWorldGrid(2048, 1125, 3)
	gridCellOccupants := make([]map[*Client]bool, worldGrid.NumCells())

	for i := 0; i < worldGrid.NumCells(); i++ {
		gridCellOccupants[i] = make(map[*Client]bool)
	}

	return &Server{
		Clients: make(map[string]*Client),
		recvbuf: make([]byte, 1024),
		buffers: &sync.Pool{
			New: func() interface{} { return new(bytes.Buffer) },
		},
		worldGrid:         worldGrid,
		gridCellOccupants: gridCellOccupants,
	}
}

type Server struct {
	sync.Mutex
	listener          *net.UDPConn
	Clients           map[string]*Client
	recvbuf           []byte
	buffers           *sync.Pool
	worldGrid         grid.WorldGrid
	gridCellOccupants []map[*Client]bool
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
				InitialTick: binary.BigEndian.Uint16(data[52:54]),
				Addr:        addr,
				Conn:        i.listener,
				Buffers:     i.buffers,
				Clients:     i.Clients,
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

func mapGamePosToWorldMapPos(x, y float32) (mx, my float32) {
	mx = 0.183583939*x - 10.0328626
	my = -0.183613514*y + 773.060633
	return
}

func (i *Server) RunTimer() {
	for {
		i.Lock()
		for k, client := range i.Clients {
			currentCell := client.currentGridCell
			if !client.Active() {
				log.Printf("Removing inactive client %v", client.Addr.String())

				if currentCell != nil {
					delete(i.gridCellOccupants[currentCell.ID], client)
				}

				delete(i.Clients, k)
			} else if client.IsReady() {
				mappedX, mappedY := mapGamePosToWorldMapPos(client.posX, client.posY)
				newCell := i.worldGrid.FindCell(mappedX, mappedY)
				if newCell == nil {
					log.Printf("Could not find cell for client %v at (%f, %f) [WM: (%f, %f)]", client.Addr, client.posX, client.posY, mappedX, mappedY)
				} else if currentCell != newCell {
					if currentCell != nil {
						delete(i.gridCellOccupants[currentCell.ID], client)
					}

					log.Printf("Client %v is now in cell %d", client.Addr, newCell.ID)
					i.gridCellOccupants[newCell.ID][client] = true
					client.currentGridCell = newCell
					currentCell = newCell
				}

				if currentCell != nil {
					client.recalculateSlots(i.gridCellOccupants[currentCell.ID])
				}
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
