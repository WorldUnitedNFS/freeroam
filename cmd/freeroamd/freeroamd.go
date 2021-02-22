// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/WorldUnitedNFS/freeroam"
	"github.com/WorldUnitedNFS/freeroam/fms"
	"github.com/google/gops/agent"
	"github.com/pelletier/go-toml"
)

func main() {
	var config freeroam.Config

	configFile, configErr := os.Open("config.toml")

	if os.IsNotExist(configErr) {
		config = freeroam.DefaultConfig()
		configFile, configErr = os.Create("config.toml")

		if configErr != nil {
			log.Fatal(configErr)
			return
		}

		configMarshalled, configErr := toml.Marshal(config)

		if configErr != nil {
			log.Fatal(configErr)
			return
		}

		_, configErr = configFile.Write(configMarshalled)
		if configErr != nil {
			log.Fatal(configErr)
			return
		}

		log.Print("Generated default config")
	} else {
		configBytes, configErr := ioutil.ReadFile("config.toml")
		if configErr != nil {
			log.Fatal(configErr)
			return
		}

		configErr = toml.Unmarshal(configBytes, &config)

		if configErr != nil {
			log.Fatal(configErr)
			return
		}
	}

	if err := agent.Listen(agent.Options{ShutdownCleanup: true}); err != nil {
		log.Print(err)
	}

	i := freeroam.NewServer()

	if config.FMS.ListenAddress != "" {
		mapSrv := fms.NewMapServer(i, config.FMS)

		fmsMux := http.NewServeMux()
		fmsMux.HandleFunc("/ws", mapSrv.Handle)

		go mapSrv.Run()
		go func() {
			log.Printf("Starting FMS on %v", config.FMS.ListenAddress)
			err := http.ListenAndServe(config.FMS.ListenAddress, fmsMux)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	log.Printf("Starting server on %v", config.UDP.ListenAddress)
	if err := i.Listen(config.UDP.ListenAddress); err != nil {
		log.Fatal(err)
	}
}
