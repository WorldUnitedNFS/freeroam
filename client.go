// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package freeroam

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
	CliTime         []byte
	Addr            *net.UDPAddr
	Conn            *net.UDPConn
	Buffers         *sync.Pool
	Clients         map[string]*Client
	AllowedPersonas []int
}

func newClient(opts ClientConfig) *Client {
	slots := make(map[int]*slotInfo)
	for i := 0; i < 14; i++ {
		slots[i] = nil
	}
	c := &Client{
		Addr:            opts.Addr,
		conn:            opts.Conn,
		startTime:       time.Now(),
		cliTime:         binary.BigEndian.Uint16(opts.CliTime),
		seq:             0,
		Slots:           slots,
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
	cliTime         uint16
	seq             uint16
	carPos          CarPosPacket
	chanInfo        []byte
	playerInfo      []byte
	Slots           map[int]*slotInfo
	LastPacket      time.Time
	Ping            int
	PersonaName     string
	allowedPersonas []int
	ackMissedCount  int
	updateID        uint8
	buffers         *sync.Pool
	clients         map[string]*Client
	timeDiffDiff    int
	hasCalcDD       bool
	posRecvTD       uint16
}

func (c *Client) registerUpdate() {
	c.updateID++
	if c.updateID == 0 {
		c.updateID = 1
	}
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
	binary.Write(buf, binary.BigEndian, c.getTimeDiff())
	binary.Write(buf, binary.BigEndian, c.cliTime)
	buf.Write([]byte{0x49, 0x26, 0x03, 0x01})
	c.SendRawPacket(buf.Bytes())
	c.buffers.Put(buf)
}

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
	for i := 0; i < 14; i++ {
		slot := c.Slots[i]
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
			if c.IsOk() && !bytes.Equal(innerData[2:], c.carPos.packet[2:]) {
				updated = true
			}
			c.carPos.Update(innerData)
			iTime := binary.BigEndian.Uint16(innerData[0:2])
			c.posRecvTD = c.getTimeDiff()
			c.Ping = int(c.getTimeDiff() - iTime)
		}
	}
	if c.IsOk() {
		if updated {
			c.registerUpdate()
		}
		c.sendPlayerSlots()
	}
}

func (self *Client) getClosestPlayers(clients []*Client) []*Client {
	closePlayers := make([]clientPosSortInfo, 0)
	for _, client := range clients {
		if !client.IsOk() || client.Addr == self.Addr {
			continue
		}
		distance := Distance(self.GetPos(), client.GetPos())
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

func (self *Client) removeSlot(client *Client) {
	index := func() int {
		for i, c := range self.Slots {
			if c != nil && c.Client == client {
				return i
			}
		}
		return -1
	}()
	self.Slots[index] = nil
}

func (self *Client) addSlot(client *Client) {
	index := func() int {
		suitableSlots := make([]int, 0)
		for i, c := range self.Slots {
			if c == nil {
				suitableSlots = append(suitableSlots, i)
			}
		}
		sort.Ints(suitableSlots)
		return suitableSlots[0]
	}()
	self.Slots[index] = &slotInfo{
		Client: client,
	}
}

func (self *Client) recalculateSlots(clients []*Client) {
	players := self.getClosestPlayers(clients)
	oldPlayers := make([]*Client, 0)
	for _, v := range self.Slots {
		if v != nil {
			oldPlayers = append(oldPlayers, v.Client)
		}
	}
	diff := ArrayDiff(oldPlayers, players)
	for _, c := range diff.Removed {
		self.removeSlot(c)
	}
	for _, c := range diff.Added {
		self.addSlot(c)
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
	binary.Write(buf, binary.BigEndian, c.getTimeDiff())
	binary.Write(buf, binary.BigEndian, c.cliTime)
	binary.Write(buf, binary.BigEndian, seq)
	buf.Write([]byte{0xff, 0xff, 0x00})
	fullsSent := 0
	for i := 0; i < 14; i++ {
		slot := c.Slots[i]
		if slot == nil {
			buf.Write([]byte{0xff, 0xff})
		} else {
			time := uint16(int(c.getTimeDiff()) - (int(slot.Client.getTimeDiff()) - int(slot.Client.posRecvTD)))
			if slot.HasSentFull && slot.Client.posRecvTD == slot.LastCPTime {
				buf.Write([]byte{0x00, 0xff})
			} else if fullsSent >= 3 {
				slot.Client.writeFullPosPacket(buf, time)
				slot.LastCPTime = slot.Client.posRecvTD
			} else if !slot.HasSentFull {
				slot.Client.writeFullSlotPacket(buf, time)
				slot.HasSentFull = true
				slot.PacketSentSeq = seq
				slot.LastCPTime = slot.Client.posRecvTD
				fullsSent++
			} else if slot.UpdateACKed || slot.ACKMissedCount < 5 {
				slot.Client.writeFullPosPacket(buf, time)
				slot.LastCPTime = slot.Client.posRecvTD
			} else {
				slot.Client.writeFullSlotPacket(buf, time)
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

func (c Client) GetPos() Vector {
	return c.carPos.Pos()
}

func (c *Client) SendRawPacket(b []byte) error {
	_, err := c.conn.WriteToUDP(b, c.Addr)
	return err
}

func (c Client) IsOk() bool {
	return c.chanInfo != nil && c.playerInfo != nil && c.carPos.Valid()
}

func (c Client) writeFullPosPacket(buf *bytes.Buffer, time uint16) {
	buf.WriteByte(0x00)                       // Slot start
	buf.WriteByte(0x12)                       // Type
	buf.WriteByte(byte(len(c.carPos.packet))) // Size
	buf.Write(c.carPos.Packet(time))
	buf.WriteByte(0xff) // Slot end
}

func (c Client) writeFullSlotPacket(buf *bytes.Buffer, time uint16) {
	buf.WriteByte(0x00) // Slot start
	buf.WriteByte(0x00) // Type
	buf.WriteByte(0x22) // Size
	buf.Write(c.chanInfo)
	buf.WriteByte(0x01) // Type
	buf.WriteByte(0x41) // Size
	buf.Write(c.playerInfo)
	buf.WriteByte(0x12)                       // Type
	buf.WriteByte(byte(len(c.carPos.packet))) // Size
	buf.Write(c.carPos.Packet(time))
	buf.WriteByte(0xff) // Slot end
}
