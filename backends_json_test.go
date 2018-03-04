// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

package libconfd

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestJsonBackend(t *testing.T) {
	os.MkdirAll("./_temp", 0766)

	jsonFile := "./_temp/backend-json-file.json"
	err := ioutil.WriteFile(jsonFile, []byte(tJsonFileContent), 0666)
	if err != nil {
		t.Fatal(err)
	}

	c := NewJsonBackendClient(jsonFile)
	m, err := c.GetValues([]string{""})
	if err != nil {
		t.Fatal(err)
	}

	if v := m["/key"]; v != "foobar" {
		t.Fatal(v)
	}
}

var tJsonFileContent = `
{
    "/key": "foobar",
    "/database/host": "127.0.0.1",
    "/database/password": "p@sSw0rd",
    "/database/port": "3306",
    "/database/username": "libconfd",
    "/upstream/app1": "10.0.1.10:8080",
    "/upstream/app2": "10.0.1.11:8080",
    "/prefix/database/host": "127.0.0.1",
    "/prefix/database/password": "p@sSw0rd",
    "/prefix/database/port": "3306",
    "/prefix/database/username": "libconfd",
    "/prefix/upstream/app1": "10.0.1.10:8080",
    "/prefix/upstream/app2": "10.0.1.11:8080",
    "-": ""
}
`
