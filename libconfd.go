// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

package libconfd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type Config struct {
	ConfDir       string
	ConfigDir     string
	KeepStageFile bool
	Noop          bool
	Prefix        string
	SyncOnly      bool
	TemplateDir   string
	PGPPrivateKey []byte
}

type StoreClient interface {
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error)
}

type Options struct {
	Onetime  bool
	Watch    bool
	Interval int
}

func ServeConfd(cfg Config, client StoreClient, opt Options) {
	logger.Info("Starting confd")

	if opt.Onetime {
		if err := Process(cfg, client); err != nil {
			logger.Fatal(err)
		}
		os.Exit(0)
	}

	stopChan := make(chan bool)
	doneChan := make(chan bool)
	errChan := make(chan error, 10)

	var processor Processor
	switch {
	case opt.Watch:
		processor = WatchProcessor(cfg, stopChan, doneChan, errChan)
	default:
		processor = IntervalProcessor(cfg, stopChan, doneChan, errChan, opt.Interval)
	}

	go processor.Process(client)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case err := <-errChan:
			logger.Error(err)
		case s := <-signalChan:
			logger.Info(fmt.Sprintf("Captured %v. Exiting...", s))
			close(doneChan)
		case <-doneChan:
			os.Exit(0)
		}
	}
}
