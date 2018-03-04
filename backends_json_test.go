// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

package libconfd

import (
	"testing"
)

func TestJsonBackend(t *testing.T) {
	c := NewJsonBackendClient("./testdata/simple.json")
	m, err := c.GetValues([]string{""})
	if err != nil {
		t.Fatal(err)
	}

	if v := m["/key"]; v != "foobar" {
		t.Fatal(v)
	}
}
