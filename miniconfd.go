// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"

	"openpitrix.io/libconfd"
)

var (
	cfgfile = flag.String("config-file", "./confd.toml", "config file")
)

func main() {
	flag.Parse()

	cfg := libconfd.MustLoadConfig(*cfgfile)
	client := libconfd.NewFileBackendsClient(cfg.File)

	libconfd.NewProcessor().Run(cfg, client)
}
