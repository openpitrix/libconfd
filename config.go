// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

package libconfd

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	// The path to confd configs.
	// If the confdir is rel path, must convert to abs path.
	//
	// abspath = filepath.Join(ConfigPath, Config.ConfDir)
	//
	ConfDir string `toml:"confdir"`

	// Ignored template name list
	IgnoredList []string `toml:"ignored"`

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
# The path to confd configs.
# If the confdir is rel path, must convert to abs path.
#
# abspath = filepath.Join(ConfigPath, Config.ConfDir)
#
confdir = "confd"

# Ignored template name list
ignored = ["ignored.tmpl"]

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

func newDefaultConfig() (p *Config) {
	p = new(Config)
	_, err := toml.Decode(defaultConfigContent, p)
	if err != nil {
		logger.Panic(err)
	}
	if !filepath.IsAbs(p.ConfDir) {
		absdir, err := filepath.Abs(".")
		if err != nil {
			logger.Panic(err)
		}
		p.ConfDir = filepath.Clean(filepath.Join(absdir, p.ConfDir))
	}
	if p.File != "" && !filepath.IsAbs(p.File) {
		absdir, err := filepath.Abs(".")
		if err != nil {
			logger.Panic(err)
		}
		p.File = filepath.Clean(filepath.Join(absdir, p.File))
	}
	return
}

func MustLoadConfig(path string) *Config {
	p, err := LoadConfig(path)
	if err != nil {
		logger.Fatal(err)
	}
	return p
}

func LoadConfig(path string) (p *Config, err error) {
	p = new(Config)
	_, err = toml.DecodeFile(path, p)
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(p.ConfDir) {
		absdir, err := filepath.Abs(filepath.Dir(path))
		logger.Debugln(absdir)
		if err != nil {
			return nil, err
		}
		p.ConfDir = filepath.Clean(filepath.Join(absdir, p.ConfDir))
	}
	if p.File != "" && !filepath.IsAbs(p.File) {
		absdir, err := filepath.Abs(filepath.Dir(path))
		logger.Debugln(absdir)
		if err != nil {
			return nil, err
		}
		p.File = filepath.Clean(filepath.Join(absdir, p.File))
	}
	return p, nil
}

func (p *Config) Valid() error {
	if !filepath.IsAbs(p.ConfDir) {
		return fmt.Errorf("ConfDir is not abs path: %s", p.ConfDir)
	}
	if p.File != "" && !filepath.IsAbs(p.File) {
		return fmt.Errorf("BackendFile is not abs path: %s", p.File)
	}

	if !dirExists(p.ConfDir) {
		return fmt.Errorf("ConfDir not exists: %s", p.ConfDir)
	}
	if p.File != "" && !fileExists(p.File) {
		return fmt.Errorf("BackendFile not exists: %s", p.File)
	}

	if p.Interval < 0 {
		return fmt.Errorf("invalid Interval: %d", p.Interval)
	}
	if !newLogLevel(p.LogLevel).Valid() {
		return fmt.Errorf("invalid LogLevel: %s", p.LogLevel)
	}

	return nil
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

func (p *Config) GetDefaultTemplateOutputDir() string {
	return filepath.Join(p.ConfDir, "templates_output")
}
