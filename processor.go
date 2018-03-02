// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
	"sync"
	"time"
)

type Processor struct {
	config Config
	client Client
	option *options

	stopChan chan bool
	doneChan chan bool
	errChan  chan error
	wg       sync.WaitGroup
}

func NewProcessor(cfg Config, client Client) *Processor {
	return &Processor{
		config: cfg.Clone(),
		client: client,
	}
}

func (p *Processor) IsRunning() bool {
	return false
}

func (p *Processor) Run(opts ...Options) error {
	p.option = newOptions(opts...)

	if p.option.useOnetimeMode {
		return p.runOnce(opts...)
	}

	if p.option.useIntervalMode || !p.client.WatchEnabled() {
		return p.runInIntervalMode(opts...)
	}

	return p.runInWatchMode(opts...)
}

func (p *Processor) Stop() error {
	return nil
}

func (p *Processor) runOnce(opts ...Options) error {
	ts, err := MakeAllTemplateResourceProcessor(p.config, p.client)
	if err != nil {
		return err
	}

	var allErrors []error
	for _, t := range ts {
		if err := t.Process(); err != nil {
			allErrors = append(allErrors, err)
			logger.Error(err)
		}
	}
	if len(allErrors) > 0 {
		return allErrors[0]
	}

	return nil
}

func (p *Processor) runInIntervalMode(opts ...Options) error {
	defer close(p.doneChan)
	for {
		ts, err := MakeAllTemplateResourceProcessor(p.config, p.client)
		if err != nil {
			logger.Warning(err)
			return err
		}

		for _, t := range ts {
			if err := t.Process(); err != nil {
				logger.Error(err)
			}
		}

		select {
		case <-p.stopChan:
			break
		case <-time.After(p.option.GetInterval()):
			continue
		}
	}
}

func (p *Processor) runInWatchMode(opts ...Options) error {
	defer close(p.doneChan)
	ts, err := MakeAllTemplateResourceProcessor(p.config, p.client)
	if err != nil {
		logger.Warning(err)
		return err
	}
	for _, t := range ts {
		t := t
		p.wg.Add(1)
		go p.monitorPrefix(t)
	}
	p.wg.Wait()
	return nil
}

func (p *Processor) monitorPrefix(t *TemplateResourceProcessor, opts ...Options) {
	defer p.wg.Done()
	keys := t.getAbsKeys()
	for {
		index, err := t.client.WatchPrefix(t.Prefix, keys, t.lastIndex, p.stopChan)
		if err != nil {
			p.errChan <- err
			// Prevent backend errors from consuming all resources.
			time.Sleep(time.Second * 2)
			continue
		}
		t.lastIndex = index
		if err := t.Process(opts...); err != nil {
			p.errChan <- err
		}
	}
}
