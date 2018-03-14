// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"flag"

	"github.com/chai2010/libconfd"
)

var (
	cfgfile = flag.String("config-file", "./confd.toml", "config file")
)

func main() {
	flag.Parse()

	cfg := libconfd.MustLoadConfig(*cfgfile)
	client := libconfd.NewFileBackendsClient(cfg.File)

	libconfd.GetLogger().SetLevel(cfg.LogLevel)
	libconfd.NewProcessor().Run(cfg, client)
}
