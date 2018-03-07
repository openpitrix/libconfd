// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

package libconfd

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

var _ Client = (*TomlBackend)(nil)

type TomlBackend struct {
	TOMLFile string
}

func NewTomlBackendClient(TOMLFile string) *TomlBackend {
	return &TomlBackend{
		TOMLFile: TOMLFile,
	}
}

func (_ *TomlBackend) WatchEnabled() bool {
	return false
}

func (_ *TomlBackend) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	return 0, fmt.Errorf("do not support watch")
}

func (p *TomlBackend) GetValues(keys []string) (m map[string]string, err error) {
	var dataMap map[string]string
	_, err = toml.DecodeFile(p.TOMLFile, &dataMap)
	if err != nil {
		return nil, err
	}

	// skip invalid key
	m = make(map[string]string)
	for k, v := range dataMap {
		if strings.HasPrefix(k, "/") {
			m[k] = v
		}
	}

	return m, nil
}
