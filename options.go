// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
	"text/template"
	"time"
)

type options struct {
	useOnetimeMode  bool
	useIntervalMode bool
	defaultInterval time.Duration
	funcMap         template.FuncMap
	funcMapUpdater  []func(m template.FuncMap)

	hookBeforeCheckCmd  func(trName, cmd string, err error)
	hookAfterCheckCmd   func(trName, cmd string, err error)
	hookBeforeReloadCmd func(trName, cmd string, err error)
	hookAfterReloadCmd  func(trName, cmd string, err error)
}

type Options func(*options)

func newOptions(opts ...Options) *options {
	p := new(options)
	p.defaultInterval = time.Second * 600
	p.ApplyOptions(opts...)
	return p
}

func (opt *options) ApplyOptions(opts ...Options) {
	for _, fn := range opts {
		fn(opt)
	}
}
func (opt *options) GetInterval() time.Duration {
	if opt.defaultInterval > 0 {
		return opt.defaultInterval
	}
	return time.Second * 600
}

func WithOnetimeMode() Options {
	return func(opt *options) {
		opt.useOnetimeMode = true
	}
}

func WithIntervalMode() Options {
	return func(opt *options) {
		opt.useIntervalMode = true
	}
}

func WithInterval(interval time.Duration) Options {
	return func(opt *options) {
		opt.defaultInterval = interval
	}
}

func WithFuncMap(maps ...template.FuncMap) Options {
	return func(opt *options) {
		if opt.funcMap == nil {
			opt.funcMap = make(template.FuncMap)
		}
		for _, m := range maps {
			for k, fn := range m {
				opt.funcMap[k] = fn
			}
		}
	}
}

func WithFuncMapUpdater(funcMapUpdater ...func(m template.FuncMap)) Options {
	return func(opt *options) {
		opt.funcMapUpdater = append(opt.funcMapUpdater, funcMapUpdater...)
	}
}

func WithHookBeforeCheckCmd(fn func(trName, cmd string, err error)) Options {
	return func(opt *options) {
		opt.hookBeforeCheckCmd = fn
	}
}

func WithHookAfterCheckCmd(fn func(trName, cmd string, err error)) Options {
	return func(opt *options) {
		opt.hookAfterCheckCmd = fn
	}
}

func WithHookBeforeReloadCmd(fn func(trName, cmd string, err error)) Options {
	return func(opt *options) {
		opt.hookBeforeReloadCmd = fn
	}
}

func WithHookAfterReloadCmd(fn func(trName, cmd string, err error)) Options {
	return func(opt *options) {
		opt.hookAfterReloadCmd = fn
	}
}
