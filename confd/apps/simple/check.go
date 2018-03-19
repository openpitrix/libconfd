// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"reflect"
)

func checkConfigFile(path string) {
	panic("TODO")
}

func checkJsonResponse(cfg *Config) {
	var got Config

	addr := fmt.Sprintf("http://%s:%d", getLocalIP(), cfg.ListenPort)
	if err := getJsonByURL(addr, &got); err != nil {
		log.Fatalf("checkJsonResponse: getJsonByURL: %v\n", err)
	}

	if !reflect.DeepEqual(cfg, &got) {
		log.Printf("checkJsonResponse: expect = %v\n", cfg)
		log.Printf("checkJsonResponse: got = %v\n", &got)
		log.Fatal()
	}

	// OK
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

func getJsonByURL(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
