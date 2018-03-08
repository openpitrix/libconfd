// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
	"sync"
	"sync/atomic"
	"time"
)

type Processor struct {
	config Config
	client Client
	option *options

	wg     sync.WaitGroup
	stoped int32
	runing int32
}

func NewProcessor(cfg Config, client Client) *Processor {
	logger.Debugln(getFuncName())
	logger.Debugf("%#v\n", cfg)

	return &Processor{
		config: cfg.Clone(),
		client: client,
	}
}

func (p *Processor) IsRunning() bool {
	return atomic.LoadInt32(&p.runing) == 1
}

func (p *Processor) isStoped() bool {
	return atomic.LoadInt32(&p.stoped) == 1
}

func (p *Processor) Run(opts ...Options) {
	logger.Debugln(getFuncName())

	if !atomic.CompareAndSwapInt32(&p.runing, 0, 1) {
		logger.Warning("Processor is running")
		return
	}

	p.option = newOptions(opts...)

	if p.option.useOnetimeMode {
		logger.Debugln("use onetime mode")

		p.wg.Add(1)
		go p.runOnce(opts...)
		return
	}

	if p.option.useIntervalMode {
		logger.Debugln("use interval mode")

		p.wg.Add(1)
		go p.runInIntervalMode(opts...)
		return
	}

	if p.option.useWatchMode {
		logger.Debugln("use watch mode")

		p.wg.Add(1)
		go p.runInWatchMode(opts...)
		return
	}

	if p.client.WatchEnabled() {
		logger.Debugln("default watch mode")

		p.wg.Add(1)
		go p.runInWatchMode(opts...)
		return
	}

	logger.Debugln("default interval mode")

	p.wg.Add(1)
	go p.runInIntervalMode(opts...)
	return
}

func (p *Processor) Stop() {
	logger.Debugln(getFuncName())

	if !p.IsRunning() {
		return
	}

	atomic.StoreInt32(&p.stoped, 1)
	p.wg.Wait()

	atomic.StoreInt32(&p.runing, 0)
	atomic.StoreInt32(&p.stoped, 0)
}

func (p *Processor) runOnce(opts ...Options) error {
	logger.Debugln(getFuncName())

	defer p.wg.Done()

	ts, err := MakeAllTemplateResourceProcessor(p.config, p.client)
	if err != nil {
		return err
	}

	var allErrors []error
	for _, t := range ts {
		if p.isStoped() {
			break
		}

		if err := t.Process(opts...); err != nil {
			allErrors = append(allErrors, err)
			logger.Error(err)
		}
	}
	if len(allErrors) > 0 {
		return allErrors[0]
	}

	return nil
}

func (p *Processor) runInIntervalMode(opts ...Options) {
	logger.Debugln(getFuncName())

	defer p.wg.Done()

	for {
		if p.isStoped() {
			return
		}

		ts, err := MakeAllTemplateResourceProcessor(p.config, p.client)
		if err != nil {
			logger.Warning(err)
			return
		}

		for _, t := range ts {
			if p.isStoped() {
				return
			}

			if err := t.Process(opts...); err != nil {
				logger.Error(err)
			}
		}

		time.Sleep(p.option.GetInterval())
	}
}

func (p *Processor) runInWatchMode(opts ...Options) {
	logger.Debugln(getFuncName())

	defer p.wg.Done()

	ts, err := MakeAllTemplateResourceProcessor(p.config, p.client)
	if err != nil {
		logger.Warning(err)
		return
	}

	var wg sync.WaitGroup
	var stopChan = make(chan bool)

	for i := 0; i < len(ts); i++ {
		wg.Add(1)
		go p.monitorPrefix(ts[i], &wg, stopChan, opts...)
	}

	for {
		if p.isStoped() {
			stopChan <- true
			break
		}

		time.Sleep(time.Second / 2)
	}

	wg.Wait()
	return
}

func (p *Processor) monitorPrefix(
	t *TemplateResourceProcessor,
	wg *sync.WaitGroup, stopChan chan bool,
	opts ...Options,
) {
	logger.Debugln(getFuncName())

	defer wg.Done()

	keys := t.getAbsKeys()
	for {
		if p.isStoped() {
			break
		}

		index, err := t.client.WatchPrefix(t.Prefix, keys, t.lastIndex, stopChan)
		if err != nil {
			logger.Error(err)
		}

		t.lastIndex = index
		if err := t.Process(opts...); err != nil {
			logger.Error(err)
		}
	}
}
