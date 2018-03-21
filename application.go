// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package libconfd

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
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

func (p *Application) List(re string) {
	_, paths, err := ListTemplateResource(p.cfg.ConfDir)
	if err != nil {
		logger.Fatal(err)
	}
	for _, s := range paths {
		basename := filepath.Base(s)
		if re == "" {
			fmt.Println(basename)
			continue
		}
		matched, err := regexp.MatchString(re, basename)
		if err != nil {
			logger.Fatal(err)
		}
		if matched {
			fmt.Println(basename)
		}
	}
}

func (p *Application) Info(names ...string) {
	if len(names) == 0 {
		_, paths, err := ListTemplateResource(p.cfg.ConfDir)
		if err != nil {
			logger.Fatal(err)
		}
		names = paths
	}
	for _, name := range names {
		if !strings.HasSuffix(name, ".toml") {
			name += ".toml"
		}
		tc, err := LoadTemplateResourceFile(p.cfg.ConfDir, name)
		if err != nil {
			logger.Fatal(err)
		}
		fmt.Println(tc.TomlString())
	}
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
