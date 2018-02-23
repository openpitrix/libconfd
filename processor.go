// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
	"fmt"
	"sync"
	"time"
)

type Processor interface {
	Process()
}

func Process(config Config) error {
	ts, err := getTemplateResources(config)
	if err != nil {
		return err
	}
	return process(ts)
}

func process(ts []*TemplateResource) error {
	var lastErr error
	for _, t := range ts {
		if err := t.process(); err != nil {
			logger.Error(err.Error())
			lastErr = err
		}
	}
	return lastErr
}

type intervalProcessor struct {
	config   Config
	stopChan chan bool
	doneChan chan bool
	errChan  chan error
	interval int
}

func IntervalProcessor(config Config, stopChan, doneChan chan bool, errChan chan error, interval int) Processor {
	return &intervalProcessor{config, stopChan, doneChan, errChan, interval}
}

func (p *intervalProcessor) Process() {
	defer close(p.doneChan)
	for {
		ts, err := getTemplateResources(p.config)
		if err != nil {
			logger.Fatal(err.Error())
			break
		}
		process(ts)
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

func WatchProcessor(config Config, stopChan, doneChan chan bool, errChan chan error) Processor {
	return &watchProcessor{
		config:   config,
		stopChan: stopChan,
		doneChan: doneChan,
		errChan:  errChan,
	}
}

func (p *watchProcessor) Process() {
	defer close(p.doneChan)
	ts, err := getTemplateResources(p.config)
	if err != nil {
		logger.Fatal(err.Error())
		return
	}
	for _, t := range ts {
		t := t
		p.wg.Add(1)
		go p.monitorPrefix(t)
	}
	p.wg.Wait()
}

func (p *watchProcessor) monitorPrefix(t *TemplateResource) {
	defer p.wg.Done()
	keys := appendPrefix(t.Prefix, t.Keys)
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

func getTemplateResources(config Config) ([]*TemplateResource, error) {
	var lastError error
	templates := make([]*TemplateResource, 0)
	if logger.V(1) {
		logger.Info("Loading template resources from confdir " + config.ConfDir)
	}
	if !isFileExist(config.ConfDir) {
		logger.Warning(fmt.Sprintf("Cannot load template resources: confdir '%s' does not exist", config.ConfDir))
		return nil, nil
	}
	paths, err := recursiveFindFiles(config.ConfigDir, "*toml")
	if err != nil {
		return nil, err
	}

	if len(paths) < 1 {
		logger.Warning("Found no templates")
	}

	for _, p := range paths {
		if logger.V(1) {
			logger.Info(fmt.Sprintf("Found template: %s", p))
		}
		t, err := NewTemplateResource(p, config)
		if err != nil {
			lastError = err
			continue
		}
		templates = append(templates, t)
	}
	return templates, lastError
}
