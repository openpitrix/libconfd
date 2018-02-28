// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

package libconfd

import (
	"bytes"
	"encoding/gob"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	ConfDir       string
	ConfigDir     string
	KeepStageFile bool
	Noop          bool
	Prefix        string
	SyncOnly      bool
	TemplateDir   string
	PGPPrivateKey []byte
}

func MustLoadConfig(name string) Config {
	p, err := LoadConfig(name)
	if err != nil {
		logger.Fatal(err)
	}
	return p
}

func LoadConfig(name string) (p Config, err error) {
	md, err := toml.DecodeFile(name, p)
	if err != nil {
		return Config{}, err
	}
	if unknownKeys := md.Undecoded(); len(unknownKeys) != 0 {
		logger.Warning("config: Undecoded keys:", unknownKeys)
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

func (p *Config) Clone() Config {
	var (
		q   Config
		buf bytes.Buffer
	)

	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)

	if err := enc.Encode(p); err != nil {
		logger.Fatal(err)
	}
	if err := dec.Decode(&q); err != nil {
		logger.Fatal(err)
	}

	return q
}
