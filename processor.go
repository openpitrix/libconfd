// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
	"errors"
	"log"
	"sync"
)

var ErrShutdown = errors.New("libconfd: processor is shut down")

type Call struct {
	Config Config
	Client Client
	Opts   []Options
	Error  error
	Done   chan *Call
}

type Processor struct {
	reqMutex   sync.Mutex // protects following
	requestSeq uint64

	mutex    sync.Mutex // protects following
	seq      uint64
	pending  map[uint64]*Call
	closing  bool // user has called Close
	shutdown bool // server has told us to stop
}

func NewProcessor() *Processor {
	logger.Debugln(getFuncName())

	p := &Processor{}

	go p.input()
	return p
}

func (p *Processor) input() {
	//
}

// Close calls the underlying codec's Close method. If the connection is already
// shutting down, ErrShutdown is returned.
func (client *Processor) Close() error {
	client.mutex.Lock()
	if client.closing {
		client.mutex.Unlock()
		return ErrShutdown
	}
	client.closing = true
	client.mutex.Unlock()

	// close chan, send close single
	return nil
}

func (p *Processor) Go(cfg Config, client Client, done chan *Call, opts ...Options) *Call {
	call := new(Call)

	call.Config = cfg.Clone()
	call.Client = client
	call.Opts = append([]Options{}, opts...)

	if done == nil {
		done = make(chan *Call, 10) // buffered.
	} else {
		// If caller passes done != nil, it must arrange that
		// done has enough buffer for the number of simultaneous
		// RPCs that will be using that channel. If the channel
		// is totally unbuffered, it's best not to run at all.
		if cap(done) == 0 {
			log.Panic("rpc: done channel is unbuffered")
		}
	}
	call.Done = done
	p.send(call)
	return call
}

func (p *Processor) Run(cfg Config, client Client, opts ...Options) error {
	call := <-p.Go(cfg, client, make(chan *Call, 1), opts...).Done
	return call.Error
}

func (p *Processor) send(call *Call) {
	p.reqMutex.Lock()
	defer p.reqMutex.Unlock()

	// Register this call.
	p.mutex.Lock()
	if p.shutdown || p.closing {
		call.Error = ErrShutdown
		p.mutex.Unlock()
		call.done()
		return
	}

	seq := p.seq
	p.seq++
	p.pending[seq] = call
	p.mutex.Unlock()

	// Encode and send the request.
	p.requestSeq = seq

	var err error // = write request
	if err != nil {
		p.mutex.Lock()
		call = p.pending[seq]
		delete(p.pending, seq)
		p.mutex.Unlock()
		if call != nil {
			call.Error = err
			call.done()
		}
	}
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
