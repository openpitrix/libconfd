// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
	"text/template"
	"time"
)

type options struct {
	prv int
}

type Options func(*options)

func WithOnetimeMode() Options {
	return nil
}

func WithIntervalMode() Options {
	return nil
}

func WithInterval(interval time.Duration) Options {
	return nil
}

func WithHookBeforeCheckCmd(fn func(trName string, err error)) Options {
	return nil
}

func WithHookAfterCheckCmd(fn func(trName, cmd string, err error)) Options {
	return nil
}

func WithHookBeforeReloadCmd(fn func(trName string, err error)) Options {
	return nil
}

func WithHookAfterReloadCmd(fn func(trName, cmd string, err error)) Options {
	return nil
}

func WithFuncMap(m template.FuncMap, updateFuncMap ...func(m template.FuncMap)) Options {
	return nil
}
