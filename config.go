// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

package libconfd

import (
	"bytes"
	"encoding/gob"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	// The path to confd configs. ("/etc/confd")
	ConfDir string `toml:"confdir"`

	// The backend polling interval in seconds. (10)
	Interval int `toml:"interval"`

	// Enable noop mode. Process all template resources; skip target update.
	Noop bool `toml:"noop"`

	// The string to prefix to keys. ("/")
	Prefix string `toml:"prefix"`

	// sync without check_cmd and reload_cmd.
	SyncOnly bool `toml:"sync-only"`

	// level which confd should log messages
	// DEBUG/INFO/WARN/ERROR/PANIC
	LogLevel string `toml:"log-level"`

	// enable watch support
	Watch bool `toml:"watch"`

	// the TOML backend file to watch for changes
	File string `toml:"file"`

	// keep staged files
	KeepStageFile bool `toml:"keep-stage-file"`

	// PGP secret keyring (for use with crypt functions)
	PGPPrivateKey string `toml:"pgp-private-key"`
}

const defaultConfigContent = `
# The path to confd configs. ("/etc/confd")
confdir = "./confd"

# The backend polling interval in seconds. (10)
interval = 10

# Enable noop mode. Process all template resources; skip target update.
noop = false

# The string to prefix to keys. ("/")
prefix = "/"

# sync without check_cmd and reload_cmd.
sync-only = true

# level which confd should log messages ("DEBUG")
log-level = "DEBUG"

# enable watch support
watch = false

# the TOML backend file to watch for changes
file = "./confd/backend-file.toml"

# keep staged files
keep-stage-file = false

# PGP secret keyring (for use with crypt functions)
pgp-private-key = ""
`

func NewDefaultConfig() (p *Config) {
	p = new(Config)
	_, err := toml.Decode(defaultConfigContent, p)
	if err != nil {
		panic(err)
	}
	return
}

func MustLoadConfig(name string) *Config {
	p, err := LoadConfig(name)
	if err != nil {
		logger.Fatal(err)
	}
	return p
}

func LoadConfig(name string) (p *Config, err error) {
	p = new(Config)
	_, err = toml.DecodeFile(name, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Config) Save(name string) error {
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(p); err != nil {
		return err
	}

	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(buf.String()); err != nil {
		return err
	}
	return nil
}

func (p *Config) Clone() *Config {
	var (
		q   = new(Config)
		buf bytes.Buffer
	)

	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)

	if err := enc.Encode(p); err != nil {
		logger.Fatal(err)
	}
	if err := dec.Decode(q); err != nil {
		logger.Fatal(err)
	}

	return q
}

func (p *Config) GetConfigDir() string {
	return filepath.Join(p.ConfDir, "conf.d")
}

func (p *Config) GetTemplateDir() string {
	return filepath.Join(p.ConfDir, "templates")
}

func (p *Config) makeTemplateDir() {
	os.MkdirAll(p.GetTemplateDir(), 0744)
}
