// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

package libconfd

import (
	"reflect"
	"testing"
)

func TestNewDefaultConfig(t *testing.T) {
	_ = newDefaultConfig()
}

func TestConfig(t *testing.T) {
	tConfig := newDefaultConfig()

	p, err := LoadConfig("confd.toml")
	if err != nil {
		t.Fatal(err)
	}

	// ignord Ignored fileds
	tConfig.IgnoredList = nil
	p.IgnoredList = nil

	if !reflect.DeepEqual(p, tConfig) {
		t.Fatalf("expect = %#v, got = %#v", tConfig, p)
	}
}
