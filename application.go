// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package libconfd

import (
	"fmt"
)

type Application struct {
	cfg    *Config
	client Client
}

func NewApplication(cfg *Config, client Client) *Application {
	return &Application{
		cfg:    cfg.Clone(),
		client: client,
	}
}

func (p *Application) List() []string {
	panic("TODO")
}

func (p *Application) Info(name string) {
	panic("TODO")
}

func (p *Application) Make(name string) {
	panic("TODO")
}

func (p *Application) GetValues(keys ...string) {
	m, err := p.client.GetValues(keys)
	if err != nil {
		logger.Fatal(err)
	}

	var maxLen = 1
	for i := range keys {
		if len(keys[i]) > maxLen {
			maxLen = len(keys[i])
		}
	}

	for _, k := range keys {
		fmt.Printf("%-*s => %s\n", maxLen, k, m[k])
	}
}

func (p *Application) RunOnce(opts ...Options) {
	panic("TODO")
}

func (p *Application) Run(opts ...Options) {
	panic("TODO")
}
