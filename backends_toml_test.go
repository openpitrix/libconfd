// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package libconfd

import (
	"testing"
)

func TestTomlBackend(t *testing.T) {
	c := NewTomlBackendClient("./confd/backend-file.toml")
	m, err := c.GetValues([]string{""})
	if err != nil {
		t.Fatal(err)
	}

	if v := m["/key"]; v != "foobar" {
		t.Fatal(v)
	}
}
