// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
	"sync"
	"time"
)

type Processor interface {
	Process(client StoreClient) error
}

func NewOnetimeProcessor(cfg Config) Processor {
	return &onetimeProcessor{
		config: cfg,
	}
}

type onetimeProcessor struct {
	config Config
}

func (p *onetimeProcessor) Process(client StoreClient) error {
	ts, err := MakeTemplateResourceList(p.config, client)
	if err != nil {
		return err
	}

	var allErrors []error
	for _, t := range ts {
		if err := t.process(); err != nil {
			allErrors = append(allErrors, err)
			logger.Error(err)
		}
	}
	if len(allErrors) > 0 {
		return allErrors[0]
	}

	return nil
}

type intervalProcessor struct {
	config   Config
	stopChan chan bool
	doneChan chan bool
	errChan  chan error
	interval int
}

func NewIntervalProcessor(config Config, stopChan, doneChan chan bool, errChan chan error, interval int) Processor {
	return &intervalProcessor{config, stopChan, doneChan, errChan, interval}
}

func (p *intervalProcessor) Process(client StoreClient) error {
	defer close(p.doneChan)
	for {
		ts, err := MakeTemplateResourceList(p.config, client)
		if err != nil {
			logger.Warning(err)
			return err
		}

		for _, t := range ts {
			if err := t.process(); err != nil {
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

type watchProcessor struct {
	config   Config
	stopChan chan bool
	doneChan chan bool
	errChan  chan error
	wg       sync.WaitGroup
}

func NewWatchProcessor(config Config, stopChan, doneChan chan bool, errChan chan error) Processor {
	return &watchProcessor{
		config:   config,
		stopChan: stopChan,
		doneChan: doneChan,
		errChan:  errChan,
	}
}

func (p *watchProcessor) Process(client StoreClient) error {
	defer close(p.doneChan)
	ts, err := MakeTemplateResourceList(p.config, client)
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

func (p *watchProcessor) monitorPrefix(t *TemplateResource) {
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
		if err := t.process(); err != nil {
			p.errChan <- err
		}
	}
}
