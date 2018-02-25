// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

package libconfd

import (
	"os"
	"os/signal"
	"syscall"
)

type ConfdOptions struct {
	Onetime  bool
	Watch    bool
	Interval int
}

type Confd struct {
	cfg    Config
	client BackendClient
	opt    ConfdOptions
}

func New(cfg Config, client BackendClient, opt ConfdOptions) *Confd {
	return &Confd{
		cfg:    cfg,
		client: client,
		opt:    opt,
	}
}

func NewOnetime(cfg Config, client BackendClient) *Confd {
	return &Confd{
		cfg:    cfg,
		client: client,
		opt:    ConfdOptions{Onetime: true},
	}
}

func NewWatch(cfg Config, client BackendClient) *Confd {
	return &Confd{
		cfg:    cfg,
		client: client,
		opt:    ConfdOptions{Watch: true},
	}
}

func NewInterval(cfg Config, client BackendClient, interval int) *Confd {
	return &Confd{
		cfg:    cfg,
		client: client,
		opt:    ConfdOptions{Interval: interval},
	}
}

func (p *Confd) Run() {
	logger.Info("Starting confd")

	if p.opt.Onetime {
		var processor = NewOnetimeProcessor(p.cfg)
		if err := processor.Process(p.client); err != nil {
			logger.Fatal(err)
		}
		os.Exit(0)
	}

	stopChan := make(chan bool)
	doneChan := make(chan bool)
	errChan := make(chan error, 10)

	var processor Processor
	switch {
	case p.opt.Watch:
		processor = NewWatchProcessor(p.cfg, stopChan, doneChan, errChan)
	default:
		processor = NewIntervalProcessor(p.cfg, stopChan, doneChan, errChan, p.opt.Interval)
	}

	go processor.Process(p.client)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case err := <-errChan:
			logger.Error(err)
		case s := <-signalChan:
			logger.Infof("Captured %v. Exiting...", s)
			close(doneChan)
		case <-doneChan:
			os.Exit(0)
		}
	}
}
