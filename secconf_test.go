// Copyright secconf. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE-secconf file.

package libconfd

import (
	"bytes"
	"testing"
)

var tSecconf_encodingTests = []struct {
	in, out string
}{
	{"secret", "secret"},
}

func TestSecconfEncoding(t *testing.T) {
	for _, tt := range tSecconf_encodingTests {
		encoded, err := secconfEncode([]byte(tt.in), bytes.NewBufferString(tSecconf_pubring))
		if err != nil {
			t.Errorf("%v", err)
		}
		decoded, err := secconfDecode(encoded, bytes.NewBufferString(tSecconf_secring))
		if err != nil {
			t.Errorf("%v", err)
		}
		if tt.out != string(decoded) {
			t.Errorf("want %s, got %s", tt.out, decoded)
		}
	}
}
