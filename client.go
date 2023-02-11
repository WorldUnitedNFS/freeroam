// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package freeroam

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/WorldUnitedNFS/freeroam/math"
	"net"
	"runtime/debug"
	"sort"
	"sync"
	"time"
)

type clientPosSortInfo struct {
	Client *Client
	Length int
}

type clientPosSort []clientPosSortInfo

func (self clientPosSort) Len() int {
	return len(self)
}

func (self clientPosSort) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self clientPosSort) Less(i, j int) bool {
	return self[i].Length < self[j].Length
}

type ClientConfig struct {
	InitialTick     uint16
	Addr            *net.UDPAddr
	Conn            *net.UDPConn
	Buffers         *sync.Pool
	Clients         map[string]*Client
	AllowedPersonas []int
}

func newClient(opts ClientConfig) *Client {
	c := &Client{
		Addr:            opts.Addr,
		conn:            opts.Conn,
		startTime:       time.Now(),
		initialTick:     opts.InitialTick,
		tickDiff:        int16(opts.InitialTick - getServerTick()),
		seq:             0,
		slots:           make([]*slotInfo, 14),
		LastPacket:      time.Now(),
		clients:         opts.Clients,
		allowedPersonas: opts.AllowedPersonas,
		buffers:         opts.Buffers,
		updateID:        1,
	}
	return c
}

type Client struct {
	Addr            *net.UDPAddr
	conn            *net.UDPConn
	startTime       time.Time
	initialTick     uint16
	tickDiff        int16
	seq             uint16
	carPos          CarPosPacket
	chanInfo        []byte
	playerInfo      []byte
	slots           []*slotInfo
	LastPacket      time.Time
	PersonaName     string
	allowedPersonas []int
	ackMissedCount  int
	updateID        uint8
	buffers         *sync.Pool
	clients         map[string]*Client
	posRecvTD       uint16
}

func (c *Client) registerUpdate() {
	c.updateID++
	if c.updateID == 0 {
		c.updateID = 1
	}
}

func getServerTick() uint16 {
	return uint16(time.Now().UnixMilli())
}

func (c Client) getTimeDiff() uint16 {
	return uint16(time.Now().Sub(c.startTime).Seconds() * 1000)
}

func (c *Client) getSeq() uint16 {
	out := c.seq
	c.seq++
	return out
}

func (c *Client) replyHandshake() {
	buf := c.buffers.Get().(*bytes.Buffer)
	buf.Reset()
	binary.Write(buf, binary.BigEndian, c.getSeq())
	buf.WriteByte(0x01)
	binary.Write(buf, binary.BigEndian, getServerTick())
	binary.Write(buf, binary.BigEndian, c.tickDiff)
	buf.Write([]byte{0x49, 0x26, 0x03, 0x01})
	c.SendRawPacket(buf.Bytes())
	c.buffers.Put(buf)
}

// Active returns true if the client has communicated with the server lately.
func (c Client) Active() bool {
	return time.Now().Sub(c.LastPacket).Seconds() < 5
}

func (c *Client) processPacket(packet []byte) {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Printf("Error occured while processing packet (%v):\n", len(packet))
			fmt.Println(r)
			debug.PrintStack()
		}
	}()
	c.LastPacket = time.Now()
	srvCounter := binary.BigEndian.Uint16(packet[8:10])
	for _, slot := range c.slots {
		if slot != nil && !slot.UpdateACKed {
			if srvCounter == slot.PacketSentSeq {
				slot.UpdateACKed = true
			} else {
				slot.ACKMissedCount++
			}
		}
	}
	var updated bool
	data := packet[16 : len(packet)-5]
	reader := bytes.NewReader(data)
	for {
		ptype, err := reader.ReadByte()
		if err != nil {
			break
		}
		plen, _ := reader.ReadByte()
		innerData := make([]byte, plen)
		reader.Read(innerData)
		switch ptype {
		case 0x00:
			c.chanInfo = innerData
			updated = true
		case 0x01:
			if c.allowedPersonas != nil {
				personaID := binary.LittleEndian.Uint32(innerData[41:45])
				var allowed bool
				for _, id := range c.allowedPersonas {
					if id == int(personaID) {
						allowed = true
						break
					}
				}
				if !allowed {
					fmt.Printf("Kicking %v; %v != %v\n", c.Addr.String(), personaID, c.allowedPersonas)
					delete(c.clients, c.Addr.String())
					return
				}
			}
			c.playerInfo = innerData
			nameField := innerData[1:33]
			c.PersonaName = string(nameField[:cStrLen(nameField)])
			updated = true
		case 0x12:
			if c.IsReady() && !bytes.Equal(innerData[2:], c.carPos.packet[2:]) {
				updated = true
			}
			c.carPos.Update(innerData)
			c.posRecvTD = c.getTimeDiff()
		}
	}
	if c.IsReady() {
		if updated {
			c.registerUpdate()
		}
		c.sendPlayerSlots()
	}
}

func (c *Client) getClosestPlayers(clients []*Client) []*Client {
	closePlayers := make([]clientPosSortInfo, 0)
	for _, client := range clients {
		if !client.IsReady() || client.Addr == c.Addr {
			continue
		}
		distance := math.Distance(c.GetPos(), client.GetPos())
		closePlayers = append(closePlayers, clientPosSortInfo{
			Length: int(distance),
			Client: client,
		})
	}
	sort.Sort(clientPosSort(closePlayers))
	out := make([]*Client, min(14, len(closePlayers)))
	for i := range out {
		out[i] = closePlayers[i].Client
	}
	return out
}

func (c *Client) removeSlot(client *Client) {
	index := func() int {
		for i, slot := range c.slots {
			if slot != nil && slot.Client == client {
				return i
			}
		}
		return -1
	}()
	c.slots[index] = nil
}

func (c *Client) addSlot(client *Client) {
	index := -1
	for i, slot := range c.slots {
		if slot == nil {
			index = i
			break
		}
	}
	if index == -1 {
		panic("addSlot: tried to add client with all slots full")
	}
	c.slots[index] = &slotInfo{
		Client: client,
	}
}

func (c *Client) recalculateSlots(clients []*Client) {
	players := c.getClosestPlayers(clients)
	oldPlayers := make([]*Client, 0)
	for _, slot := range c.slots {
		if slot != nil {
			oldPlayers = append(oldPlayers, slot.Client)
		}
	}
	diff := ArrayDiff(oldPlayers, players)

	// Adding and removing slots at the same time causes some weird things to happen.
	// As a temporary workaround, only one section of the diff is handled at a time.
	// This allows the game to remove old players BEFORE swapping in new ones.
	if len(diff.Removed) > 0 {
		for _, removedClient := range diff.Removed {
			c.removeSlot(removedClient)
		}
	} else if len(diff.Added) > 0 {
		for _, addedClient := range diff.Added {
			c.addSlot(addedClient)
		}
	}
}

func (c *Client) sendPlayerSlots() {
	clients := make([]*Client, len(c.clients))
	{
		i := 0
		for _, cl := range c.clients {
			clients[i] = cl
			i++
		}
	}
	c.recalculateSlots(clients)
	buf := c.buffers.Get().(*bytes.Buffer)
	buf.Reset()
	seq := c.getSeq()
	binary.Write(buf, binary.BigEndian, seq)
	buf.WriteByte(0x02)
	binary.Write(buf, binary.BigEndian, getServerTick())
	binary.Write(buf, binary.BigEndian, c.tickDiff)
	binary.Write(buf, binary.BigEndian, seq)
	buf.Write([]byte{0xff, 0xff, 0x00})
	fullsSent := 0
	for _, slot := range c.slots {
		if slot == nil {
			buf.Write([]byte{0xff, 0xff})
		} else {
			//pktTime := uint16(int(c.getTimeDiff()) - slot.Client.Ping)
			if slot.HasSentFull && slot.Client.posRecvTD == slot.LastCPTime {
				buf.Write([]byte{0x00, 0xff})
			} else if fullsSent >= 3 {
				slot.Client.writeFullPosPacket(buf)
				slot.LastCPTime = slot.Client.posRecvTD
			} else if !slot.HasSentFull {
				slot.Client.writeFullSlotPacket(buf)
				slot.HasSentFull = true
				slot.PacketSentSeq = seq
				slot.LastCPTime = slot.Client.posRecvTD
				fullsSent++
			} else if slot.UpdateACKed || slot.ACKMissedCount < 5 {
				slot.Client.writeFullPosPacket(buf)
				slot.LastCPTime = slot.Client.posRecvTD
			} else {
				slot.Client.writeFullSlotPacket(buf)
				slot.ACKMissedCount = 0
				slot.PacketSentSeq = seq
				slot.LastCPTime = slot.Client.posRecvTD
				fullsSent++
			}
		}
	}
	buf.Write([]byte{0x01, 0x01, 0x01, 0x01})
	c.SendRawPacket(buf.Bytes())
	c.buffers.Put(buf)
}

// GetPos returns the current position of the client.
func (c Client) GetPos() math.Vector2D {
	return c.carPos.Pos()
}

// GetRotation returns the current rotation of the client.
func (c Client) GetRotation() float64 {
	return c.carPos.Rotation()
}

// SendRawPacket sends a raw UDP packet to the client.
func (c *Client) SendRawPacket(b []byte) error {
	_, err := c.conn.WriteToUDP(b, c.Addr)
	return err
}

// IsReady returns true if the client is ready to be broadcasted to other clients.
// This means that the server has valid channel info, player info and position data of the client.
func (c Client) IsReady() bool {
	return c.chanInfo != nil && c.playerInfo != nil && c.carPos.Valid()
}

func (c Client) writeFullPosPacket(buf *bytes.Buffer) {
	buf.WriteByte(0x00) // Slot start
	WriteSubpacket(buf, 0x12, c.carPos.Packet())
	buf.WriteByte(0xff) // Slot end
}

func (c Client) writeFullSlotPacket(buf *bytes.Buffer) {
	buf.WriteByte(0x00) // Slot start
	WriteSubpacket(buf, 0x00, c.chanInfo)
	WriteSubpacket(buf, 0x01, c.playerInfo)
	WriteSubpacket(buf, 0x12, c.carPos.Packet())
	buf.WriteByte(0xff) // Slot end
}
