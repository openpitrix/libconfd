// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
	"errors"
	"sync"
	"time"
)

var ErrShutdown = errors.New("libconfd: processor is shut down")

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

func NewProcessor() *Processor {
	logger.Debugln(getFuncName())

	p := &Processor{
		closeChan: make(chan bool),
	}

	p.wg.Add(1)
	go p.run()
	return p
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
		call.Error = ErrShutdown
		call.done()
	}

	p.pending = p.pending[:0]
}

func (p *Processor) Close() error {
	close(p.closeChan)
	p.wg.Wait()
	return nil
}

func (p *Processor) Go(cfg Config, client Client, opts ...Options) *Call {
	call := new(Call)

	call.Config = cfg.Clone()
	call.Client = client
	call.Opts = append([]Options{}, opts...)
	call.Done = make(chan *Call, 10) // buffered.

	p.send(call)
	return call
}

func (p *Processor) Run(cfg Config, client Client, opts ...Options) error {
	call := <-p.Go(cfg, client, opts...).Done
	return call.Error
}

func (p *Processor) send(call *Call) {
	if p.isClosing() {
		call.Error = ErrShutdown
		call.done()
		return
	}

	p.pushPendingCall(call)
}

func (p *Processor) run() {
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
}
