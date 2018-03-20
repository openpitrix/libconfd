// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package libconfd

type Application struct {
	cfg *Config
}

func NewApplication(cfg *Config, client Client) *Application {
	return &Application{
		cfg: cfg.Clone(),
	}
}

func (p *Application) Main() {
	panic("TODO")
}

func (p *Application) List() []string {
	panic("TODO")
}

func (p *Application) Info(name string) *TemplateResource {
	panic("TODO")
}

func (p *Application) Make(name string) error {
	panic("TODO")
}

func (p *Application) GetValue(key string) string {
	panic("TODO")
}

func (p *Application) RunOnce(opts ...Options) error {
	panic("TODO")
}

func (p *Application) Run(opts ...Options) error {
	panic("TODO")
}
