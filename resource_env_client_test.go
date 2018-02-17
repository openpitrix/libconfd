// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package libconfd

import (
	"fmt"
	"os"
	"strings"

	"github.com/golang/glog"
)

// Client provides a shell for the env client
type EnvClient struct{}

// NewEnvClient returns a new client
func NewEnvClient() (*EnvClient, error) {
	return &EnvClient{}, nil
}

// GetValues queries the environment for keys
func (c *EnvClient) GetValues(keys []string) (map[string]string, error) {
	allEnvVars := os.Environ()
	envMap := make(map[string]string)
	for _, e := range allEnvVars {
		index := strings.Index(e, "=")
		envMap[e[:index]] = e[index+1:]
	}

	vars := make(map[string]string)

	transform := func(key string) string {
		var replacer = strings.NewReplacer("/", "_")

		k := strings.TrimPrefix(key, "/")
		return strings.ToUpper(replacer.Replace(k))
	}
	clean := func(key string) string {
		var cleanReplacer = strings.NewReplacer("_", "/")

		newKey := "/" + key
		return cleanReplacer.Replace(strings.ToLower(newKey))
	}

	for _, key := range keys {
		k := transform(key)
		for envKey, envValue := range envMap {
			if strings.HasPrefix(envKey, k) {
				vars[clean(envKey)] = envValue
			}
		}
	}

	glog.V(1).Info(fmt.Sprintf("Key Map: %#v", vars))
	return vars, nil
}

func (c *EnvClient) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
