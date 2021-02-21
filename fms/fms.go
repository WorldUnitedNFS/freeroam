// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fms

import (
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/WorldUnitedNFS/freeroam"
	"github.com/gorilla/websocket"
)

type PlayerInfo struct {
	Name string `json:"name"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

func NewMapServer(i *freeroam.Server) *MapServer {
	return &MapServer{
		i:        i,
		conns:    make(map[string]*websocket.Conn, 0),
		upgrader: websocket.Upgrader{},
		players:  make([]PlayerInfo, 0),
	}
}

type MapServer struct {
	sync.Mutex
	i        *freeroam.Server
	conns    map[string]*websocket.Conn
	upgrader websocket.Upgrader
	players  []PlayerInfo
}

func (s *MapServer) Handle(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print(err.Error())
		return
	}
	s.conns[c.RemoteAddr().String()] = c
}

func (s *MapServer) Run() {
	for {
		s.SendPlayers()
		time.Sleep(250 * time.Millisecond)
	}
}

func (s *MapServer) SendPlayers() {
	s.Lock()
	defer s.Unlock()
	if len(s.conns) == 0 {
		return
	}

	s.i.Lock()
	players := s.players[:0]
	for _, c := range s.i.Clients {
		if c.IsReady() {
			pos := c.GetPos()
			players = append(players, PlayerInfo{
				Name: c.PersonaName,
				X:    int(math.Round(pos.X)),
				Y:    int(math.Round(pos.Y)),
			})
		}
	}
	s.i.Unlock()

	for addr, conn := range s.conns {
		err := conn.WriteJSON(players)
		if err != nil {
			delete(s.conns, addr)
		}
	}
}
