// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

// Package libconfd provides mini confd lib.
package libconfd

import (
	"errors"
	"sync"
	"time"
)

type Call struct {
	Config *Config
	Client Client
	Error  error
	Done   chan *Call
}

func (call *Call) done() {
	select {
	case call.Done <- call:
		// ok
	default:
		// We don't want to block here. It is the caller's responsibility to make
		// sure the channel has enough buffer space. See comment in Go().
		logger.Debugln("rpc: discarding Call reply due to insufficient Done chan capacity")
	}
}

type Processor struct {
	pendingMutex sync.Mutex
	pending      []*Call

	closeChan chan bool
	wg        sync.WaitGroup
}

func (p *Processor) isClosing() bool {
	if p.closeChan == nil {
		logger.Panic("closeChan is nil")
	}
	select {
	case <-p.closeChan:
		return true
	default:
		return false
	}
}

func (p *Processor) addPendingCall(call *Call) {
	p.pendingMutex.Lock()
	defer p.pendingMutex.Unlock()

	p.pending = append(p.pending, call)
}
func (p *Processor) getPendingCall() *Call {
	p.pendingMutex.Lock()
	defer p.pendingMutex.Unlock()

	if len(p.pending) == 0 {
		return nil
	}

	call := p.pending[0]
	p.pending = p.pending[1:]
	return call
}
func (p *Processor) clearPendingCall() {
	p.pendingMutex.Lock()
	defer p.pendingMutex.Unlock()

	for _, call := range p.pending {
		call.Error = errors.New("libconfd: processor is shut down")
		call.done()
	}

	p.pending = p.pending[:0]
}

func NewProcessor() *Processor {
	p := &Processor{
		closeChan: make(chan bool),
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

		for {
			if p.isClosing() {
				p.clearPendingCall()
				return
			}

			call := p.getPendingCall()
			if call == nil {
				time.Sleep(time.Second / 10)
				continue
			}

			p.wg.Add(1)
			go func() {
				defer p.wg.Done()
				defer call.done()

				logger.Debugln("process start")
				defer logger.Debugln("process done")

				p.process(call)
			}()
		}
	}()

	return p
}

func (p *Processor) Go(cfg *Config, client Client, opts ...Options) *Call {
	if client == nil {
		logger.Panic("client is nil")
	}

	call := new(Call)

	call.Config = cfg.Clone().applyOptions(opts...)
	call.Client = client
	call.Done = make(chan *Call, 10) // buffered.

	if err := cfg.Valid(); err != nil {
		call.Error = err
		call.done()
		return call
	}

	logger.SetLevel(cfg.LogLevel)

	p.addPendingCall(call)
	return call
}

func (p *Processor) Run(cfg *Config, client Client, opts ...Options) error {
	if err := cfg.Valid(); err != nil {
		return err
	}
	if client == nil {
		logger.Panic("client is nil")
	}

	logger.SetLevel(cfg.LogLevel)

	call := <-p.Go(cfg, client, opts...).Done
	return call.Error
}

func (p *Processor) Close() error {
	close(p.closeChan)
	p.wg.Wait()
	return nil
}

func (p *Processor) process(call *Call) {
	switch {
	case call.Config.Onetime:
		p.runOnce(call)
	case call.Config.Watch:
		p.runInIntervalMode(call)
	default:
		p.runInWatchMode(call)
	}
}

func (p *Processor) runOnce(call *Call) {
	ts, err := MakeAllTemplateResourceProcessor(call.Config, call.Client)
	if err != nil {
		logger.Error(err)
		call.Error = err
		return
	}

	for _, t := range ts {
		if p.isClosing() {
			return
		}

		if err := t.Process(call); err != nil {
			logger.Error(err)
		}
	}

	return
}

func (p *Processor) runInIntervalMode(call *Call) {

	for {
		if p.isClosing() {
			return
		}

		ts, err := MakeAllTemplateResourceProcessor(call.Config, call.Client)
		if err != nil {
			logger.Warning(err)
			call.Error = err
			return
		}

		for _, t := range ts {
			if p.isClosing() {
				return
			}

			if err := t.Process(call); err != nil {
				logger.Error(err)
				continue
			}
		}

		time.Sleep(time.Duration(call.Config.Interval) * time.Second)
	}
}

func (p *Processor) runInWatchMode(call *Call) {
	ts, err := MakeAllTemplateResourceProcessor(call.Config, call.Client)
	if err != nil {
		logger.Warning(err)
		return
	}

	var wg sync.WaitGroup
	var stopChan = make(chan bool)

	for i := 0; i < len(ts); i++ {
		wg.Add(1)
		go func(t *TemplateResourceProcessor) {
			defer wg.Done()
			p.monitorPrefix(t, &wg, stopChan, call)
		}(ts[i])
	}

	for {
		time.Sleep(time.Second / 2)

		if p.isClosing() {
			close(stopChan)
			break
		}
	}

	wg.Wait()
	return
}

func (p *Processor) monitorPrefix(
	t *TemplateResourceProcessor,
	wg *sync.WaitGroup, stopChan chan bool,
	call *Call,
) {
	keys := t.getAbsKeys()

	for {
		if p.isClosing() {
			return
		}

		index, err := t.client.WatchPrefix(t.Prefix, keys, t.lastIndex, stopChan)
		if err != nil {
			logger.Error(err)
		}

		t.lastIndex = index
		if err := t.Process(call); err != nil {
			logger.Error(err)
		}
	}
}
