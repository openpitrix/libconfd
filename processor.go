// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
	"context"
	"sync"
	"time"
)

type RunOptions struct {
	OnCheckCmdDone  func()
	OnReloadCmdDone func()
}

type Processor struct {
	config   Config
	stopChan chan bool
	doneChan chan bool
	errChan  chan error
	interval int
	wg       sync.WaitGroup
}

func NewProcessor(cfg Config) *Processor {
	return &Processor{
		config: cfg,
	}
}

func (p *Processor) RunOnce(ctx context.Context, client Client, opt *RunOptions) error {
	ts, err := MakeAllTemplateResourceProcessor(p.config, client)
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

func (p *Processor) RunInIntervalMode(ctx context.Context, client Client, interval int, opt *RunOptions) error {
	defer close(p.doneChan)
	for {
		ts, err := MakeAllTemplateResourceProcessor(p.config, client)
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
		case <-time.After(time.Duration(p.interval) * time.Second):
			continue
		}
	}
}

func (p *Processor) RunInWatchMode(ctx context.Context, client Client, opt *RunOptions) error {
	defer close(p.doneChan)
	ts, err := MakeAllTemplateResourceProcessor(p.config, client)
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

func (p *Processor) monitorPrefix(t *TemplateResourceProcessor) {
	defer p.wg.Done()
	keys := utilAppendPrefix(t.Prefix, t.Keys)
	for {
		index, err := t.storeClient.WatchPrefix(t.Prefix, keys, t.lastIndex, p.stopChan)
		if err != nil {
			p.errChan <- err
			// Prevent backend errors from consuming all resources.
			time.Sleep(time.Second * 2)
			continue
		}
		t.lastIndex = index
		if err := t.Process(); err != nil {
			p.errChan <- err
		}
	}
}
