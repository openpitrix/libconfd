// Copyright 2018 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a Apache-style
// license that can be found in the LICENSE file.

package libconfd

import (
	"errors"
	"strings"
)

var (
	ErrEmptySrc = errors.New("libconfd: empty src template")
	ErrNotExist = errors.New("libconfd: key does not exist")
	ErrNoMatch  = errors.New("libconfd: no keys match")
)

func notDeviceOrResourceBusyError(err error) bool {
	if err != nil && strings.Contains(err.Error(), "device or resource busy") {
		return false
	}
	return true // OK
}
