// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"encoding/json"
	"flag"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/debug"
	"time"

	"gitlab.com/sparkserver/freeroam"
	"gitlab.com/sparkserver/freeroam/fms"
)

var (
	listenAddr string
	debugAddr  string
	mapAddr    string
)

func main() {
	flag.StringVar(&listenAddr, "listen", ":9999", "Address to listen to")
	flag.StringVar(&listenAddr, "l", ":9999", "Address to listen to (shorthand)")
	flag.StringVar(&debugAddr, "debug", "localhost:6060", "Address for debug endpoint to listen to")
	flag.StringVar(&mapAddr, "map", "", "Address for map server to listen to")
	flag.Parse()

	i := freeroam.NewInstance()
	log.Printf("Starting server on %v", listenAddr)
	go i.Listen(listenAddr)

	debugMux := http.NewServeMux()
	debugMux.HandleFunc("/debug", func(rw http.ResponseWriter, req *http.Request) {
		i.Lock()
		defer i.Unlock()
		out := make([]interface{}, 0)
		for addr, client := range i.Clients {
			pos := client.GetPos()
			slots := make(map[int]*string)
			for i, slot := range client.Slots {
				if slot == nil || slot.Client == nil {
					slots[i] = nil
				} else {
					addr := slot.Client.Addr.String()
					slots[i] = &addr
				}
			}
			out = append(out, map[string]interface{}{
				"addr":     addr,
				"ping":     client.Ping,
				"idle_for": math.Round(time.Now().Sub(client.LastPacket).Seconds() * 1000),
				"pos":      []float64{pos.X, pos.Y},
				"slots":    slots,
			})
		}
		b, _ := json.Marshal(out)
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(200)
		rw.Write(b)
	})
	debugMux.HandleFunc("/debug/gc", func(rw http.ResponseWriter, req *http.Request) {
		var gcs debug.GCStats
		debug.ReadGCStats(&gcs)
		b, _ := json.Marshal(gcs)
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(200)
		rw.Write(b)
	})
	go http.ListenAndServe(debugAddr, debugMux)

	if mapAddr != "" {
		mapSrv := fms.NewMapServer(i)

		fmsMux := http.NewServeMux()
		fmsMux.HandleFunc("/ws", mapSrv.Handle)

		go mapSrv.Run()
		err := http.ListenAndServe(mapAddr, fmsMux)
		if err != nil {
			panic(err)
		}
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
}
