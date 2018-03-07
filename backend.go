// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

package libconfd

import (
	"strings"
)

type Client interface {
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error)
	WatchEnabled() bool
}

// NewFileBackendsClient create toml or json backend file client
func NewFileBackendsClient(file string) Client {
	if strings.HasSuffix(file, ".json") {
		return NewJsonBackendClient(file)
	}

	// toml
	return NewTomlBackendClient(file)
}
