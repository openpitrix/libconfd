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
	flagWatch         = flag.Bool("watch", false, "use watch mode")
	flagOnetime       = flag.Bool("onetime", false, "run once and exit")
	flagInterval      = flag.Int("interval", 10, "backend polling interval seconds")
	flagConfdir       = flag.String("confdir", "", "confd conf directory")
	flagConfigFile    = flag.String("config-file", "./testdata/confd.toml", "the confd config file")
	flagJsonFile      = flag.String("file", "./testdata/simple.json", "the JSON file to watch for changes")
	flagKeepStageFile = flag.Bool("file", false, "keep staged files")
	flagLogLevel      = flag.String("log-level", "INFO", "log level. oneof: DEBUG/INFO/WARN/ERROR/PANIC")
	flagNoop          = flag.Bool("noop", false, "only show pending changes")
)

func main() {
	flag.Parse()

	cfg := libconfd.MustLoadConfig(*flagConfigFile)
	client := libconfd.NewJsonBackendClient(*flagJsonFile)

	libconfd.NewProcessor(cfg, client).Run()
}
