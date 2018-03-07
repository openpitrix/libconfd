// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

package libconfd

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestNewDefaultConfig(t *testing.T) {
	_ = NewDefaultConfig()
}

func TestConfig(t *testing.T) {
	cfgfile := "./_confd.toml"

	defer os.Remove(cfgfile)
	ioutil.WriteFile(cfgfile, []byte(defaultConfigContent), 0666)

	p, err := LoadConfig(cfgfile)
	if err != nil {
		t.Fatal(err)
	}

	tConfig := NewDefaultConfig()
	if !reflect.DeepEqual(p, tConfig) {
		t.Fatalf("expect = %#v, got = %#v", p, tConfig)
	}
}
