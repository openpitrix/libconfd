// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package libconfd

import "github.com/BurntSushi/toml"

type BeckendConfig struct {
	Type string `toml:"type" json:"type"`
	Host string `toml:"host" json:"host"`

	User     string `toml:"user" json:"user"`
	Password string `toml:"password" json:"password"`

	ClientCAKeys string `toml:"client-ca-keys" json:"client-ca-keys"`
	ClientCert   string `toml:"client-cert" json:"client-cert"`
	ClientKey    string `toml:"client-key" json:"client-key"`
}

type BeckendClient interface {
	Type() string
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error)
	WatchEnabled() bool
}

// NewFileBackendsClient create toml backend file client
func NewFileBackendsClient(file string) BeckendClient {
	cfg := MustLoadBeckendConfig(file)
	logger.Assert(cfg.Type == (*TomlBackend)(nil).Type())

	return NewTomlBackendClient(cfg)
}

func MustLoadBeckendConfig(path string) *BeckendConfig {
	p, err := LoadBeckendConfig(path)
	if err != nil {
		logger.Fatal(err)
	}
	return p
}

func LoadBeckendConfig(path string) (p *BeckendConfig, err error) {
	p = new(BeckendConfig)
	_, err = toml.DecodeFile(path, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}
