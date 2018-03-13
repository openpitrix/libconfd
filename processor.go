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
	Config Config
	Client Client
	Opts   []Options
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
		panic("closeChan is nil")
	}
	select {
	case <-p.closeChan:
		return true
	default:
		return false
	}
}

func (p *Processor) pushPendingCall(call *Call) {
	p.pendingMutex.Lock()
	defer p.pendingMutex.Unlock()

	p.pending = append(p.pending, call)
}
func (p *Processor) popPendingCall() *Call {
	p.pendingMutex.Lock()
	defer p.pendingMutex.Unlock()

	if len(p.pending) == 0 {
		return nil
	}

	call := p.pending[0]
	p.pending = p.pending[:len(p.pending)-1]
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

			call := p.popPendingCall()
			if call == nil {
				time.Sleep(time.Second / 10)
				continue
			}

			p.wg.Add(1)
			go func() {
				defer p.wg.Done()
				p.process(call)
			}()
		}
	}()

	return p
}

func (p *Processor) Go(cfg Config, client Client, opts ...Options) *Call {
	call := new(Call)

	call.Config = cfg.Clone()
	call.Client = client
	call.Opts = append([]Options{}, opts...)
	call.Done = make(chan *Call, 10) // buffered.

	p.pushPendingCall(call)
	return call
}

func (p *Processor) Run(cfg Config, client Client, opts ...Options) error {
	call := <-p.Go(cfg, client, opts...).Done
	return call.Error
}

func (p *Processor) Close() error {
	close(p.closeChan)
	p.wg.Wait()
	return nil
}

func (p *Processor) process(call *Call) {
	opt := newOptions(call.Opts...)

	switch {
	case opt.useOnetimeMode:
		p.runOnce(call)
	case opt.useIntervalMode:
		p.runInIntervalMode(call)
	case opt.useWatchMode:
		p.runInWatchMode(call)
	default:
		if call.Client.WatchEnabled() {
			p.runInWatchMode(call)
		} else {
			p.runInIntervalMode(call)
		}
	}
}

func (p *Processor) runOnce(call *Call) {
	opt := newOptions(call.Opts...)

	ts, err := MakeAllTemplateResourceProcessor(call.Config, call.Client)
	if err != nil {
		logger.Error(err)
		return
	}

	var allErrors []error
	for _, t := range ts {
		if p.isClosing() || opt.isClosing() {
			return
		}

		if err := t.Process(call.Opts...); err != nil {
			allErrors = append(allErrors, err)
			logger.Error(err)
		}
	}

	return
}

func (p *Processor) runInIntervalMode(call *Call) {
	opt := newOptions(call.Opts...)

	for {
		if p.isClosing() || opt.isClosing() {
			return
		}

		ts, err := MakeAllTemplateResourceProcessor(call.Config, call.Client)
		if err != nil {
			logger.Warning(err)
			continue
		}

		for _, t := range ts {
			if p.isClosing() {
				return
			}

			if err := t.Process(call.Opts...); err != nil {
				logger.Error(err)
			}
		}

		time.Sleep(opt.GetInterval())
	}
}

func (p *Processor) runInWatchMode(call *Call) {
	opt := newOptions(call.Opts...)

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

		if p.isClosing() || opt.isClosing() {
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
	opt := newOptions(call.Opts...)
	keys := t.getAbsKeys()

	for {
		if p.isClosing() || opt.isClosing() {
			return
		}

		index, err := t.client.WatchPrefix(t.Prefix, keys, t.lastIndex, stopChan)
		if err != nil {
			logger.Error(err)
		}

		t.lastIndex = index
		if err := t.Process(call.Opts...); err != nil {
			logger.Error(err)
		}
	}
}
