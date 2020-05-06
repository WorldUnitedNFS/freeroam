// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/google/gops/agent"
	"gitlab.com/sparkserver/freeroam"
	"gitlab.com/sparkserver/freeroam/fms"
)

var (
	listenAddr string
	mapAddr    string
)

func main() {
	flag.StringVar(&listenAddr, "listen", ":9999", "Address to listen to")
	flag.StringVar(&mapAddr, "map", "", "Address for map server to listen to")
	flag.Parse()

	if err := agent.Listen(agent.Options{ShutdownCleanup: true}); err != nil {
		log.Print(err)
	}

	i := freeroam.NewServer()

	if mapAddr != "" {
		mapSrv := fms.NewMapServer(i)

		fmsMux := http.NewServeMux()
		fmsMux.HandleFunc("/ws", mapSrv.Handle)

		go mapSrv.Run()
		err := http.ListenAndServe(mapAddr, fmsMux)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Starting server on %v", listenAddr)
	if err := i.Listen(listenAddr); err != nil {
		log.Fatal(err)
	}
}
