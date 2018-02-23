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

func StartConfd(storeClient StoreClient, onetime, watch bool, interval int) {
	logger.Info("Starting confd")

	var templateConfig Config
	{
		// make config
	}

	if onetime {
		if err := Process(templateConfig, storeClient); err != nil {
			logger.Fatal(err.Error())
		}
		os.Exit(0)
	}

	stopChan := make(chan bool)
	doneChan := make(chan bool)
	errChan := make(chan error, 10)

	var processor Processor
	switch {
	case watch:
		processor = WatchProcessor(templateConfig, stopChan, doneChan, errChan)
	default:
		processor = IntervalProcessor(templateConfig, stopChan, doneChan, errChan, interval)
	}

	go processor.Process(storeClient)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case err := <-errChan:
			logger.Error(err.Error())
		case s := <-signalChan:
			logger.Info(fmt.Sprintf("Captured %v. Exiting...", s))
			close(doneChan)
		case <-doneChan:
			os.Exit(0)
		}
	}
}

type StoreClient interface {
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error)
}

type Confd struct {
	//
}

func NewConfd(cfg *Config) *Confd {
	return &Confd{}
}

func (p *Confd) IsRunning() bool {
	return false
}

func (p *Confd) Start() error {
	return nil
}

func (p *Confd) Stop() error {
	return nil
}
