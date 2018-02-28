// Copyright confd. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-confd file.

package env

import (
	"os"
	"strings"

	"github.com/chai2010/libconfd"
)

var logger = libconfd.GetLogger()

// _EnvClient provides a shell for the env client
type _EnvClient struct{}

// NewEnvClient returns a new client
func NewEnvClient() libconfd.Client {
	return new(_EnvClient)
}

func (_ *_EnvClient) WatchEnabled() bool {
	return false
}

// GetValues queries the environment for keys
func (_ *_EnvClient) GetValues(keys []string) (map[string]string, error) {
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

	logger.Debugf("Key Map: %#v", vars)
	return vars, nil
}

func (_ *_EnvClient) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	<-stopChan
	return 0, nil
}
