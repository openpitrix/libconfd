// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

package libconfd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

var _ Client = (*JsonBackend)(nil)

type JsonBackend struct {
	JSONFile string
}

func NewJsonBackendClient(JSONFile string) *JsonBackend {
	return &JsonBackend{
		JSONFile: JSONFile,
	}
}

func (_ *JsonBackend) WatchEnabled() bool {
	return false
}

func (_ *JsonBackend) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	return 0, fmt.Errorf("do not support watch")
}

func (p *JsonBackend) GetValues(keys []string) (map[string]string, error) {
	data, err := ioutil.ReadFile(p.JSONFile)
	if err != nil {
		return nil, err
	}

	var dataMap map[string]interface{}
	err = json.Unmarshal(data, &dataMap)
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for k, v := range dataMap {
		if s, ok := v.(string); ok && strings.HasPrefix(k, "/") {
			m[k] = s
		}
	}
	return m, nil
}
