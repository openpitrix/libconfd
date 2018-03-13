// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
	"sync"
	"time"
)

func (p *Processor) process(call *Call) {
	logger.Debugln(getFuncName())

	switch opt := newOptions(call.Opts...); true {
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
	logger.Debugln(getFuncName())

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
	logger.Debugln(getFuncName())

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
	logger.Debugln(getFuncName())

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
	logger.Debugln(getFuncName())

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
